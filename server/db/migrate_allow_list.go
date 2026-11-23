package db

import (
	cmp "cmp"
	json "encoding/json"
	slices "slices"
)

// migrateAllowList is migration 1. Projects used to be open to everyone,
// with a per-project list of people they were kept from
// (excluded_user_ids). They are now off for everyone, with a list of people
// allowed (allowed_user_ids). Nobody should lose an app they already use, so
// each project is allowed to whoever has an active membership on it (not
// left), unless they were kept from it. Everyone else starts off.
func migrateAllowList( tx *Tx ) ( err error ) {
	members := map[string][]string{}
	err = tx.EachRecord( BucketMemberships , func( key []byte , record map[string]any ) ( changed bool , visit_err error ) {
		if left , has := record[ "left_at" ]; has && left != nil { return }
		project , project_ok := record[ "project_id" ].( json.Number )
		user , user_ok := record[ "user_id" ].( json.Number )
		if project_ok == false || user_ok == false { return }
		members[ project.String() ] = append( members[ project.String() ] , user.String() )
		return
	} )
	if err != nil { return }
	err = tx.EachRecord( BucketProjects , func( key []byte , record map[string]any ) ( changed bool , visit_err error ) {
		id , ok := record[ "id" ].( json.Number )
		if ok == false { return }
		excluded := []string{}
		if list , is_list := record[ "excluded_user_ids" ].( []any ); is_list {
			for _ , item := range list {
				if number , is_number := item.( json.Number ); is_number { excluded = append( excluded , number.String() ) }
			}
		}
		allowed := []json.Number{}
		seen := map[string]bool{}
		for _ , user := range members[ id.String() ] {
			if seen[ user ] || slices.Contains( excluded , user ) { continue }
			seen[ user ] = true
			allowed = append( allowed , json.Number( user ) )
		}
		slices.SortFunc( allowed , func( a , b json.Number ) int {
			first , _ := a.Int64()
			second , _ := b.Int64()
			return cmp.Compare( first , second )
		} )
		delete( record , "excluded_user_ids" )
		if len( allowed ) > 0 { record[ "allowed_user_ids" ] = allowed }
		changed = true
		return
	} )
	return
}
