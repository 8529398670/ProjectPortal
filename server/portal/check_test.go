package portal

import (
	slices "slices"
	testing "testing"
	time "time"

	models "projectportal/server/models"
	remote "projectportal/server/remote"
)

func v3Project( facts models.KeyFacts , capabilities ...string ) ( project *models.Project ) {
	project = &models.Project{ AdminKey: "sealed" , KeyStatus: models.KeyOK , Contract: 3 , Capabilities: capabilities , KeyFacts: facts }
	return
}

func TestKeyWarnings( t *testing.T ) {
	now := time.Date( 2026 , 11 , 22 , 0 , 0 , 0 , 0 , time.UTC )
	soon , later , gone := now.Add( 10 * 24 * time.Hour ) , now.Add( 200 * 24 * time.Hour ) , now.Add( -time.Hour )
	all := []string{ remote.CapServiceAccounts , remote.CapKeyMinting , remote.CapKeyGrants }
	cases := []struct {
		name    string
		project *models.Project
		want    []string
	}{
		{ "connect key, healthy" , v3Project( models.KeyFacts{ Kind: "service" , Grants: true , MayMintKeys: true , ExpiresAt: &later } , all... ) , []string{} },
		{ "owner's key" , v3Project( models.KeyFacts{ Owner: true , Kind: "person" , Grants: true , MayMintKeys: true } , all... ) , []string{ WarnOwnerKey } },
		{ "hand-made admin" , v3Project( models.KeyFacts{ Kind: "person" , Grants: true , MayMintKeys: true } , all... ) , []string{ WarnNotService } },
		{ "cannot mint" , v3Project( models.KeyFacts{ Kind: "service" , Grants: true } , all... ) , []string{ WarnCannotMint } },
		{ "no minting in app" , v3Project( models.KeyFacts{ Kind: "service" , Grants: true } , remote.CapServiceAccounts ) , []string{} },
		{ "expires soon" , v3Project( models.KeyFacts{ Kind: "service" , Grants: true , MayMintKeys: true , ExpiresAt: &soon } , all... ) , []string{ WarnExpiresSoon } },
		{ "expired" , v3Project( models.KeyFacts{ Kind: "service" , Grants: true , MayMintKeys: true , ExpiresAt: &gone } , all... ) , []string{ WarnExpired } },
		{ "v2 app" , &models.Project{ AdminKey: "sealed" , KeyStatus: models.KeyOK , Contract: 2 } , []string{} },
		{ "broken key" , &models.Project{ AdminKey: "sealed" , KeyStatus: models.KeyInvalid , Contract: 3 , KeyFacts: models.KeyFacts{ Owner: true } } , []string{} },
	}
	for _ , each := range cases {
		if got := KeyWarnings( each.project , now ); slices.Equal( got , each.want ) == false {
			t.Errorf( "%s: %v, want %v" , each.name , got , each.want )
		}
	}
}

func TestFactsOf( t *testing.T ) {
	expires := time.Now().Add( time.Hour )
	facts := FactsOf( &remote.Me{ ID: 9 , IsOwner: false , Kind: "service" , Skill: "go-webserver" , SkillCommit: "abc1234" ,
		Key: &remote.KeyGrant{ ID: "k" , Name: "ProjectPortal" , MayMintKeys: true , ExpiresAt: &expires } } )
	if facts.AccountID != 9 || facts.Grants == false || facts.MayMintKeys == false || facts.ExpiresAt == nil || facts.SkillCommit != "abc1234" {
		t.Fatalf( "facts: %+v" , facts )
	}
	if FactsOf( &remote.Me{ IsOwner: true } ).Owner == false { t.Fatal( "is_owner not read as owner" ) }
	if FactsOf( &remote.Me{} ).Grants { t.Fatal( "grants without a key" ) }
}
