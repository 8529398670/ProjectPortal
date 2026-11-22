package routes

import (
	testing "testing"

	models "projectportal/server/models"
	portal "projectportal/server/portal"
	remote "projectportal/server/remote"
)

func TestMayActOnFollowsTheLadder( t *testing.T ) {
	connect := &models.Project{ Contract: 3 , KeyFacts: models.KeyFacts{ Kind: "service" } }
	owners := &models.Project{ Contract: 3 , KeyFacts: models.KeyFacts{ Owner: true } }
	old := &models.Project{ Contract: 2 }
	cases := []struct {
		name    string
		project *models.Project
		user    remote.User
		want    bool
	}{
		{ "owner, never" , owners , remote.User{ Owner: true , Role: "admin" } , false },
		{ "ordinary account" , connect , remote.User{ Role: "user" } , true },
		{ "viewer account" , connect , remote.User{ Role: "viewer" } , true },
		{ "another admin, service key" , connect , remote.User{ Role: "admin" } , false },
		{ "a service account, service key" , connect , remote.User{ Role: "admin" , Kind: "service" } , false },
		{ "another admin, owner's key" , owners , remote.User{ Role: "admin" } , true },
		{ "v2 app: let it answer" , old , remote.User{ Role: "admin" } , true },
	}
	for _ , each := range cases {
		if got := portal.MayActOn( each.project , each.user ); got != each.want { t.Errorf( "%s: %v" , each.name , got ) }
	}
}

func TestMintingBlocked( t *testing.T ) {
	grants := []string{ remote.CapKeyMinting , remote.CapKeyGrants }
	if mintingBlocked( &models.Project{ Capabilities: grants , KeyFacts: models.KeyFacts{ Grants: true , MayMintKeys: true } } ) {
		t.Error( "a key that may mint was blocked" )
	}
	if mintingBlocked( &models.Project{ Capabilities: grants , KeyFacts: models.KeyFacts{ Grants: true } } ) == false {
		t.Error( "a key that may not mint was let through" )
	}
	if mintingBlocked( &models.Project{ Capabilities: []string{ remote.CapKeyMinting } } ) {
		t.Error( "a v2 app was blocked without knowing" )
	}
}

func TestBehindTemplate( t *testing.T ) {
	if behindTemplate( "" , "abc" ) || behindTemplate( "abc1234" , "" ) { t.Error( "unknown counted as behind" ) }
	if behindTemplate( "917bdb0" , "917bdb0aa11" ) || behindTemplate( "917BDB0" , "917bdb0" ) { t.Error( "same commit counted as behind" ) }
	if behindTemplate( "917bdb0" , "ea9f39c" ) == false { t.Error( "different commit not behind" ) }
}
