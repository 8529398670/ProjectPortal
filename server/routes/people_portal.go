package routes

import (
	errors "errors"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	db "projectportal/server/db"
	logs "projectportal/server/logs"
	models "projectportal/server/models"
	portal "projectportal/server/portal"
	remote "projectportal/server/remote"
)

// A person's portal account and their app accounts, the other way round from
// models.LinkMemberships. Someone who made a portal account and joined apps
// from it needs nothing here: their memberships link their accounts. For
// people from before the portal (backfill), an admin makes them a portal
// account from the People screen, or links one they already have, and the
// app accounts on that person become that account's memberships, so Open
// signs them into the accounts they already had.

type personPortalUserRequest struct {
	PersonID    uint64 `json:"person_id" path:"person_id" jsonschema:"minimum=1" error:"invalid person id"`
	DisplayName string `json:"display_name,omitempty" jsonschema:"maxLength=80" jsonschema_description:"The account's name here. Left out: the person's name." error:"display name must be 1-80 characters"`
	Role        string `json:"role" jsonschema:"enum=admin,enum=user,enum=viewer" error:"role must be admin, user or viewer"`
	AccountExpiresAt *time.Time `json:"account_expires_at,omitempty" jsonschema_description:"When the account stops working (RFC 3339). Left out: never." error:"account_expires_at must be an RFC 3339 time"`
	models.LoginLinkRequest
}

type personPortalUserReply struct {
	UserID           uint64         `json:"user_id"`
	Person           *models.Person `json:"person"`
	MembershipsAdded int            `json:"memberships_added" jsonschema_description:"How many of the person's app accounts became this account's, so Open signs into them."`
	loginLinkReply
}

// createPersonPortalUser makes a portal account for a person who has none,
// and puts it under them.
//
// @route("POST /api/admin/people/{person_id}/portal-user", summary="Make a person a portal account",
//     tag="People", role=ADMIN, status=201, errors=(400, 404, 409),
//     description="For someone from before the portal: make them an account here, put it under the person, "
//         "and make each of their app accounts its own (as matching would), so Open signs them into the "
//         "accounts they have. Answers with user_id and its first login link. A person who already has an "
//         "account here is 409: link that one instead.",
//     mcp_note="Answers with login_url (open it to sign in; it works once unless max_uses says otherwise).",
//     mcp="admin_create_person_portal_user", call=withLoginLinkURL,
// )
func ( handlers *Handlers ) createPersonPortalUser( c fiber.Ctx , in *personPortalUserRequest ) ( reply *personPortalUserReply , err error ) {
	book , err := models.ReadPeople( handlers.Store )
	if err != nil { return }
	person := book.People[ in.PersonID ]
	if person == nil { return nil , refuse( fiber.StatusNotFound , "no such person" ) }
	for _ , account := range book.AccountsOf( person.ID ) {
		if account.ProjectID == models.PortalProject {
			return nil , refuse( fiber.StatusConflict , person.DisplayName + " already has an account here: link that one" )
		}
	}
	name := in.DisplayName
	if name == "" { name = person.DisplayName }
	if models.ValidDisplayName( name ) == false { return nil , refuse( fiber.StatusBadRequest , "display name must be 1-80 characters" ) }
	options , refused := handlers.loginLinkOptions( c , in.LoginLinkRequest )
	if refused != nil { return nil , refused }
	user , create_err := models.CreateUserExpiring( handlers.Store , name , in.Role , in.AccountExpiresAt )
	if errors.Is( create_err , models.ErrExpiryPast ) { return nil , refuse( fiber.StatusBadRequest , "account_expires_at must be in the future" ) }
	if create_err != nil { return nil , create_err }
	handlers.audit( c , logs.KindAdmin , "user created" , fiber.Map{ "user_id": user.ID , "name": user.DisplayName , "role": user.Role , "expires_at": user.ExpiresAt , "person_id": person.ID } )
	if err = models.RecordPortalUser( handlers.Store , user ); err != nil { return }
	linked , link_err := models.LinkAccounts( handlers.Store , 0 , person.ID , []models.Account{ { ProjectID: models.PortalProject , RemoteUserID: user.ID } } )
	if link_err != nil { return nil , peopleRefusal( link_err ) }
	handlers.audit( c , logs.KindAdmin , "people linked" , fiber.Map{ "person_id": person.ID , "accounts": 1 } )
	added , err := handlers.personMemberships( c , person.ID )
	if err != nil { return }
	credential , info , err := models.IssueLoginLink( handlers.Store , user.ID , options )
	if err != nil { return }
	reply = &personPortalUserReply{ UserID: user.ID , Person: linked , MembershipsAdded: added , loginLinkReply: newLoginLinkReply( credential , info ) }
	return
}

// personMemberships makes each of a person's app accounts a membership of
// their one portal account, where the portal could sign them into it and
// nothing here says otherwise. It never replaces a membership: one that
// points elsewhere is an admin's or the person's own call. A person with no
// portal account, or several, gets none. It answers how many it made.
func ( handlers *Handlers ) personMemberships( c fiber.Ctx , person_id uint64 ) ( added int , err error ) {
	book , err := models.ReadPeople( handlers.Store )
	if err != nil { return }
	accounts := book.AccountsOf( person_id )
	user_id := uint64( 0 )
	for _ , account := range accounts {
		if account.ProjectID != models.PortalProject { continue }
		if user_id != 0 { return } // two portal accounts: whose would they be?
		user_id = account.RemoteUserID
	}
	if user_id == 0 { return }
	user , get_err := models.GetUser( handlers.Store , user_id )
	if get_err != nil || user.KindOrPerson() == models.KindService { return }
	// Where someone's Open lands is theirs: only the owner points another admin's.
	if handlers.ladderOn( c , user_id , "link a project account for" ) != nil { return }
	for _ , account := range accounts {
		if account.ProjectID == models.PortalProject { continue }
		snapshot := book.Snapshots[ account.Key() ]
		if snapshot == nil || snapshot.Service() || snapshot.HandMadePortal() || snapshot.Portal { continue }
		project , project_err := models.GetProject( handlers.Store , account.ProjectID )
		if project_err != nil { continue }
		seen := remote.User{ ID: snapshot.Account.RemoteUserID , Role: snapshot.Role , Kind: snapshot.Kind , Owner: snapshot.Owner }
		// The owner, and an admin unless the key is the owner's: they attach their own key.
		if portal.MayActOn( project , seen ) == false { continue }
		existing , existing_err := models.GetMembership( handlers.Store , user_id , project.ID )
		// Active, or left pointing at another account there: not ours to change.
		if existing_err == nil && ( existing.Active() || existing.RemoteUserID != snapshot.Account.RemoteUserID ) { continue }
		if existing_err != nil && errors.Is( existing_err , db.ErrNotFound ) == false { return added , existing_err }
		owners , owners_err := handlers.remoteOwners( project.ID )
		if owners_err != nil { return added , owners_err }
		if holder := owners[ snapshot.Account.RemoteUserID ]; holder != nil && holder.ID != user_id { continue }
		membership := models.Membership{
			UserID: user_id , ProjectID: project.ID , RemoteUserID: snapshot.Account.RemoteUserID ,
			RemoteName: snapshot.DisplayName , RemoteRole: snapshot.Role , Via: models.ViaLinked ,
		}
		if _ , err = models.AddMembership( handlers.Store , membership , true ); err != nil { return }
		if err = handlers.allowMember( c , project.ID , user_id ); err != nil { return }
		handlers.audit( c , logs.KindAdmin , "project account linked" , fiber.Map{ "project_id": project.ID , "user_id": user_id , "remote_user_id": snapshot.Account.RemoteUserID , "person_id": person_id } )
		added++
	}
	return
}
