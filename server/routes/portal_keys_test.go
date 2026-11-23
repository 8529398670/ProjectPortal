package routes

import (
	json "encoding/json"
	http "net/http"
	httptest "net/http/httptest"
	strings "strings"
	sync "sync"
	testing "testing"

	fiber "github.com/gofiber/fiber/v3"

	models "projectportal/server/models"
)

// What a portal key, or a second admin here, may do to the apps in the
// catalog (security audit of 2026-11-22: H3, H4, M3). The portal holds an
// admin key on every app, so anything one of its own keys can reach there is
// reached on all of them at once.

// fakeApp is a catalog app that answers the calls the portal makes and
// remembers which ones it got.
type fakeApp struct {
	server *httptest.Server
	lock   sync.Mutex
	calls  []string
}

func newFakeApp( t *testing.T ) ( app *fakeApp ) {
	t.Helper()
	app = &fakeApp{}
	app.server = httptest.NewServer( http.HandlerFunc( func( writer http.ResponseWriter , request *http.Request ) {
		app.lock.Lock()
		app.calls = append( app.calls , request.Method + " " + request.URL.Path )
		app.lock.Unlock()
		writer.Header().Set( "Content-Type" , "application/json" )
		var reply any = map[string]any{ "ok": true }
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/api/admin/users":
			reply = map[string]any{ "total": 3 , "users": []map[string]any{
				{ "id": 1 , "display_name": "Their owner" , "role": "admin" , "owner": true , "kind": "person" },
				{ "id": 5 , "display_name": "Alice there" , "role": "user" , "kind": "person" },
				{ "id": 6 , "display_name": "Nobody's" , "role": "user" , "kind": "person" },
			} }
		case request.Method == http.MethodPost && request.URL.Path == "/api/admin/users":
			writer.WriteHeader( http.StatusCreated )
			reply = map[string]any{ "user_id": 9 , "login_path": "/login/new.account" }
		case strings.HasSuffix( request.URL.Path , "/reissue-login" ) || request.URL.Path == "/api/account/login-link":
			reply = map[string]any{ "login_path": "/login/fresh.link" }
		}
		json.NewEncoder( writer ).Encode( reply )
	} ) )
	t.Cleanup( app.server.Close )
	return
}

func ( app *fakeApp ) called( suffix string ) ( result bool ) {
	app.lock.Lock()
	defer app.lock.Unlock()
	for _ , call := range app.calls {
		if strings.HasSuffix( call , suffix ) { result = true }
	}
	return
}

// catalogProject puts app in the catalog with an admin key, and alice on it
// as account 5 there.
func catalogProject( t *testing.T , fixture *authFixture , app *fakeApp ) ( project *models.Project ) {
	t.Helper()
	sealed , err := models.SealRemoteKey( fixture.cfg.SecretKey , "portalkey.secretsecretsecret" )
	if err != nil { t.Fatal( err ) }
	project , err = models.CreateProject( fixture.store , models.ProjectFields{ Name: "Board" , BaseURL: app.server.URL , Listed: true } , sealed )
	if err != nil { t.Fatal( err ) }
	_ , err = models.AddMembership( fixture.store , models.Membership{ UserID: fixture.alice.ID , ProjectID: project.ID , RemoteUserID: 5 , RemoteName: "Alice there" , RemoteRole: "user" , Via: models.ViaLinked } , true )
	if err != nil { t.Fatal( err ) }
	return
}

// H3: a key, even one narrowed to a user, made sign-in links on every app.
func TestAKeyNeedsSignInLinksToOpenOrJoinAProject( t *testing.T ) {
	fixture := newAuthFixture( t )
	app := newFakeApp( t )
	project := catalogProject( t , fixture , app )
	other := catalogProject( t , fixture , newFakeApp( t ) )
	_ = models.RemoveMembership( fixture.store , fixture.alice.ID , other.ID )
	key , stored , err := models.IssueAPIKey( fixture.store , fixture.alice.ID , "script" , models.RoleUser , 0 )
	if err != nil { t.Fatal( err ) }
	open := "/api/projects/" + idText( project.ID ) + "/open"
	join := "/api/projects/" + idText( other.ID ) + "/join"

	if status , reply := fixture.withKey( t , key , http.MethodPost , open , fiber.Map{} ); status != fiber.StatusForbidden { t.Fatalf( "open with a key: %d %v" , status , reply ) }
	if status , reply := fixture.withKey( t , key , http.MethodPost , join , fiber.Map{} ); status != fiber.StatusForbidden { t.Fatalf( "join with a key: %d %v" , status , reply ) }
	if app.called( "/reissue-login" ) { t.Fatal( "the app was asked for a link anyway" ) }

	// Her own session always may; a key may once she lets it.
	alice := fixture.signIn( t , fixture.alice.ID )
	if status , reply := send( t , alice , open , fiber.Map{} ); status != fiber.StatusOK || reply[ "login_url" ] == "" { t.Fatalf( "open from a session: %d %v" , status , reply ) }
	if status , reply := send( t , alice , "/api/keys/" + stored.ID + "/sign-in" , fiber.Map{ "may_sign_in": true } ); status != fiber.StatusOK { t.Fatalf( "%d %v" , status , reply ) }
	if status , reply := fixture.withKey( t , key , http.MethodPost , open , fiber.Map{} ); status != fiber.StatusOK { t.Fatalf( "open with an allowed key: %d %v" , status , reply ) }
}

// H4: an admin key here left admin accounts, and sign-ins as anyone, on
// every app.
func TestAPortalKeyCannotLeaveAdminsOnTheApps( t *testing.T ) {
	fixture := newAuthFixture( t )
	app := newFakeApp( t )
	project := catalogProject( t , fixture , app )
	key , _ , err := models.IssueAPIKey( fixture.store , fixture.owner.ID , "agent" , models.RoleAdmin , 0 )
	if err != nil { t.Fatal( err ) }
	base := "/api/admin/projects/" + idText( project.ID )

	status , reply := fixture.withKey( t , key , http.MethodPost , base + "/remote-users" , fiber.Map{ "display_name": "Evil" , "role": "admin" } )
	if status != fiber.StatusForbidden || reply[ "code" ] != "session_only" { t.Fatalf( "an admin account made with a key: %d %v" , status , reply ) }
	if status , reply = fixture.withKey( t , key , http.MethodPost , base + "/remote-users/6/role" , fiber.Map{ "role": "admin" } ); status != fiber.StatusForbidden || reply[ "code" ] != "session_only" {
		t.Fatalf( "made an admin there with a key: %d %v" , status , reply )
	}
	if status , reply = fixture.withKey( t , key , http.MethodPost , base + "/remote-users/5/login-link" , fiber.Map{} ); status != fiber.StatusForbidden {
		t.Fatalf( "signed in as someone there with a key: %d %v" , status , reply )
	}
	if app.called( "POST /api/admin/users" ) || app.called( "/role" ) || app.called( "/reissue-login" ) { t.Fatal( "the app was asked anyway" ) }

	// Below admin, a key still runs accounts there.
	if status , reply = fixture.withKey( t , key , http.MethodPost , base + "/remote-users" , fiber.Map{ "display_name": "Guest" , "role": "user" } ); status != fiber.StatusCreated { t.Fatalf( "a guest: %d %v" , status , reply ) }
	if status , reply = fixture.withKey( t , key , http.MethodPost , base + "/remote-users/6/role" , fiber.Map{ "role": "user" } ); status != fiber.StatusOK { t.Fatalf( "role user: %d %v" , status , reply ) }

	// And a browser session does all of it.
	owner := fixture.signIn( t , fixture.owner.ID )
	if status , reply = send( t , owner , base + "/remote-users" , fiber.Map{ "display_name": "Helper" , "role": "admin" } ); status != fiber.StatusCreated { t.Fatalf( "session admin account: %d %v" , status , reply ) }
	if status , reply = send( t , owner , base + "/remote-users/5/login-link" , fiber.Map{} ); status != fiber.StatusOK { t.Fatalf( "session login link: %d %v" , status , reply ) }
}

// M3: the owner rule here did not reach the owner's accounts on the apps.
func TestASecondAdminCannotReachTheOwnersProjectAccounts( t *testing.T ) {
	fixture := newAuthFixture( t )
	app := newFakeApp( t )
	project := catalogProject( t , fixture , app )
	if _ , err := models.AddMembership( fixture.store , models.Membership{ UserID: fixture.owner.ID , ProjectID: project.ID , RemoteUserID: 6 , RemoteName: "Nobody's" , RemoteRole: "user" , Via: models.ViaLinked } , true ); err != nil { t.Fatal( err ) }
	second , _ := models.CreateUser( fixture.store , "Second admin" , models.RoleAdmin )
	admin := fixture.signIn( t , second.ID )
	owner_id := idText( fixture.owner.ID )
	base := "/api/admin/projects/" + idText( project.ID )

	for path , body := range map[ string ]fiber.Map{
		"/api/admin/users/" + owner_id + "/project-accounts": { "on_project": "disable" },
		base + "/members/" + owner_id + "/remove":            { "on_project": "revoke" },
		base + "/access/" + owner_id:                         { "allowed": false , "on_project": "disable" },
		base + "/link/" + owner_id:                           { "remote_user_id": 6 },
	} {
		if status , reply := send( t , admin , path , body ); status != fiber.StatusForbidden { t.Fatalf( "%s on the owner: %d %v" , path , status , reply ) }
	}
	if app.called( "/disabled" ) || app.called( "/revoke-access" ) { t.Fatal( "the owner's account there was acted on" ) }

	// An account there that is already someone's is not anyone else's.
	if status , reply := send( t , admin , base + "/link/" + idText( second.ID ) , fiber.Map{ "remote_user_id": 5 } ); status != fiber.StatusConflict {
		t.Fatalf( "linked to alice's account there: %d %v" , status , reply )
	}

	// Below admin, the ladder lets them, and the owner reaches everyone.
	if status , reply := send( t , admin , "/api/admin/users/" + idText( fixture.alice.ID ) + "/project-accounts" , fiber.Map{ "on_project": "disable" } ); status != fiber.StatusOK { t.Fatalf( "alice: %d %v" , status , reply ) }
	owner := fixture.signIn( t , fixture.owner.ID )
	if status , reply := send( t , owner , base + "/members/" + idText( second.ID ) + "/remove" , fiber.Map{} ); status != fiber.StatusNotFound && status != fiber.StatusOK { t.Fatalf( "owner: %d %v" , status , reply ) }
}
