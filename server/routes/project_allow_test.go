package routes

import (
	http "net/http"
	testing "testing"

	fiber "github.com/gofiber/fiber/v3"

	models "projectportal/server/models"
)

// Every project is off for a person until an admin switches it on for them;
// then it is in their list and they can join it. Switching it off again
// takes it away.
func TestAProjectIsOffForSomeoneUntilAnAdminAllowsIt( t *testing.T ) {
	fixture := newAuthFixture( t )
	app := newFakeApp( t )
	sealed , err := models.SealRemoteKey( fixture.cfg.SecretKey , "portalkey.secretsecretsecret" )
	if err != nil { t.Fatal( err ) }
	project , err := models.CreateProject( fixture.store , models.ProjectFields{ Name: "Board" , BaseURL: app.server.URL , Listed: true } , sealed )
	if err != nil { t.Fatal( err ) }
	alice , owner := fixture.signIn( t , fixture.alice.ID ) , fixture.signIn( t , fixture.owner.ID )
	listed := func() ( count int ) {
		status , reply := alice.do( t , http.MethodGet , "/api/projects" , nil )
		if status != fiber.StatusOK { t.Fatalf( "list: %d %v" , status , reply ) }
		projects , _ := reply[ "projects" ].( []any )
		count = len( projects )
		return
	}
	join := "/api/projects/" + idText( project.ID ) + "/join"
	access := "/api/admin/projects/" + idText( project.ID ) + "/access/" + idText( fixture.alice.ID )

	if listed() != 0 { t.Fatal( "a new project is in her list" ) }
	if status , reply := send( t , alice , join , fiber.Map{} ); status != fiber.StatusNotFound { t.Fatalf( "join before allowed: %d %v" , status , reply ) }
	if app.called( "/api/admin/users" ) { t.Fatal( "the app was asked to make an account" ) }

	// The admin's per-person list shows it off.
	mine := "/api/admin/users/" + idText( fixture.alice.ID ) + "/projects"
	status , reply := owner.do( t , http.MethodGet , mine , nil )
	rows , _ := reply[ "rows" ].( []any )
	if status != fiber.StatusOK || len( rows ) != 1 || rows[ 0 ].( map[string]any )[ "allowed" ] != false { t.Fatalf( "per-person list: %d %v" , status , reply ) }
	if status , _ := alice.do( t , http.MethodGet , mine , nil ); status != fiber.StatusForbidden { t.Fatalf( "a user read the per-person list: %d" , status ) }

	if status , reply := send( t , owner , access , fiber.Map{ "allowed": true } ); status != fiber.StatusOK { t.Fatalf( "allow: %d %v" , status , reply ) }
	if listed() != 1 { t.Fatal( "allowed, yet not in her list" ) }
	if status , reply := send( t , alice , join , fiber.Map{} ); status != fiber.StatusOK && status != fiber.StatusCreated { t.Fatalf( "join once allowed: %d %v" , status , reply ) }
	_ , reply = owner.do( t , http.MethodGet , mine , nil )
	row := reply[ "rows" ].( []any )[ 0 ].( map[string]any )
	if row[ "allowed" ] != true || row[ "member" ] != true { t.Fatalf( "after joining: %v" , row ) }

	if status , reply := send( t , owner , access , fiber.Map{ "allowed": false } ); status != fiber.StatusOK { t.Fatalf( "switch off: %d %v" , status , reply ) }
	if listed() != 0 { t.Fatal( "switched off, yet still in her list" ) }
	open := "/api/projects/" + idText( project.ID ) + "/open"
	if status , reply := send( t , alice , open , fiber.Map{} ); status != fiber.StatusNotFound { t.Fatalf( "open once switched off: %d %v" , status , reply ) }
}

// Putting someone on a project is allowing them it.
func TestPuttingSomeoneOnAProjectAllowsIt( t *testing.T ) {
	fixture := newAuthFixture( t )
	app := newFakeApp( t )
	sealed , err := models.SealRemoteKey( fixture.cfg.SecretKey , "portalkey.secretsecretsecret" )
	if err != nil { t.Fatal( err ) }
	project , err := models.CreateProject( fixture.store , models.ProjectFields{ Name: "Board" , BaseURL: app.server.URL , Listed: true } , sealed )
	if err != nil { t.Fatal( err ) }
	owner := fixture.signIn( t , fixture.owner.ID )
	grant := "/api/admin/projects/" + idText( project.ID ) + "/grant/" + idText( fixture.alice.ID )
	if status , reply := send( t , owner , grant , fiber.Map{} ); status != fiber.StatusOK { t.Fatalf( "grant: %d %v" , status , reply ) }
	stored , err := models.GetProject( fixture.store , project.ID )
	if err != nil { t.Fatal( err ) }
	if stored.Allows( fixture.alice.ID ) == false { t.Fatal( "granted, yet not allowed" ) }
}
