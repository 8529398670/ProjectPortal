package models

import (
	errors "errors"
	slices "slices"
	strings "strings"
	time "time"

	db "projectportal/server/db"
	encryption "projectportal/server/encryption"
)

// Membership is one person on one project: which account they have there,
// and how the portal signs them into it.
//
// RemoteUserID is their account on the project. The portal signs them in by
// asking the project, with its admin key, for a fresh login link for that
// account. PersonalKey, when set, is their own key on that project instead
// (sealed like Project.AdminKey): the way in for an account the admin key may
// not reissue for, such as the project's owner or another admin.
type Membership struct {
	UserID       uint64     `json:"user_id"`
	ProjectID    uint64     `json:"project_id"`
	RemoteUserID uint64     `json:"remote_user_id,omitempty"`
	RemoteName   string     `json:"remote_name,omitempty"`
	RemoteRole   string     `json:"remote_role,omitempty"`
	// RemoteExpiresAt mirrors when their account there stops working (an app
	// with account_expiry), as last set or seen from the portal.
	RemoteExpiresAt *time.Time `json:"remote_expires_at,omitempty"`
	PersonalKey  string     `json:"personal_key,omitempty"`
	Via          string     `json:"via"`
	Favorite     bool       `json:"favorite,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	JoinedAt     time.Time  `json:"joined_at"`
	LastOpenedAt *time.Time `json:"last_opened_at,omitempty"`
	// LeftAt is set while someone has left the project here: the record stays
	// so Undo can put it back, and joining again reuses the account the
	// project already has for them instead of making a second one.
	LeftAt *time.Time `json:"left_at,omitempty"`
}

// Active is a membership that has not been left.
func ( membership *Membership ) Active() ( result bool ) {
	result = membership != nil && membership.LeftAt == nil
	return
}

const (
	ViaJoined  = "joined"  // the portal made the account
	ViaLinked  = "linked"  // an admin matched an account the project already had
	ViaOwnKey  = "own_key" // the person attached their own key

	maxTags      = 10
	maxTagLength = 30
)

var (
	ErrAlreadyMember = errors.New( "already on this project" )
	ErrBadTags       = errors.New( "up to 10 tags, each 1-30 characters" )
)

func membershipKey( user_id uint64 , project_id uint64 ) ( key []byte ) {
	key = append( encryption.Uint64ToBytes( user_id ) , encryption.Uint64ToBytes( project_id )... )
	return
}

func GetMembership( store *db.Store , user_id uint64 , project_id uint64 ) ( membership *Membership , err error ) {
	membership = &Membership{}
	if err = store.Get( db.BucketMemberships , membershipKey( user_id , project_id ) , membership ); err != nil {
		membership = nil
	}
	return
}

// MembershipsOf is every project one person is on, by project id.
func MembershipsOf( store *db.Store , user_id uint64 ) ( memberships map[uint64]*Membership , err error ) {
	memberships = map[uint64]*Membership{}
	err = store.ForEach( db.BucketMemberships , func() any { return &Membership{} } , func( key []byte , item any ) bool {
		membership := item.( *Membership )
		if membership.UserID == user_id && membership.Active() { memberships[ membership.ProjectID ] = membership }
		return true
	} )
	return
}

// MembersOf is everyone on one project, by portal user id.
func MembersOf( store *db.Store , project_id uint64 ) ( memberships map[uint64]*Membership , err error ) {
	memberships = map[uint64]*Membership{}
	err = store.ForEach( db.BucketMemberships , func() any { return &Membership{} } , func( key []byte , item any ) bool {
		membership := item.( *Membership )
		if membership.ProjectID == project_id && membership.Active() { memberships[ membership.UserID ] = membership }
		return true
	} )
	return
}

// MemberCounts is how many people are on each project.
func MemberCounts( store *db.Store ) ( counts map[uint64]int , err error ) {
	counts = map[uint64]int{}
	err = store.ForEach( db.BucketMemberships , func() any { return &Membership{} } , func( key []byte , item any ) bool {
		if membership := item.( *Membership ); membership.Active() { counts[ membership.ProjectID ]++ }
		return true
	} )
	return
}

// AddMembership records an account on a project. replace says whether an
// existing membership may be overwritten (an admin linking someone again);
// otherwise it is ErrAlreadyMember. Favorite and tags survive a replace.
func AddMembership( store *db.Store , membership Membership , replace bool ) ( stored *Membership , err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		key := membershipKey( membership.UserID , membership.ProjectID )
		existing := &Membership{}
		if get_err := tx.Get( db.BucketMemberships , key , existing ); get_err == nil {
			if replace == false && existing.Active() {
				tx_err = ErrAlreadyMember
				return
			}
			membership.Favorite , membership.Tags = existing.Favorite , existing.Tags
		}
		membership.JoinedAt = time.Now().UTC()
		stored = &membership
		tx_err = tx.Put( db.BucketMemberships , key , stored )
		return
	} )
	if err != nil { stored = nil }
	return
}

func updateMembership( store *db.Store , user_id uint64 , project_id uint64 , mutate func( membership *Membership ) error ) ( membership *Membership , err error ) {
	err = store.UpdateValue( db.BucketMemberships , membershipKey( user_id , project_id ) , func() any { return &Membership{} } , func( item any ) ( mutate_err error ) {
		membership = item.( *Membership )
		mutate_err = mutate( membership )
		return
	} )
	if err != nil { membership = nil }
	return
}

func SetFavorite( store *db.Store , user_id uint64 , project_id uint64 , favorite bool ) ( membership *Membership , err error ) {
	membership , err = updateMembership( store , user_id , project_id , func( membership *Membership ) ( mutate_err error ) {
		membership.Favorite = favorite
		return
	} )
	return
}

// CleanTags trims, lowercases and de-duplicates tags, and checks the limits.
func CleanTags( tags []string ) ( clean []string , err error ) {
	clean = []string{}
	for _ , tag := range tags {
		tag = strings.ToLower( strings.TrimSpace( tag ) )
		if tag == "" || slices.Contains( clean , tag ) { continue }
		if len( tag ) > maxTagLength {
			err = ErrBadTags
			return
		}
		clean = append( clean , tag )
	}
	if len( clean ) > maxTags { err = ErrBadTags }
	return
}

func SetTags( store *db.Store , user_id uint64 , project_id uint64 , tags []string ) ( membership *Membership , err error ) {
	clean , err := CleanTags( tags )
	if err != nil { return }
	membership , err = updateMembership( store , user_id , project_id , func( membership *Membership ) ( mutate_err error ) {
		membership.Tags = clean
		return
	} )
	return
}

// SetPersonalKey attaches (or, with "", removes) a person's own key. A
// membership made only by attaching a key goes when the key does.
func SetPersonalKey( store *db.Store , user_id uint64 , project_id uint64 , sealed_key string ) ( membership *Membership , err error ) {
	membership , err = updateMembership( store , user_id , project_id , func( membership *Membership ) ( mutate_err error ) {
		membership.PersonalKey = sealed_key
		return
	} )
	if err == nil && sealed_key == "" && membership.Via == ViaOwnKey && membership.RemoteUserID == 0 {
		err = RemoveMembership( store , user_id , project_id )
	}
	return
}

func MarkOpened( store *db.Store , user_id uint64 , project_id uint64 ) ( err error ) {
	_ , err = updateMembership( store , user_id , project_id , func( membership *Membership ) ( mutate_err error ) {
		now := time.Now().UTC()
		membership.LastOpenedAt = &now
		return
	} )
	return
}

// LeaveProject marks a membership left. It stays for RestoreWindow, so Undo
// can put it back, and its account on the project is reused by a later join.
func LeaveProject( store *db.Store , user_id uint64 , project_id uint64 ) ( err error ) {
	_ , err = updateMembership( store , user_id , project_id , func( membership *Membership ) ( mutate_err error ) {
		if membership.Active() == false {
			mutate_err = db.ErrNotFound
			return
		}
		now := time.Now().UTC()
		membership.LeftAt = &now
		return
	} )
	return
}

// UndoLeave puts a left membership back.
func UndoLeave( store *db.Store , user_id uint64 , project_id uint64 ) ( membership *Membership , err error ) {
	membership , err = updateMembership( store , user_id , project_id , func( membership *Membership ) ( mutate_err error ) {
		if membership.LeftAt == nil {
			mutate_err = ErrRestoreGone
			return
		}
		if time.Since( *membership.LeftAt ) > RestoreWindow {
			mutate_err = ErrRestoreExpired
			return
		}
		membership.LeftAt = nil
		return
	} )
	if errors.Is( err , db.ErrNotFound ) { err = ErrRestoreGone }
	return
}

// RemoveMembership forgets a membership entirely.
func RemoveMembership( store *db.Store , user_id uint64 , project_id uint64 ) ( err error ) {
	err = store.Delete( db.BucketMemberships , membershipKey( user_id , project_id ) )
	return
}

// SyncRemoteAccount applies a change made to an account on a project to every
// membership here that points at it.
func SyncRemoteAccount( store *db.Store , project_id uint64 , remote_user_id uint64 , mutate func( membership *Membership ) ) ( err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		changed := []*Membership{}
		tx_err = tx.ForEach( db.BucketMemberships , func() any { return &Membership{} } , func( key []byte , item any ) bool {
			membership := item.( *Membership )
			if membership.ProjectID == project_id && membership.RemoteUserID == remote_user_id { changed = append( changed , membership ) }
			return true
		} )
		if tx_err != nil { return }
		for _ , membership := range changed {
			mutate( membership )
			if tx_err = tx.Put( db.BucketMemberships , membershipKey( membership.UserID , membership.ProjectID ) , membership ); tx_err != nil { return }
		}
		return
	} )
	return
}

// ForgetRemoteAccount removes every membership pointing at an account that
// was deleted on the project.
func ForgetRemoteAccount( store *db.Store , project_id uint64 , remote_user_id uint64 ) ( removed int , err error ) {
	removed , err = store.DeleteWhere( db.BucketMemberships , func() any { return &Membership{} } , func( key []byte , item any ) bool {
		membership := item.( *Membership )
		return membership.ProjectID == project_id && membership.RemoteUserID == remote_user_id
	} )
	return
}

// PurgeLeftKeys drops the personal key of a membership left longer than the
// undo window: nothing will use it again, and a key is not kept for nothing.
// The record itself stays, so joining again reuses the project's account.
func PurgeLeftKeys( store *db.Store ) ( err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		tx_err = clearPersonalKeysTx( tx , func( membership *Membership ) bool {
			return membership.LeftAt != nil && time.Since( *membership.LeftAt ) > RestoreWindow
		} )
		return
	} )
	return
}

// clearPersonalKeysTx drops the personal key of every membership match
// picks. Collected first and written after: bolt's cursor must not see its
// bucket change under it.
func clearPersonalKeysTx( tx *db.Tx , match func( membership *Membership ) bool ) ( err error ) {
	keys , changed := [][]byte{} , []*Membership{}
	err = tx.ForEach( db.BucketMemberships , func() any { return &Membership{} } , func( key []byte , item any ) bool {
		membership := item.( *Membership )
		if membership.PersonalKey != "" && match( membership ) {
			membership.PersonalKey = ""
			keys , changed = append( keys , key ) , append( changed , membership )
		}
		return true
	} )
	if err != nil { return }
	for index , key := range keys {
		if err = tx.Put( db.BucketMemberships , key , changed[ index ] ); err != nil { return }
	}
	return
}

// PurgeOrphanMemberships drops the memberships of accounts that no longer
// exist here. Deleting an account in the template does not know about this
// bucket, so the purge catches up.
func PurgeOrphanMemberships( store *db.Store ) ( removed int , err error ) {
	users , err := ListUsers( store )
	if err != nil { return }
	alive := map[uint64]bool{}
	for _ , user := range users { alive[ user.ID ] = true }
	removed , err = store.DeleteWhere( db.BucketMemberships , func() any { return &Membership{} } , func( key []byte , item any ) bool {
		return alive[ item.( *Membership ).UserID ] == false
	} )
	return
}
