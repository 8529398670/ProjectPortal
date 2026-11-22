package models

import (
	errors "errors"
	fmt "fmt"
	slices "slices"
	sort "sort"
	strconv "strconv"
	strings "strings"
	time "time"
	unicode "unicode"

	db "projectportal/server/db"
	encryption "projectportal/server/encryption"
)

// People across the catalog (phase 3b): one real person, and the accounts
// they have in each app, so duplicates, renames and the same person in
// several apps read as one. The apps never learn about this; it is the
// portal's own bookkeeping.
//
//   - An Account is one account in one app, by its id there, which never
//     changes when it is renamed. Project 0 is this portal: its own users.
//   - A Person holds a set of accounts through AccountLinks. An account
//     belongs to one person at most.
//   - AccountSnapshot is what the last people-sync saw of an account, so
//     the screen and the suggestions need no calls to the apps.
//   - A dismissed pair is two accounts an admin said are not the same
//     person, so that suggestion never comes back.
//
// Nothing is linked without an admin, with one exception: an account a
// membership already ties to a portal user (the portal made it for them, an
// admin matched it, or they attached their own key) is linked to that user,
// since an admin or the person themself already said so.

// PortalProject is the project id the portal's own accounts go under.
const PortalProject uint64 = 0

// Account is one account in one app.
type Account struct {
	ProjectID    uint64 `json:"project_id"`
	RemoteUserID uint64 `json:"remote_user_id"`
}

// Key is the account's key in the people buckets.
func ( account Account ) Key() ( key string ) {
	key = strconv.FormatUint( account.ProjectID , 10 ) + ":" + strconv.FormatUint( account.RemoteUserID , 10 )
	return
}

func accountOfKey( key string ) ( account Account , ok bool ) {
	project , user , found := strings.Cut( key , ":" )
	if found == false { return }
	project_id , project_err := strconv.ParseUint( project , 10 , 64 )
	user_id , user_err := strconv.ParseUint( user , 10 , 64 )
	if project_err != nil || user_err != nil { return }
	account , ok = Account{ ProjectID: project_id , RemoteUserID: user_id } , true
	return
}

// Person is one real person.
type Person struct {
	ID          uint64    `json:"id"`
	DisplayName string    `json:"display_name"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AccountLink puts an account under a person.
type AccountLink struct {
	Account  Account   `json:"account"`
	PersonID uint64    `json:"person_id"`
	LinkedBy uint64    `json:"linked_by,omitempty"` // 0: from a membership
	LinkedAt time.Time `json:"linked_at"`
}

// Device is one signed-in device of an account, as the app reported it.
type Device struct {
	IP         string     `json:"ip,omitempty"`
	UserAgent  string     `json:"user_agent,omitempty"`
	Label      string     `json:"label,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}

// AccountSnapshot is an account as the last people-sync saw it.
type AccountSnapshot struct {
	Account     Account `json:"account"`
	DisplayName string  `json:"display_name"`
	Role        string  `json:"role"`
	// Kind is the app's word, "person" or "service" (empty before contract
	// 3). Portal is the account the portal's own key belongs to there.
	Kind     string `json:"kind,omitempty"`
	Portal   bool   `json:"portal,omitempty"`
	Owner    bool   `json:"owner,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Expired  bool   `json:"expired,omitempty"`
	// Devices is nil when the app cannot list them (no admin_sessions, or
	// the portal's key may not read that account's), empty when it has none.
	Devices  []Device  `json:"devices"` // no omitempty: [] and null differ
	SyncedAt time.Time `json:"synced_at"`
}

// Service is whether the snapshot is a service account, never a person's.
func ( snapshot *AccountSnapshot ) Service() ( result bool ) {
	result = snapshot.Kind == KindService || snapshot.Portal
	return
}

// HandMadePortal is a person's account named like the portal's own: an
// admin account someone made by hand for the portal before Connect existed.
// It is never a real person; it should be deleted in that app.
func ( snapshot *AccountSnapshot ) HandMadePortal() ( result bool ) {
	result = snapshot.Service() == false && NormalName( snapshot.DisplayName ) == NormalName( DefaultPortalName )
	return
}

var (
	ErrServiceLink    = errors.New( "that is a service account: it belongs to no person" )
	ErrBadPersonName  = errors.New( "name must be 1-80 characters" )
	ErrBadNotes       = errors.New( "notes must be at most 2000 characters" )
	ErrSamePerson     = errors.New( "those are the same person" )
)

const personNotesMax = 2000

// NormalName is a display name for comparing: lower case, letters and
// digits only, so "Ann Lee", "ann.lee" and "ANNLEE" meet.
func NormalName( name string ) ( normal string ) {
	builder := strings.Builder{}
	for _ , letter := range strings.ToLower( name ) {
		if unicode.IsLetter( letter ) || unicode.IsDigit( letter ) { builder.WriteRune( letter ) }
	}
	normal = builder.String()
	return
}

func dismissKey( a Account , b Account ) ( key string ) {
	first , second := a.Key() , b.Key()
	if second < first { first , second = second , first }
	key = first + "|" + second
	return
}

// --- the snapshot -------------------------------------------------------------

// ReplaceSnapshots stores what a sync saw of one project's accounts, in place
// of what was there. An account that is gone from the app loses its link
// too: its id will not come back.
func ReplaceSnapshots( store *db.Store , project_id uint64 , snapshots []AccountSnapshot ) ( err error ) {
	seen := map[string]bool{}
	for _ , snapshot := range snapshots { seen[ snapshot.Account.Key() ] = true }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		gone := func( key []byte , item any ) ( result bool ) {
			account , ok := accountOfKey( string( key ) )
			result = ok && account.ProjectID == project_id && seen[ string( key ) ] == false
			return
		}
		if _ , tx_err = tx.DeleteWhere( db.BucketAccountSnapshots , func() any { return &AccountSnapshot{} } , gone ); tx_err != nil { return }
		if _ , tx_err = tx.DeleteWhere( db.BucketAccountLinks , func() any { return &AccountLink{} } , gone ); tx_err != nil { return }
		for index := range snapshots {
			snapshot := snapshots[ index ]
			if snapshot.Account.ProjectID != project_id { continue }
			if tx_err = tx.Put( db.BucketAccountSnapshots , []byte( snapshot.Account.Key() ) , &snapshot ); tx_err != nil { return }
		}
		return
	} )
	return
}

// ForgetProjectAccounts drops a deleted project's snapshots and links.
func ForgetProjectAccounts( store *db.Store , project_id uint64 ) ( err error ) {
	err = ReplaceSnapshots( store , project_id , nil )
	return
}

// --- people ---------------------------------------------------------------------

// PeopleBook is everything the People screen and the suggestions read, in one
// look at the database.
type PeopleBook struct {
	People    map[uint64]*Person
	Links     map[string]*AccountLink
	Snapshots map[string]*AccountSnapshot
	Dismissed map[string]bool
}

// ReadPeople is the whole book.
func ReadPeople( store *db.Store ) ( book *PeopleBook , err error ) {
	book = &PeopleBook{ People: map[uint64]*Person{} , Links: map[string]*AccountLink{} , Snapshots: map[string]*AccountSnapshot{} , Dismissed: map[string]bool{} }
	// A write transaction only for one consistent view: nothing is written.
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		if tx_err = tx.ForEach( db.BucketPeople , func() any { return &Person{} } , func( _ []byte , item any ) bool {
			person := item.( *Person )
			book.People[ person.ID ] = person
			return true
		} ); tx_err != nil { return }
		if tx_err = tx.ForEach( db.BucketAccountLinks , func() any { return &AccountLink{} } , func( key []byte , item any ) bool {
			book.Links[ string( key ) ] = item.( *AccountLink )
			return true
		} ); tx_err != nil { return }
		if tx_err = tx.ForEach( db.BucketAccountSnapshots , func() any { return &AccountSnapshot{} } , func( key []byte , item any ) bool {
			book.Snapshots[ string( key ) ] = item.( *AccountSnapshot )
			return true
		} ); tx_err != nil { return }
		tx_err = tx.ForEach( db.BucketPeopleDismissed , func() any { return &time.Time{} } , func( key []byte , _ any ) bool {
			book.Dismissed[ string( key ) ] = true
			return true
		} )
		return
	} )
	if err != nil { book = nil }
	return
}

// AccountsOf is a person's accounts, in project then id order.
func ( book *PeopleBook ) AccountsOf( person_id uint64 ) ( accounts []Account ) {
	accounts = []Account{}
	for _ , link := range book.Links {
		if link.PersonID == person_id { accounts = append( accounts , link.Account ) }
	}
	sortAccounts( accounts )
	return
}

func sortAccounts( accounts []Account ) {
	sort.Slice( accounts , func( a int , b int ) ( less bool ) {
		if accounts[ a ].ProjectID != accounts[ b ].ProjectID { return accounts[ a ].ProjectID < accounts[ b ].ProjectID }
		less = accounts[ a ].RemoteUserID < accounts[ b ].RemoteUserID
		return
	} )
}

func personTx( tx *db.Tx , person_id uint64 ) ( person *Person , err error ) {
	person = &Person{}
	if err = tx.Get( db.BucketPeople , encryption.Uint64ToBytes( person_id ) , person ); err != nil { person = nil }
	return
}

func newPersonTx( tx *db.Tx , name string ) ( person *Person , err error ) {
	id , err := tx.NextSequence( db.BucketPeople )
	if err != nil { return }
	now := time.Now().UTC()
	name = strings.TrimSpace( name )
	if name == "" { name = "Person " + strconv.FormatUint( id , 10 ) }
	if len( []rune( name ) ) > 80 { name = string( []rune( name )[ :80 ] ) }
	person = &Person{ ID: id , DisplayName: name , CreatedAt: now , UpdatedAt: now }
	err = tx.Put( db.BucketPeople , encryption.Uint64ToBytes( id ) , person )
	return
}

func snapshotTx( tx *db.Tx , account Account ) ( snapshot *AccountSnapshot ) {
	found := &AccountSnapshot{}
	if tx.Get( db.BucketAccountSnapshots , []byte( account.Key() ) , found ) == nil { snapshot = found }
	return
}

// LinkAccounts puts accounts under one person: person_id when it is not 0,
// otherwise the person one of them already has (merging when they have
// several), otherwise a new person named after the first. Service accounts
// are refused. actor_id 0 is a membership, not an admin.
func LinkAccounts( store *db.Store , actor_id uint64 , person_id uint64 , accounts []Account ) ( person *Person , err error ) {
	if len( accounts ) == 0 && person_id == 0 { return nil , errors.New( "no accounts to link" ) }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		person , tx_err = linkAccountsTx( tx , actor_id , person_id , accounts )
		return
	} )
	if err != nil { person = nil }
	return
}

func linkAccountsTx( tx *db.Tx , actor_id uint64 , person_id uint64 , accounts []Account ) ( person *Person , err error ) {
	owners := []uint64{}
	name := ""
	for _ , account := range accounts {
		snapshot := snapshotTx( tx , account )
		if snapshot != nil && snapshot.Service() { return nil , ErrServiceLink }
		if name == "" && snapshot != nil { name = snapshot.DisplayName }
		link := &AccountLink{}
		if tx.Get( db.BucketAccountLinks , []byte( account.Key() ) , link ) == nil && slices.Contains( owners , link.PersonID ) == false {
			owners = append( owners , link.PersonID )
		}
	}
	switch {
	case person_id != 0:
		if person , err = personTx( tx , person_id ); err != nil { return }
	case len( owners ) > 0:
		if person , err = personTx( tx , owners[ 0 ] ); err != nil { return }
	default:
		if person , err = newPersonTx( tx , name ); err != nil { return }
	}
	for _ , other := range owners {
		if other == person.ID { continue }
		if err = mergeTx( tx , person , other ); err != nil { return }
	}
	now := time.Now().UTC()
	for _ , account := range accounts {
		link := &AccountLink{ Account: account , PersonID: person.ID , LinkedBy: actor_id , LinkedAt: now }
		if err = tx.Put( db.BucketAccountLinks , []byte( account.Key() ) , link ); err != nil { return }
	}
	person.UpdatedAt = now
	err = tx.Put( db.BucketPeople , encryption.Uint64ToBytes( person.ID ) , person )
	return
}

// mergeTx moves every account of from_id to into, keeps its notes, and
// deletes it.
func mergeTx( tx *db.Tx , into *Person , from_id uint64 ) ( err error ) {
	from , err := personTx( tx , from_id )
	if err != nil { return }
	moved := []*AccountLink{}
	err = tx.ForEach( db.BucketAccountLinks , func() any { return &AccountLink{} } , func( _ []byte , item any ) bool {
		if link := item.( *AccountLink ); link.PersonID == from_id { moved = append( moved , link ) }
		return true
	} )
	if err != nil { return }
	for _ , link := range moved {
		link.PersonID = into.ID
		if err = tx.Put( db.BucketAccountLinks , []byte( link.Account.Key() ) , link ); err != nil { return }
	}
	if strings.TrimSpace( from.Notes ) != "" {
		into.Notes = strings.TrimSpace( strings.TrimSpace( into.Notes ) + "\n" + from.Notes )
		if len( into.Notes ) > personNotesMax { into.Notes = into.Notes[ :personNotesMax ] }
	}
	err = tx.Delete( db.BucketPeople , encryption.Uint64ToBytes( from_id ) )
	return
}

// MergePeople puts everything of from_id under into_id.
func MergePeople( store *db.Store , into_id uint64 , from_id uint64 ) ( person *Person , err error ) {
	if into_id == from_id { return nil , ErrSamePerson }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		if person , tx_err = personTx( tx , into_id ); tx_err != nil { return }
		if tx_err = mergeTx( tx , person , from_id ); tx_err != nil { return }
		person.UpdatedAt = time.Now().UTC()
		tx_err = tx.Put( db.BucketPeople , encryption.Uint64ToBytes( person.ID ) , person )
		return
	} )
	if err != nil { person = nil }
	return
}

// UnlinkAccount takes an account off its person: a split. It also marks it
// as not that person's other accounts, so the same suggestion does not
// bring it straight back. A person left with no accounts is deleted unless
// it has notes.
func UnlinkAccount( store *db.Store , account Account ) ( err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		link := &AccountLink{}
		if tx_err = tx.Get( db.BucketAccountLinks , []byte( account.Key() ) , link ); tx_err != nil { return }
		if tx_err = tx.Delete( db.BucketAccountLinks , []byte( account.Key() ) ); tx_err != nil { return }
		left := 0
		tx_err = tx.ForEach( db.BucketAccountLinks , func() any { return &AccountLink{} } , func( _ []byte , item any ) bool {
			other := item.( *AccountLink )
			if other.PersonID != link.PersonID { return true }
			left++
			_ = tx.Put( db.BucketPeopleDismissed , []byte( dismissKey( account , other.Account ) ) , time.Now().UTC() )
			return true
		} )
		if tx_err != nil || left > 0 { return }
		person , get_err := personTx( tx , link.PersonID )
		if get_err == nil && strings.TrimSpace( person.Notes ) == "" {
			tx_err = tx.Delete( db.BucketPeople , encryption.Uint64ToBytes( person.ID ) )
		}
		return
	} )
	return
}

// DismissPairs says each account in a is not the person of each in b.
func DismissPairs( store *db.Store , a []Account , b []Account ) ( err error ) {
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		now := time.Now().UTC()
		for _ , first := range a {
			for _ , second := range b {
				if first == second { continue }
				if tx_err = tx.Put( db.BucketPeopleDismissed , []byte( dismissKey( first , second ) ) , now ); tx_err != nil { return }
			}
		}
		return
	} )
	return
}

// EditPerson renames a person and sets their notes.
func EditPerson( store *db.Store , person_id uint64 , name string , notes string ) ( person *Person , err error ) {
	name , notes = strings.TrimSpace( name ) , strings.TrimSpace( notes )
	if length := len( []rune( name ) ); length < 1 || length > 80 { return nil , ErrBadPersonName }
	if len( notes ) > personNotesMax { return nil , ErrBadNotes }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		if person , tx_err = personTx( tx , person_id ); tx_err != nil { return }
		person.DisplayName , person.Notes , person.UpdatedAt = name , notes , time.Now().UTC()
		tx_err = tx.Put( db.BucketPeople , encryption.Uint64ToBytes( person.ID ) , person )
		return
	} )
	if err != nil { person = nil }
	return
}

// LinkMemberships links each account a membership ties to a portal user to
// that user's own account here (project 0), unless an admin split them
// before. It answers how many accounts it linked.
func LinkMemberships( store *db.Store ) ( linked int , err error ) {
	pairs := [][2]Account{}
	err = store.ForEach( db.BucketMemberships , func() any { return &Membership{} } , func( _ []byte , item any ) bool {
		membership := item.( *Membership )
		if membership.Active() && membership.RemoteUserID != 0 {
			pairs = append( pairs , [2]Account{
				{ ProjectID: PortalProject , RemoteUserID: membership.UserID },
				{ ProjectID: membership.ProjectID , RemoteUserID: membership.RemoteUserID },
			} )
		}
		return true
	} )
	if err != nil { return }
	err = store.Write( func( tx *db.Tx ) ( tx_err error ) {
		for _ , pair := range pairs {
			dismissed := &time.Time{}
			if tx.Get( db.BucketPeopleDismissed , []byte( dismissKey( pair[ 0 ] , pair[ 1 ] ) ) , dismissed ) == nil { continue }
			here , there := &AccountLink{} , &AccountLink{}
			here_err := tx.Get( db.BucketAccountLinks , []byte( pair[ 0 ].Key() ) , here )
			there_err := tx.Get( db.BucketAccountLinks , []byte( pair[ 1 ].Key() ) , there )
			if here_err == nil && there_err == nil { continue } // both placed already: an admin's call
			if snapshot := snapshotTx( tx , pair[ 1 ] ); snapshot != nil && snapshot.Service() { continue }
			if _ , tx_err = linkAccountsTx( tx , 0 , 0 , pair[ : ] ); tx_err != nil {
				if errors.Is( tx_err , ErrServiceLink ) {
					tx_err = nil
					continue
				}
				return
			}
			linked++
		}
		return
	} )
	return
}

// AccountLabel is how an account reads in messages: "name (project 3, #12)".
func AccountLabel( snapshot *AccountSnapshot , account Account ) ( label string ) {
	name := "?"
	if snapshot != nil { name = snapshot.DisplayName }
	label = fmt.Sprintf( "%s (project %d, #%d)" , name , account.ProjectID , account.RemoteUserID )
	return
}
