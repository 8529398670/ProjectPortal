package models

import (
	errors "errors"
	net_url "net/url"
	slices "slices"
	sort "sort"
	strings "strings"
	time "time"

	db "projectportal/server/db"
	encryption "projectportal/server/encryption"
)

// Project is one app in the portal's catalog: another go-webserver or
// python-webserver deployment, its address, and the admin API key the portal
// uses there to create accounts and mint login links. Only admins add them.
//
// AdminKey is sealed on its own (PurposeRemoteKeys), like a TOTP secret,
// because server.encrypt_at_rest can be off and a key cannot be hashed: the
// portal has to send it. No route ever hands it back.
type Project struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	BaseURL     string `json:"base_url"`
	Icon        string `json:"icon,omitempty"`
	Color       string `json:"color,omitempty"`
	AdminKey    string `json:"admin_key,omitempty"`
	// Listed is whether people see it at all; ExcludedUserIDs are the people
	// it is kept from while it is listed. Everyone else may join it.
	Listed          bool     `json:"listed"`
	ExcludedUserIDs []uint64 `json:"excluded_user_ids,omitempty"`
	Position        int      `json:"position"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	// What the last check found (the project-health job, or Check now).
	Health    string     `json:"health,omitempty"`     // up, down
	KeyStatus string     `json:"key_status,omitempty"` // ok, invalid, not_admin, missing
	CheckNote string     `json:"check_note,omitempty"`
	CheckedAt *time.Time `json:"checked_at,omitempty"`
	SoftDelete
}

const (
	HealthUp   = "up"
	HealthDown = "down"

	KeyOK       = "ok"
	KeyInvalid  = "invalid"
	KeyNotAdmin = "not_admin"
	KeyMissing  = "missing"
)

var (
	ErrBadBaseURL    = errors.New( "the address must be http:// or https:// with a host, and nothing after the host but a path" )
	ErrBadProjectName = errors.New( "name must be 1-80 characters" )
)

// NormalizeBaseURL is the address the portal calls and sends people to:
// scheme and host, an optional path, never a query, a fragment or a password.
func NormalizeBaseURL( raw string ) ( result string , err error ) {
	parsed , parse_err := net_url.Parse( strings.TrimSpace( raw ) )
	if parse_err != nil || ( parsed.Scheme != "http" && parsed.Scheme != "https" ) || parsed.Host == "" ||
		parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		err = ErrBadBaseURL
		return
	}
	result = parsed.Scheme + "://" + strings.ToLower( parsed.Host ) + strings.TrimRight( parsed.Path , "/" )
	return
}

func validProjectName( name string ) ( result bool ) {
	trimmed := strings.TrimSpace( name )
	result = len( trimmed ) >= 1 && len( trimmed ) <= 80
	return
}

// SealRemoteKey and OpenRemoteKey seal an API key of another project for
// storing, and open it again to send.
func SealRemoteKey( secret_key string , plain string ) ( sealed string , err error ) {
	if plain == "" { return }
	sealed , err = encryption.ChaChaEncryptString( encryption.SubKey( secret_key , encryption.PurposeRemoteKeys ) , plain )
	return
}

func OpenRemoteKey( secret_key string , sealed string ) ( plain string , err error ) {
	if sealed == "" { return }
	plain , err = encryption.ChaChaDecryptBase64String( encryption.SubKey( secret_key , encryption.PurposeRemoteKeys ) , sealed )
	return
}

// ProjectFields is what an admin sets on a project, for create and edit.
type ProjectFields struct {
	Name        string
	Description string
	BaseURL     string
	Icon        string
	Color       string
	Listed      bool
}

func ( fields *ProjectFields ) check() ( err error ) {
	if validProjectName( fields.Name ) == false {
		err = ErrBadProjectName
		return
	}
	fields.Name = strings.TrimSpace( fields.Name )
	fields.Description = strings.TrimSpace( fields.Description )
	fields.BaseURL , err = NormalizeBaseURL( fields.BaseURL )
	return
}

func CreateProject( store *db.Store , fields ProjectFields , sealed_key string ) ( project *Project , err error ) {
	if err = fields.check(); err != nil { return }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		next_id , tx_err := tx.NextSequence( db.BucketProjects )
		if tx_err != nil { return }
		now := time.Now().UTC()
		project = &Project{
			ID: next_id , Name: fields.Name , Description: fields.Description , BaseURL: fields.BaseURL ,
			Icon: fields.Icon , Color: fields.Color , Listed: fields.Listed , AdminKey: sealed_key ,
			Position: int( next_id ) , CreatedAt: now , UpdatedAt: now ,
		}
		if sealed_key == "" { project.KeyStatus = KeyMissing }
		tx_err = tx.Put( db.BucketProjects , encryption.Uint64ToBytes( project.ID ) , project )
		return
	} )
	if err != nil { project = nil }
	return
}

// GetProject is a project that is not deleted.
func GetProject( store *db.Store , project_id uint64 ) ( project *Project , err error ) {
	project = &Project{}
	if err = store.Get( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , project ); err != nil {
		project = nil
		return
	}
	if project.Deleted() {
		project = nil
		err = db.ErrNotFound
	}
	return
}

// ListProjects is every project not deleted, in catalog order.
func ListProjects( store *db.Store ) ( projects []*Project , err error ) {
	projects = []*Project{}
	err = store.ForEach( db.BucketProjects , func() any { return &Project{} } , func( key []byte , item any ) bool {
		project := item.( *Project )
		if project.Deleted() == false { projects = append( projects , project ) }
		return true
	} )
	sort.SliceStable( projects , func( i , j int ) bool {
		if projects[ i ].Position != projects[ j ].Position { return projects[ i ].Position < projects[ j ].Position }
		return projects[ i ].ID < projects[ j ].ID
	} )
	return
}

// updateProject changes a live project. A deleted one is not found.
func updateProject( store *db.Store , project_id uint64 , mutate func( project *Project ) error ) ( project *Project , err error ) {
	err = store.UpdateValue( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , func() any { return &Project{} } , func( item any ) ( mutate_err error ) {
		project = item.( *Project )
		if project.Deleted() {
			mutate_err = db.ErrNotFound
			return
		}
		if mutate_err = mutate( project ); mutate_err != nil { return }
		project.UpdatedAt = time.Now().UTC()
		return
	} )
	if err != nil { project = nil }
	return
}

// EditProject changes what an admin set. A new address is another app, so
// every key stored for the old one is dropped with it, the admin key and each
// member's own: otherwise the next call would send them to whatever the new
// address is. moved says that happened.
func EditProject( store *db.Store , project_id uint64 , fields ProjectFields ) ( project *Project , moved bool , err error ) {
	if err = fields.check(); err != nil { return }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		project = &Project{}
		if tx_err = tx.Get( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , project ); tx_err != nil { return }
		if project.Deleted() {
			tx_err = db.ErrNotFound
			return
		}
		if project.BaseURL != fields.BaseURL {
			moved = true
			project.AdminKey , project.KeyStatus , project.Health , project.CheckNote , project.CheckedAt = "" , KeyMissing , "" , "" , nil
			if tx_err = clearPersonalKeysTx( tx , func( membership *Membership ) bool { return membership.ProjectID == project_id } ); tx_err != nil { return }
		}
		project.Name , project.Description , project.BaseURL = fields.Name , fields.Description , fields.BaseURL
		project.Icon , project.Color , project.Listed = fields.Icon , fields.Color , fields.Listed
		project.UpdatedAt = time.Now().UTC()
		tx_err = tx.Put( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , project )
		return
	} )
	if err != nil { project , moved = nil , false }
	return
}

// SetProjectKey stores a new admin key, already checked against the project,
// with the role the project reported for it.
func SetProjectKey( store *db.Store , project_id uint64 , sealed_key string ) ( project *Project , err error ) {
	project , err = updateProject( store , project_id , func( project *Project ) ( mutate_err error ) {
		project.AdminKey = sealed_key
		project.KeyStatus , project.CheckNote = KeyOK , ""
		if sealed_key == "" { project.KeyStatus = KeyMissing }
		return
	} )
	return
}

// SetProjectAccess is who a listed project is kept from.
func SetProjectAccess( store *db.Store , project_id uint64 , excluded []uint64 ) ( project *Project , err error ) {
	clean := []uint64{}
	for _ , user_id := range excluded {
		if user_id != 0 && slices.Contains( clean , user_id ) == false { clean = append( clean , user_id ) }
	}
	slices.Sort( clean )
	project , err = updateProject( store , project_id , func( project *Project ) ( mutate_err error ) {
		project.ExcludedUserIDs = clean
		return
	} )
	return
}

// SetUserExcluded keeps one project from one person, or lets them see it again.
func SetUserExcluded( store *db.Store , project_id uint64 , user_id uint64 , excluded bool ) ( project *Project , err error ) {
	project , err = updateProject( store , project_id , func( project *Project ) ( mutate_err error ) {
		project.ExcludedUserIDs = slices.DeleteFunc( project.ExcludedUserIDs , func( id uint64 ) bool { return id == user_id } )
		if excluded { project.ExcludedUserIDs = append( project.ExcludedUserIDs , user_id ) }
		slices.Sort( project.ExcludedUserIDs )
		return
	} )
	return
}

// RecordCheck is what a check found.
func RecordCheck( store *db.Store , project_id uint64 , health string , key_status string , note string ) ( project *Project , err error ) {
	err = store.UpdateValue( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , func() any { return &Project{} } , func( item any ) ( mutate_err error ) {
		project = item.( *Project )
		now := time.Now().UTC()
		project.Health , project.KeyStatus , project.CheckNote , project.CheckedAt = health , key_status , note , &now
		return
	} )
	if err != nil { project = nil }
	return
}

// OrderProjects puts the catalog in the order given; ids left out keep
// their place after the ones named.
func OrderProjects( store *db.Store , ids []uint64 ) ( err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		for position , project_id := range ids {
			project := &Project{}
			if get_err := tx.Get( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , project ); get_err != nil { continue }
			project.Position = position - len( ids )
			if tx_err = tx.Put( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , project ); tx_err != nil { return }
		}
		return
	} )
	return
}

func DeleteProject( store *db.Store , project_id uint64 , actor_id uint64 ) ( err error ) {
	err = store.UpdateValue( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , func() any { return &Project{} } , func( item any ) ( mutate_err error ) {
		project := item.( *Project )
		if project.markDeleted( actor_id ) == false { mutate_err = db.ErrNotFound }
		return
	} )
	return
}

func RestoreProject( store *db.Store , project_id uint64 , actor_id uint64 ) ( err error ) {
	owner_id , err := OwnerID( store )
	if err != nil { return }
	err = store.UpdateValue( db.BucketProjects , encryption.Uint64ToBytes( project_id ) , func() any { return &Project{} } , func( item any ) ( mutate_err error ) {
		mutate_err = item.( *Project ).unmarkDeleted( actor_id , true , owner_id )
		return
	} )
	if errors.Is( err , db.ErrNotFound ) { err = ErrRestoreGone }
	return
}

// PurgeDeletedProjects removes projects past their undo window, with every
// membership of theirs.
func PurgeDeletedProjects( store *db.Store ) ( removed int , err error ) {
	gone := []uint64{}
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		removed , tx_err = tx.DeleteWhere( db.BucketProjects , func() any { return &Project{} } , func( key []byte , item any ) bool {
			project := item.( *Project )
			if project.deletedLongAgo() == false { return false }
			gone = append( gone , project.ID )
			return true
		} )
		if tx_err != nil || len( gone ) == 0 { return }
		_ , tx_err = tx.DeleteWhere( db.BucketMemberships , func() any { return &Membership{} } , func( key []byte , item any ) bool {
			return slices.Contains( gone , item.( *Membership ).ProjectID )
		} )
		return
	} )
	return
}

// Excludes is whether the project is kept from this person.
func ( project *Project ) Excludes( user_id uint64 ) ( result bool ) {
	result = slices.Contains( project.ExcludedUserIDs , user_id )
	return
}

// CanSee is the one access rule: a project is open to someone when it is
// listed and not kept from them. is_admin is the caller's effective role (an
// admin's user-scoped key is a user here): admins see every live project, so
// they can manage and test it.
func CanSee( user_id uint64 , is_admin bool , project *Project ) ( result bool ) {
	if project == nil || project.Deleted() { return }
	if is_admin {
		result = true
		return
	}
	result = project.Listed && project.Excludes( user_id ) == false
	return
}
