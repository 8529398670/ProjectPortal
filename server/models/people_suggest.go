package models

import (
	sort "sort"
	strconv "strconv"
	time "time"
)

// Why two accounts may be the same person. Each is a hint, never proof:
// an admin confirms or dismisses every suggestion.
const (
	ReasonSameName  = "same_name"  // the same name, ignoring case, spaces and punctuation
	ReasonSharedIP  = "shared_ip"  // devices on the same address within sharedIPWindow
	ReasonSameAgent = "same_agent" // devices with the very same browser string
	ReasonCloseIn   = "close_in"   // signed in within closeInWindow of each other
)

const (
	sharedIPWindow   = 7 * 24 * time.Hour
	closeInWindow    = 10 * time.Minute
	suggestThreshold = 2
)

var reasonWeight = map[string]int{ ReasonSameName: 3 , ReasonSharedIP: 2 , ReasonSameAgent: 1 , ReasonCloseIn: 1 }

// Group is either a person (PersonID set) or one account nobody has placed.
type Group struct {
	PersonID uint64    `json:"person_id,omitempty"`
	Accounts []Account `json:"accounts"`
}

func ( group Group ) key() ( key string ) {
	if group.PersonID != 0 {
		key = "p" + strconv.FormatUint( group.PersonID , 10 )
		return
	}
	key = "a" + group.Accounts[ 0 ].Key()
	return
}

// Suggestion is two groups that look like one person.
type Suggestion struct {
	A       Group    `json:"a"`
	B       Group    `json:"b"`
	Score   int      `json:"score"`
	Reasons []string `json:"reasons"`
}

// Suggest is every pair of groups worth an admin's look, best first. Service
// accounts and hand-made portal accounts are never in one, nor is a pair an
// admin dismissed.
func Suggest( book *PeopleBook ) ( suggestions []Suggestion ) {
	suggestions = []Suggestion{}
	groups := map[string]Group{}
	group_of := map[string]string{}
	for key , snapshot := range book.Snapshots {
		if snapshot.Service() || snapshot.HandMadePortal() { continue }
		group := Group{ Accounts: []Account{ snapshot.Account } }
		if link := book.Links[ key ]; link != nil { group = Group{ PersonID: link.PersonID } }
		group_key := group.key()
		existing , found := groups[ group_key ]
		if found { group = existing }
		if group.PersonID != 0 && found { group.Accounts = append( group.Accounts , snapshot.Account ) }
		if group.PersonID != 0 && found == false { group.Accounts = []Account{ snapshot.Account } }
		groups[ group_key ] , group_of[ key ] = group , group_key
	}
	keys := make( []string , 0 , len( book.Snapshots ) )
	for key := range group_of { keys = append( keys , key ) }
	sort.Strings( keys )
	type pairKey struct{ a , b string }
	found := map[pairKey]*Suggestion{}
	reasons := map[pairKey]map[string]bool{}
	blocked := map[pairKey]bool{}
	for i , first_key := range keys {
		first := book.Snapshots[ first_key ]
		for _ , second_key := range keys[ i + 1: ] {
			second := book.Snapshots[ second_key ]
			group_a , group_b := group_of[ first_key ] , group_of[ second_key ]
			if group_a == group_b { continue }
			if group_b < group_a { group_a , group_b = group_b , group_a }
			pair := pairKey{ group_a , group_b }
			// Dismissed: an admin said that pair of groups is not one person.
			// Two accounts in one app are suggested too: that is a duplicate.
			if book.Dismissed[ dismissKey( first.Account , second.Account ) ] {
				blocked[ pair ] = true
				continue
			}
			hits := pairReasons( first , second )
			if len( hits ) == 0 { continue }
			if reasons[ pair ] == nil { reasons[ pair ] = map[string]bool{} }
			for _ , hit := range hits { reasons[ pair ][ hit ] = true }
			if found[ pair ] == nil { found[ pair ] = &Suggestion{ A: groups[ group_a ] , B: groups[ group_b ] } }
		}
	}
	for pair , suggestion := range found {
		if blocked[ pair ] { continue }
		for reason := range reasons[ pair ] {
			suggestion.Reasons = append( suggestion.Reasons , reason )
			suggestion.Score += reasonWeight[ reason ]
		}
		if suggestion.Score < suggestThreshold { continue }
		sort.Slice( suggestion.Reasons , func( a int , b int ) bool { return reasonWeight[ suggestion.Reasons[ a ] ] > reasonWeight[ suggestion.Reasons[ b ] ] } )
		sortAccounts( suggestion.A.Accounts )
		sortAccounts( suggestion.B.Accounts )
		suggestions = append( suggestions , *suggestion )
	}
	sort.Slice( suggestions , func( a int , b int ) ( less bool ) {
		if suggestions[ a ].Score != suggestions[ b ].Score { return suggestions[ a ].Score > suggestions[ b ].Score }
		less = suggestions[ a ].A.key() + suggestions[ a ].B.key() < suggestions[ b ].A.key() + suggestions[ b ].B.key()
		return
	} )
	return
}

func pairReasons( a *AccountSnapshot , b *AccountSnapshot ) ( reasons []string ) {
	if name := NormalName( a.DisplayName ); name != "" && name == NormalName( b.DisplayName ) { reasons = append( reasons , ReasonSameName ) }
	ip , agent , close := false , false , false
	for _ , first := range a.Devices {
		for _ , second := range b.Devices {
			if first.IP != "" && first.IP == second.IP && within( lastSeen( first ) , lastSeen( second ) , sharedIPWindow ) { ip = true }
			if first.UserAgent != "" && first.UserAgent == second.UserAgent { agent = true }
			if within( first.CreatedAt , second.CreatedAt , closeInWindow ) { close = true }
		}
	}
	if ip { reasons = append( reasons , ReasonSharedIP ) }
	if agent { reasons = append( reasons , ReasonSameAgent ) }
	if close { reasons = append( reasons , ReasonCloseIn ) }
	return
}

func lastSeen( device Device ) ( at time.Time ) {
	at = device.CreatedAt
	if device.LastSeenAt != nil && device.LastSeenAt.After( at ) { at = *device.LastSeenAt }
	return
}

func within( a time.Time , b time.Time , window time.Duration ) ( result bool ) {
	if a.IsZero() || b.IsZero() { return }
	gap := a.Sub( b )
	if gap < 0 { gap = -gap }
	result = gap <= window
	return
}
