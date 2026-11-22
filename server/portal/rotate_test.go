package portal

import (
	io "io"
	http "net/http"
	httptest "net/http/httptest"
	filepath "path/filepath"
	strings "strings"
	sync "sync"
	testing "testing"
	time "time"

	config "projectportal/server/config"
	db "projectportal/server/db"
	models "projectportal/server/models"
	remote "projectportal/server/remote"
)

var testSecret = strings.Repeat( "ab" , 32 )

// rotatingApp is a v3 app that takes one secret at a time and rotates it.
type rotatingApp struct {
	mutex   sync.Mutex
	secret  string
	rotated int
	// lose, when set, takes the rotation but never answers it.
	lose bool
}

func ( app *rotatingApp ) ServeHTTP( w http.ResponseWriter , r *http.Request ) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	w.Header().Set( "Content-Type" , "application/json" )
	if r.URL.Path == "/api/health" {
		_ , _ = io.WriteString( w , `{"ok":true}` )
		return
	}
	if r.Header.Get( "Authorization" ) != "Bearer " + app.secret {
		w.WriteHeader( http.StatusUnauthorized )
		_ , _ = io.WriteString( w , `{"error":"sign in","code":"not_authenticated"}` )
		return
	}
	switch r.Method + " " + r.URL.Path {
	case "GET /api/me":
		_ , _ = io.WriteString( w , `{"id":4,"role":"admin","contract":3,"kind":"service","capabilities":["key_rotation","key_grants"],
			"key":{"id":"k1","name":"ProjectPortal","role":"admin","may_mint_keys":true}}` )
	case "POST /api/keys/self/rotate":
		app.rotated++
		app.secret = "secret-" + string( rune( 'a' + app.rotated ) )
		if app.lose {
			w.WriteHeader( http.StatusBadGateway )
			return
		}
		_ , _ = io.WriteString( w , `{"credential":"` + app.secret + `"}` )
	default:
		w.WriteHeader( http.StatusNotFound )
	}
}

func rotationFixture( t *testing.T , app *rotatingApp ) ( store *db.Store , project *models.Project ) {
	t.Helper()
	server := httptest.NewServer( app )
	t.Cleanup( server.Close )
	store , err := db.Open( &config.Config{ DatabasePath: filepath.Join( t.TempDir() , "test.db" ) } )
	if err != nil { t.Fatal( err ) }
	t.Cleanup( func() { store.Close() } )
	sealed , err := models.SealRemoteKey( testSecret , app.secret )
	if err != nil { t.Fatal( err ) }
	project , err = models.CreateProject( store , models.ProjectFields{ Name: "App" , BaseURL: server.URL , Icon: "kanban" , Listed: true } , sealed )
	if err != nil { t.Fatal( err ) }
	if project , err = Check( store , testSecret , project ); err != nil { t.Fatal( err ) }
	if project.KeyStatus != models.KeyOK || remote.Has( project.Capabilities , remote.CapKeyRotation ) == false {
		t.Fatalf( "fixture project not ready: %+v" , project )
	}
	return
}

func TestRotateStoresTheNewSecret( t *testing.T ) {
	app := &rotatingApp{ secret: "secret-a" }
	store , project := rotationFixture( t , app )
	rotated , err := Rotate( store , testSecret , project )
	if err != nil { t.Fatal( err ) }
	stored , _ := models.OpenRemoteKey( testSecret , rotated.AdminKey )
	if stored != app.secret || app.rotated != 1 { t.Fatalf( "stored %q, app has %q" , stored , app.secret ) }
	if rotated.KeyStatus != models.KeyOK || rotated.KeyRotatedAt == nil { t.Fatalf( "after rotation: %+v" , rotated ) }
}

func TestRotateLostAnswerMarksTheKey( t *testing.T ) {
	app := &rotatingApp{ secret: "secret-a" }
	store , project := rotationFixture( t , app )
	app.lose = true
	if _ , err := Rotate( store , testSecret , project ); err == nil { t.Fatal( "a lost rotation answered no error" ) }
	after , _ := models.GetProject( store , project.ID )
	if after.KeyStatus != models.KeyInvalid || strings.Contains( after.CheckNote , "Connect ProjectPortal again" ) == false {
		t.Fatalf( "lost key not flagged: %q %q" , after.KeyStatus , after.CheckNote )
	}
}

func TestRotateKeepsAKeyChangedMeanwhile( t *testing.T ) {
	app := &rotatingApp{ secret: "secret-a" }
	store , project := rotationFixture( t , app )
	replacement , _ := models.SealRemoteKey( testSecret , "someone-elses" )
	if _ , err := models.SetProjectKey( store , project.ID , replacement ); err != nil { t.Fatal( err ) }
	if _ , err := Rotate( store , testSecret , project ); err == nil { t.Fatal( "rotation overwrote a new key" ) }
	after , _ := models.GetProject( store , project.ID )
	if after.AdminKey != replacement { t.Fatal( "the new key was replaced" ) }
}

func TestRotationDue( t *testing.T ) {
	now := time.Now()
	recent , old := now.Add( -time.Hour ) , now.Add( -40 * 24 * time.Hour )
	ready := func( at *time.Time ) *models.Project {
		return &models.Project{ AdminKey: "s" , KeyStatus: models.KeyOK , Capabilities: []string{ remote.CapKeyRotation } , KeyRotatedAt: at }
	}
	month := 30 * 24 * time.Hour
	if RotationDue( ready( &recent ) , month , now ) { t.Error( "a fresh key was due" ) }
	if RotationDue( ready( &old ) , month , now ) == false { t.Error( "an old key was not due" ) }
	if RotationDue( ready( nil ) , month , now ) == false { t.Error( "a key of unknown age was not due" ) }
	if RotationDue( ready( &old ) , 0 , now ) { t.Error( "rotation off, still due" ) }
	if RotationDue( &models.Project{ AdminKey: "s" , KeyStatus: models.KeyOK , KeyRotatedAt: &old } , month , now ) { t.Error( "an app without rotation was due" ) }
}
