package routes

import (
	errors "errors"
	strings "strings"
	testing "testing"

	models "projectportal/server/models"
	remote "projectportal/server/remote"
)

// A project that turns a key away from its login-link route (Sign-in links
// off, or a build from before keys could) is worded as that, with its own
// words kept; anything else is passed on as it was.
func TestKeyMayNotSignIn( t *testing.T ) {
	project := &models.Project{ Name: "Rook" }
	for _ , status := range []int{ 403 , 404 , 405 } {
		cause := &remote.Error{ Status: status , Message: "the project said: nope" }
		got := keyMayNotSignIn( project , cause )
		var worded *remote.Error
		if errors.As( got , &worded ) == false || worded.Status != status {
			t.Fatalf( "%d: got %v" , status , got )
		}
		if strings.Contains( worded.Message , "Sign-in links" ) == false || strings.Contains( worded.Message , "nope" ) == false {
			t.Fatalf( "%d: message %q" , status , worded.Message )
		}
	}
	for _ , cause := range []error{ &remote.Error{ Status: 401 , Message: "gone" } , &remote.Error{ Message: "could not reach" } , errors.New( "other" ) } {
		if got := keyMayNotSignIn( project , cause ); got != cause {
			t.Fatalf( "changed %v into %v" , cause , got )
		}
	}
}
