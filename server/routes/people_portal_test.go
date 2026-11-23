package routes

import (
	strings "strings"
	testing "testing"

	fiber "github.com/gofiber/fiber/v3"

	models "projectportal/server/models"
)

// Backfill: a person known only from the apps gets a portal account from the
// People screen, and their app accounts become its own.
func TestPersonGetsAPortalAccount( t *testing.T ) {
	fixture := newAuthFixture( t )
	project , err := models.CreateProject( fixture.store , models.ProjectFields{ Name: "Board" , BaseURL: "http://127.0.0.1:1" , Listed: true } , "" )
	if err != nil { t.Fatal( err ) }
	_ = models.ReplaceSnapshots( fixture.store , project.ID , []models.AccountSnapshot{
		{ Account: models.Account{ ProjectID: project.ID , RemoteUserID: 3 } , DisplayName: "Bob" , Role: "user" , Kind: "person" },
		{ Account: models.Account{ ProjectID: project.ID , RemoteUserID: 9 } , DisplayName: "Boss" , Role: "user" , Kind: "person" , Owner: true },
	} )
	person , err := models.LinkAccounts( fixture.store , fixture.owner.ID , 0 , []models.Account{
		{ ProjectID: project.ID , RemoteUserID: 3 } , { ProjectID: project.ID , RemoteUserID: 9 } } )
	if err != nil { t.Fatal( err ) }
	owner := fixture.signIn( t , fixture.owner.ID )
	path := "/api/admin/people/" + idText( person.ID ) + "/portal-user"
	status , reply := send( t , owner , path , fiber.Map{ "role": "user" } )
	if status != fiber.StatusCreated { t.Fatalf( "create: %d %v" , status , reply ) }
	if strings.HasPrefix( reply[ "login_path" ].( string ) , "/login/" ) == false { t.Fatalf( "no login link: %v" , reply ) }
	// The owner's account there is never signed into from outside.
	if added := reply[ "memberships_added" ].( float64 ); added != 1 { t.Fatalf( "memberships added: %v" , added ) }
	user_id := uint64( reply[ "user_id" ].( float64 ) )
	user , err := models.GetUser( fixture.store , user_id )
	if err != nil || user.DisplayName != person.DisplayName { t.Fatalf( "user: %v %v" , user , err ) }
	membership , err := models.GetMembership( fixture.store , user_id , project.ID )
	if err != nil || membership.RemoteUserID != 3 || membership.Via != models.ViaLinked { t.Fatalf( "membership: %+v %v" , membership , err ) }
	stored , _ := models.GetProject( fixture.store , project.ID )
	if stored.Allows( user_id ) == false { t.Fatal( "not allowed the project" ) }
	_ , list := owner.do( t , "GET" , "/api/admin/people" , nil )
	people := list[ "people" ].( []any )
	if len( people ) != 1 || len( people[ 0 ].( map[string]any )[ "accounts" ].( []any ) ) != 3 { t.Fatalf( "people: %v" , people ) }
	// A second one is refused: link the one they have.
	if status , _ = send( t , owner , path , fiber.Map{ "role": "user" } ); status != fiber.StatusConflict { t.Fatalf( "second account: %d" , status ) }
}

// Linking a portal account someone already has to a person does the same.
func TestLinkingAPortalAccountMakesMemberships( t *testing.T ) {
	fixture := newAuthFixture( t )
	project , err := models.CreateProject( fixture.store , models.ProjectFields{ Name: "Board" , BaseURL: "http://127.0.0.1:1" , Listed: true } , "" )
	if err != nil { t.Fatal( err ) }
	_ = models.ReplaceSnapshots( fixture.store , project.ID , []models.AccountSnapshot{
		{ Account: models.Account{ ProjectID: project.ID , RemoteUserID: 3 } , DisplayName: "alice" , Role: "user" , Kind: "person" },
	} )
	_ = models.RecordPortalUser( fixture.store , fixture.alice )
	owner := fixture.signIn( t , fixture.owner.ID )
	status , reply := send( t , owner , "/api/admin/people/link" , fiber.Map{ "accounts": []fiber.Map{
		{ "project_id": 0 , "remote_user_id": fixture.alice.ID } , { "project_id": project.ID , "remote_user_id": 3 } } } )
	if status != 200 || reply[ "memberships_added" ] != float64( 1 ) { t.Fatalf( "link: %d %v" , status , reply ) }
	membership , err := models.GetMembership( fixture.store , fixture.alice.ID , project.ID )
	if err != nil || membership.RemoteUserID != 3 { t.Fatalf( "membership: %+v %v" , membership , err ) }
}
