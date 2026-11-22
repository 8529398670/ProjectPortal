package models

import (
	slices "slices"
	testing "testing"
	time "time"
)

func snap( project uint64 , id uint64 , name string , devices ...Device ) ( snapshot AccountSnapshot ) {
	snapshot = AccountSnapshot{ Account: Account{ ProjectID: project , RemoteUserID: id } , DisplayName: name , Role: "user" , Kind: "person" , Devices: devices }
	return
}

func TestSuggestFindsTheSamePerson( t *testing.T ) {
	store := testStore( t )
	now := time.Now().UTC()
	phone := Device{ IP: "10.0.0.7" , UserAgent: "Safari iPhone" , CreatedAt: now }
	if err := ReplaceSnapshots( store , 1 , []AccountSnapshot{ snap( 1 , 10 , "Ann Lee" ) , snap( 1 , 11 , "Bob" , phone ) , snap( 1 , 12 , "ann.lee" ) } ); err != nil { t.Fatal( err ) }
	if err := ReplaceSnapshots( store , 2 , []AccountSnapshot{ snap( 2 , 20 , "ANN LEE" ) , snap( 2 , 21 , "Robert" , phone ) } ); err != nil { t.Fatal( err ) }
	book , err := ReadPeople( store )
	if err != nil { t.Fatal( err ) }
	suggestions := Suggest( book )
	pairs := map[string][]string{}
	for _ , s := range suggestions { pairs[ s.A.Accounts[ 0 ].Key() + "~" + s.B.Accounts[ 0 ].Key() ] = s.Reasons }
	if slices.Contains( pairs[ "1:10~2:20" ] , ReasonSameName ) == false { t.Errorf( "same name across apps missed: %v" , pairs ) }
	if slices.Contains( pairs[ "1:10~1:12" ] , ReasonSameName ) == false { t.Errorf( "duplicate in one app missed: %v" , pairs ) }
	if reasons := pairs[ "1:11~2:21" ]; slices.Contains( reasons , ReasonSharedIP ) == false || slices.Contains( reasons , ReasonSameAgent ) == false {
		t.Errorf( "shared device missed: %v" , pairs )
	}
}

func TestServiceAndHandMadePortalAccountsAreNeverPeople( t *testing.T ) {
	store := testStore( t )
	service := snap( 1 , 2 , "ProjectPortal" )
	service.Kind = KindService
	hand := snap( 2 , 3 , "ProjectPortal" )
	hand.Role = "admin"
	if err := ReplaceSnapshots( store , 1 , []AccountSnapshot{ service } ); err != nil { t.Fatal( err ) }
	if err := ReplaceSnapshots( store , 2 , []AccountSnapshot{ hand } ); err != nil { t.Fatal( err ) }
	book , _ := ReadPeople( store )
	if suggestions := Suggest( book ); len( suggestions ) != 0 { t.Fatalf( "service accounts suggested: %+v" , suggestions ) }
	if book.Snapshots[ "2:3" ].HandMadePortal() == false { t.Fatal( "hand-made portal account not spotted" ) }
	if _ , err := LinkAccounts( store , 1 , 0 , []Account{ service.Account } ); err != ErrServiceLink { t.Fatalf( "service account linked: %v" , err ) }
}

func TestLinkMergeSplitAndDismiss( t *testing.T ) {
	store := testStore( t )
	a , b , c := snap( 1 , 1 , "Ann" ) , snap( 2 , 1 , "Ann" ) , snap( 3 , 1 , "Ann" )
	for _ , each := range []AccountSnapshot{ a , b , c } {
		if err := ReplaceSnapshots( store , each.Account.ProjectID , []AccountSnapshot{ each } ); err != nil { t.Fatal( err ) }
	}
	first , err := LinkAccounts( store , 9 , 0 , []Account{ a.Account , b.Account } )
	if err != nil || first.DisplayName != "Ann" { t.Fatalf( "link: %+v %v" , first , err ) }
	second , _ := LinkAccounts( store , 9 , 0 , []Account{ c.Account } )
	if _ , err = MergePeople( store , first.ID , second.ID ); err != nil { t.Fatal( err ) }
	book , _ := ReadPeople( store )
	if len( book.People ) != 1 || len( book.AccountsOf( first.ID ) ) != 3 { t.Fatalf( "merge: %d people, %v" , len( book.People ) , book.AccountsOf( first.ID ) ) }
	if err = UnlinkAccount( store , c.Account ); err != nil { t.Fatal( err ) }
	book , _ = ReadPeople( store )
	for _ , s := range Suggest( book ) {
		for _ , group := range []Group{ s.A , s.B } {
			if slices.Contains( group.Accounts , c.Account ) { t.Fatalf( "a split account came straight back: %+v" , s ) }
		}
	}
	// Gone from its app: the link goes with it.
	if err = ReplaceSnapshots( store , 2 , nil ); err != nil { t.Fatal( err ) }
	book , _ = ReadPeople( store )
	if len( book.AccountsOf( first.ID ) ) != 1 { t.Fatalf( "vanished account kept its link: %v" , book.AccountsOf( first.ID ) ) }
}

func TestDismissStopsASuggestion( t *testing.T ) {
	store := testStore( t )
	a , b := snap( 1 , 1 , "Sam" ) , snap( 2 , 1 , "Sam" )
	_ = ReplaceSnapshots( store , 1 , []AccountSnapshot{ a } )
	_ = ReplaceSnapshots( store , 2 , []AccountSnapshot{ b } )
	if err := DismissPairs( store , []Account{ a.Account } , []Account{ b.Account } ); err != nil { t.Fatal( err ) }
	book , _ := ReadPeople( store )
	if suggestions := Suggest( book ); len( suggestions ) != 0 { t.Fatalf( "dismissed pair suggested: %+v" , suggestions ) }
}

func TestNormalName( t *testing.T ) {
	if NormalName( "Ann-Lee 2" ) != "annlee2" || NormalName( "Projekt Portal" ) == NormalName( DefaultPortalName ) { t.Fatal( NormalName( "Ann-Lee 2" ) ) }
}

func TestNoDevicesIsNotUnknownDevices( t *testing.T ) {
	store := testStore( t )
	listed , unknown := snap( 1 , 1 , "A" ) , snap( 1 , 2 , "B" )
	listed.Devices = []Device{}
	_ = ReplaceSnapshots( store , 1 , []AccountSnapshot{ listed , unknown } )
	book , _ := ReadPeople( store )
	if book.Snapshots[ "1:1" ].Devices == nil || book.Snapshots[ "1:2" ].Devices != nil { t.Fatal( "listed-but-none and unknown devices mixed up" ) }
}
