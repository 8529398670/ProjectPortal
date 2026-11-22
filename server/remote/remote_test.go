package remote

import (
	context "context"
	errors "errors"
	io "io"
	http "net/http"
	httptest "net/http/httptest"
	strings "strings"
	testing "testing"
)

// fakeApp is a catalog app that answers path with status and body.
func fakeApp( t *testing.T , routes map[string]func( w http.ResponseWriter , r *http.Request ) ) ( base_url string ) {
	t.Helper()
	server := httptest.NewServer( http.HandlerFunc( func( w http.ResponseWriter , r *http.Request ) {
		handler , ok := routes[ r.Method + " " + r.URL.Path ]
		if ok == false {
			w.WriteHeader( http.StatusNotFound )
			_ , _ = io.WriteString( w , `{"error":"not found","code":"not_found"}` )
			return
		}
		handler( w , r )
	} ) )
	t.Cleanup( server.Close )
	base_url = server.URL
	return
}

func answer( status int , body string ) func( http.ResponseWriter , *http.Request ) {
	return func( w http.ResponseWriter , _ *http.Request ) {
		w.Header().Set( "Content-Type" , "application/json" )
		w.WriteHeader( status )
		_ , _ = io.WriteString( w , body )
	}
}

func TestCheckKeyReadsContractV3( t *testing.T ) {
	base := fakeApp( t , map[string]func( http.ResponseWriter , *http.Request ){
		"GET /api/me": answer( 200 , `{"id":7,"display_name":"ProjectPortal","role":"admin","is_owner":false,
			"contract":3,"capabilities":["key_grants","key_rotation"],"owner":false,"kind":"service",
			"skill":"go-webserver","skill_commit":"917bdb0",
			"key":{"id":"k1","name":"ProjectPortal","role":"admin","may_mint_keys":true,"may_sign_in":false,
			"expires_at":"2027-11-22T00:00:00Z"}}` ),
	} )
	me , err := CheckKey( context.Background() , base , "secret" )
	if err != nil { t.Fatal( err ) }
	if me.Kind != KindService || me.Skill != "go-webserver" || me.SkillCommit != "917bdb0" || me.Contract != 3 {
		t.Fatalf( "v3 fields not read: %+v" , me )
	}
	if me.Key == nil || me.Key.ID != "k1" || me.Key.MayMintKeys == false || me.Key.ExpiresAt == nil {
		t.Fatalf( "key grant not read: %+v" , me.Key )
	}
}

func TestCheckKeyOnV2AppHasNoGrant( t *testing.T ) {
	base := fakeApp( t , map[string]func( http.ResponseWriter , *http.Request ){
		"GET /api/me": answer( 200 , `{"id":2,"role":"admin","contract":2,"capabilities":[]}` ),
	} )
	me , err := CheckKey( context.Background() , base , "secret" )
	if err != nil { t.Fatal( err ) }
	if me.Key != nil || me.Kind != "" { t.Fatalf( "v2 app grew v3 fields: %+v" , me ) }
}

func TestRefusalBranchesOnCode( t *testing.T ) {
	cases := []struct {
		body string
		code string
		want string
	}{
		{ `{"error":"only the owner","code":"owner_required"}` , CodeOwnerRequired , "only that project's owner" },
		{ `{"error":"nope","code":"owner_protected"}` , CodeOwnerProtected , "nobody may do this" },
		{ `{"error":"key_minting_off","code":"key_minting_off"}` , CodeKeyMintingOff , "Connect ProjectPortal again" },
		{ `{"error":"bad name","code":"invalid","field":"name"}` , CodeInvalid , "bad name (name)" },
		{ `{"error":"slow","code":"rate_limited"}` , CodeRateLimited , "slow down" },
		{ `{"error":"service","code":"service_account"}` , CodeServiceAccount , "service account" },
		{ `{"error":"something new","code":"brand_new"}` , "brand_new" , "the project said: something new" },
	}
	for _ , each := range cases {
		base := fakeApp( t , map[string]func( http.ResponseWriter , *http.Request ){
			"POST /api/admin/users/3/role": answer( 403 , each.body ),
		} )
		err := SetRole( context.Background() , base , "k" , 3 , "user" )
		if CodeOf( err ) != each.code { t.Errorf( "%s: code %q" , each.body , CodeOf( err ) ) }
		if err == nil || strings.Contains( err.Error() , each.want ) == false {
			t.Errorf( "%s: message %v, want %q" , each.body , err , each.want )
		}
	}
}

func TestRefusalFallsBackToWordsOnV2( t *testing.T ) {
	base := fakeApp( t , map[string]func( http.ResponseWriter , *http.Request ){
		"POST /api/admin/users/3/role": answer( 403 , `{"error":"forbidden"}` ),
	} )
	err := SetRole( context.Background() , base , "k" , 3 , "user" )
	if CodeOf( err ) != "" || strings.Contains( err.Error() , "lets only its owner" ) == false {
		t.Fatalf( "v2 forbidden not worded: %v" , err )
	}
}

func TestNotAuthenticatedIsUnauthorized( t *testing.T ) {
	base := fakeApp( t , map[string]func( http.ResponseWriter , *http.Request ){
		"GET /api/me": answer( 401 , `{"error":"sign in","code":"not_authenticated"}` ),
	} )
	_ , err := CheckKey( context.Background() , base , "old" )
	var remote_err *Error
	if errors.As( err , &remote_err ) == false || remote_err.Unauthorized() == false { t.Fatalf( "not unauthorized: %v" , err ) }
}

func TestRotateSelf( t *testing.T ) {
	var sent string
	base := fakeApp( t , map[string]func( http.ResponseWriter , *http.Request ){
		"POST /api/keys/self/rotate": func( w http.ResponseWriter , r *http.Request ) {
			sent = r.Header.Get( "Authorization" )
			answer( 200 , `{"credential":"new-secret"}` )( w , r )
		},
	} )
	credential , err := RotateSelf( context.Background() , base , "old-secret" )
	if err != nil || credential != "new-secret" { t.Fatalf( "rotate: %q %v" , credential , err ) }
	if sent != "Bearer old-secret" { t.Fatalf( "sent %q" , sent ) }
}

func TestRotateSelfWithoutCredentialFails( t *testing.T ) {
	base := fakeApp( t , map[string]func( http.ResponseWriter , *http.Request ){
		"POST /api/keys/self/rotate": answer( 200 , `{}` ),
	} )
	if _ , err := RotateSelf( context.Background() , base , "k" ); err == nil { t.Fatal( "empty rotate answer taken" ) }
}

func TestUserSessions( t *testing.T ) {
	base := fakeApp( t , map[string]func( http.ResponseWriter , *http.Request ){
		"GET /api/admin/users/5/sessions": answer( 200 , `{"sessions":[{"id":"s1","created_at":"2026-11-01T10:00:00Z",
			"last_seen_at":"2026-11-20T10:00:00Z","ip":"10.0.0.2","user_agent":"Safari","label":"phone"}]}` ),
	} )
	sessions , err := UserSessions( context.Background() , base , "k" , 5 )
	if err != nil || len( sessions ) != 1 || sessions[ 0 ].IP != "10.0.0.2" || sessions[ 0 ].LastSeenAt == nil {
		t.Fatalf( "sessions: %+v %v" , sessions , err )
	}
}
