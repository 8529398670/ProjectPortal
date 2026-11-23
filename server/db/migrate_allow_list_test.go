package db

import (
	slices "slices"
	testing "testing"

	encryption "projectportal/server/encryption"
)

// Migration 1: a project is allowed to whoever is on it now (not left, not
// kept from it), and to nobody else.
func TestTheExcludeListBecomesAnAllowList( t *testing.T ) {
	store , err := Open( sealingConfig( t ) )
	if err != nil { t.Fatal( err ) }
	defer store.Close()
	member := func( user uint64 , project uint64 , left bool ) {
		record := map[string]any{ "user_id": user , "project_id": project , "via": "joined" }
		if left { record[ "left_at" ] = "2026-11-01T00:00:00Z" }
		key := append( encryption.Uint64ToBytes( user ) , encryption.Uint64ToBytes( project )... )
		if put_err := store.Put( BucketMemberships , key , record ); put_err != nil { t.Fatal( put_err ) }
	}
	project := func( id uint64 , excluded []uint64 ) {
		record := map[string]any{ "id": id , "name": "P" , "listed": true , "base_url": "https://x.test" }
		if excluded != nil { record[ "excluded_user_ids" ] = excluded }
		if put_err := store.Put( BucketProjects , encryption.Uint64ToBytes( id ) , record ); put_err != nil { t.Fatal( put_err ) }
	}
	project( 1 , []uint64{ 3 } )
	project( 2 , nil )
	project( 3 , nil )
	member( 4 , 1 , false ) // on it: allowed
	member( 2 , 1 , false ) // on it: allowed
	member( 5 , 1 , true )  // left it: not allowed
	member( 3 , 1 , false ) // kept from it: still not allowed
	member( 2 , 2 , false )

	if err = store.Write( func( tx *Tx ) error { return migrateAllowList( tx ) } ); err != nil { t.Fatal( err ) }

	type shape struct {
		Name     string   `json:"name"`
		Listed   bool     `json:"listed"`
		Allowed  []uint64 `json:"allowed_user_ids"`
		Excluded []uint64 `json:"excluded_user_ids"`
	}
	read := func( id uint64 ) ( got shape ) {
		if get_err := store.Get( BucketProjects , encryption.Uint64ToBytes( id ) , &got ); get_err != nil { t.Fatal( get_err ) }
		return
	}
	if got := read( 1 ); slices.Equal( got.Allowed , []uint64{ 2 , 4 } ) == false || got.Excluded != nil || got.Name != "P" || got.Listed == false {
		t.Fatalf( "project 1: %+v" , got )
	}
	if got := read( 2 ); slices.Equal( got.Allowed , []uint64{ 2 } ) == false { t.Fatalf( "project 2: %+v" , got ) }
	if got := read( 3 ); len( got.Allowed ) != 0 { t.Fatalf( "project 3 should be off for everyone: %+v" , got ) }
}
