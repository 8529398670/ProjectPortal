package routes

import (
	fiber "github.com/gofiber/fiber/v3"
	testing "testing"

	models "projectportal/server/models"
)

func TestPeopleScreenLinksAndSplits( t *testing.T ) {
	fixture := newAuthFixture( t )
	owner := fixture.signIn( t , fixture.owner.ID )
	// Two apps' worth of accounts, as a sync would have stored them.
	_ = models.ReplaceSnapshots( fixture.store , 0 , []models.AccountSnapshot{
		{ Account: models.Account{ ProjectID: 0 , RemoteUserID: fixture.alice.ID } , DisplayName: "Alice" , Role: "user" , Kind: "person" },
	} )
	_ = models.ReplaceSnapshots( fixture.store , 7 , []models.AccountSnapshot{
		{ Account: models.Account{ ProjectID: 7 , RemoteUserID: 3 } , DisplayName: "alice" , Role: "user" , Kind: "person" },
		{ Account: models.Account{ ProjectID: 7 , RemoteUserID: 4 } , DisplayName: "ProjectPortal" , Role: "admin" , Kind: "service" , Portal: true },
		{ Account: models.Account{ ProjectID: 7 , RemoteUserID: 5 } , DisplayName: "ProjectPortal" , Role: "admin" , Kind: "person" },
	} )
	status , reply := owner.do( t , "GET" , "/api/admin/people" , nil )
	if status != 200 { t.Fatalf( "list: %d %v" , status , reply ) }
	if service := reply[ "service" ].( []any ); len( service ) != 2 { t.Fatalf( "service accounts: %v" , service ) }
	if suggestions := reply[ "suggestions" ].( []any ); len( suggestions ) != 1 { t.Fatalf( "suggestions: %v" , suggestions ) }
	status , reply = send( t , owner , "/api/admin/people/link" , fiber.Map{ "accounts": []fiber.Map{
		{ "project_id": 0 , "remote_user_id": fixture.alice.ID } , { "project_id": 7 , "remote_user_id": 3 } } } )
	if status != 200 { t.Fatalf( "link: %d %v" , status , reply ) }
	status , _ = send( t , owner , "/api/admin/people/link" , fiber.Map{ "accounts": []fiber.Map{ { "project_id": 7 , "remote_user_id": 4 } } } )
	if status != 400 { t.Fatalf( "service account linked: %d" , status ) }
	_ , reply = owner.do( t , "GET" , "/api/admin/people" , nil )
	if people := reply[ "people" ].( []any ); len( people ) != 1 || len( people[ 0 ].( map[string]any )[ "accounts" ].( []any ) ) != 2 {
		t.Fatalf( "people: %v" , people )
	}
	status , _ = send( t , owner , "/api/admin/people/unlink" , fiber.Map{ "project_id": 7 , "remote_user_id": 3 } )
	if status != 200 { t.Fatalf( "unlink: %d" , status ) }
	_ , reply = owner.do( t , "GET" , "/api/admin/people" , nil )
	if suggestions := reply[ "suggestions" ].( []any ); len( suggestions ) != 0 { t.Fatalf( "split pair suggested again: %v" , suggestions ) }
}

func TestPeopleNeedsAnAdmin( t *testing.T ) {
	fixture := newAuthFixture( t )
	alice := fixture.signIn( t , fixture.alice.ID )
	if status , _ := alice.do( t , "GET" , "/api/admin/people" , nil ); status != 403 && status != 404 { t.Fatalf( "a user saw people: %d" , status ) }
}
