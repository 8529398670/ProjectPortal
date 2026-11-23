package models

import (
	testing "testing"
)

// Projects are off for everyone until an admin allows a person; admins see
// every live project; unlisted hides one even from people allowed it.
func TestAProjectIsOffUntilAllowed( t *testing.T ) {
	store := testStore( t )
	project , err := CreateProject( store , ProjectFields{ Name: "Board" , BaseURL: "https://board.test" , Listed: true } , "" )
	if err != nil { t.Fatal( err ) }
	if CanSee( 7 , false , project ) { t.Fatal( "a new project is open to someone nobody allowed" ) }
	if CanSee( 7 , true , project ) == false { t.Fatal( "an admin cannot see it" ) }
	if project , err = SetUserAllowed( store , project.ID , 7 , true ); err != nil { t.Fatal( err ) }
	if CanSee( 7 , false , project ) == false { t.Fatal( "allowed, yet hidden" ) }
	if CanSee( 8 , false , project ) { t.Fatal( "allowing one person opened it to another" ) }
	project.Listed = false
	if CanSee( 7 , false , project ) { t.Fatal( "unlisted, yet shown" ) }
	project.Listed = true
	if project , err = SetUserAllowed( store , project.ID , 7 , false ); err != nil { t.Fatal( err ) }
	if CanSee( 7 , false , project ) { t.Fatal( "switched off, yet shown" ) }
}
