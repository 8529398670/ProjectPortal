package routes

import (
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	apidoc "projectportal/server/apidoc"
	logs "projectportal/server/logs"
	models "projectportal/server/models"
	portal "projectportal/server/portal"
	remote "projectportal/server/remote"
	security "projectportal/server/security"
)

// Manage project: running a catalog app's accounts, invite links and API keys
// from here, through the admin key the portal holds there (portal contract
// v2, CROSS_PROJECT.md). Each action that needs a capability checks the app
// has it first, and says which upgrade it is missing when it does not.
//
// Minting an API key on another app is browser-only, like every route that
// stores or makes a key here: a leaked portal key must not be able to leave
// keys behind on other apps. For the same reason a key may not make or
// promote an admin account there (adminRoleRefusal), nor ask for a login link
// for any account there (remoteLoginLink): either is a way into every app at
// once that outlives revoking the key here. Invite links, and accounts below
// admin, follow the template's rule: an admin key may make them.

const (
	remoteRoleAdmin  = "admin"
	remoteRoleUser   = "user"
	remoteRoleViewer = "viewer"
)

// capabilityNames words each capability for an admin reading why something
// is not offered.
var capabilityNames = map[string]string{
	remote.CapViewerRole:    "view-only accounts",
	remote.CapAccountExpiry: "accounts that expire",
	remote.CapKeyMinting:    "API keys made from the portal",
	remote.CapInviteLinks:   "invite links",
	remote.CapKeySignIn:     "signing in with your own key",
	remote.CapKeyRotation:   "key rotation",
}

// managedProject is a project an admin is about to act on, with its admin
// key opened.
func ( handlers *Handlers ) managedProject( project_id uint64 ) ( project *models.Project , key string , refused error ) {
	project , get_err := models.GetProject( handlers.Store , project_id )
	if get_err != nil {
		refused = projectFieldsRefusal( get_err )
		return
	}
	key , refused = handlers.adminKeyOf( project )
	return
}

// needs refuses unless the project has capability. What was recorded may be
// from before the app was upgraded, so a miss asks the app again first.
func ( handlers *Handlers ) needs( project *models.Project , capability string ) ( refused error ) {
	if remote.Has( project.Capabilities , capability ) { return }
	if checked , check_err := portal.Check( handlers.Store , handlers.cfg().SecretKey , project ); check_err == nil {
		*project = *checked
		if remote.Has( project.Capabilities , capability ) { return }
	}
	refused = refuse( fiber.StatusConflict , project.Name + " does not support " + capabilityNames[ capability ] +
		" yet: run the webserver-upgrade skill in it (portal contract v2)" )
	return
}

// needsMinting is needs( key_minting ), and on an app that says what its
// keys may do (key_grants), that the portal's key may make keys.
func ( handlers *Handlers ) needsMinting( project *models.Project ) ( refused error ) {
	if refused = handlers.needs( project , remote.CapKeyMinting ); refused != nil { return }
	if mintingBlocked( project ) {
		refused = refuse( fiber.StatusConflict , project.Name + ": the portal's key there may not make keys. In that app, Connect ProjectPortal again from the owner's API keys screen, and put the new key on the project here" )
	}
	return
}

func mintingBlocked( project *models.Project ) ( blocked bool ) {
	blocked = remote.Has( project.Capabilities , remote.CapKeyGrants ) && project.KeyFacts.Grants && project.KeyFacts.MayMintKeys == false
	return
}

// roleAllowed checks a role for an account or key there; viewer only on an
// app that has it.
func ( handlers *Handlers ) roleAllowed( project *models.Project , role string ) ( refused error ) {
	switch role {
	case remoteRoleAdmin , remoteRoleUser:
	case remoteRoleViewer:
		refused = handlers.needs( project , remote.CapViewerRole )
	default:
		refused = refuse( fiber.StatusBadRequest , "role must be admin, user or viewer" )
	}
	return
}

// adminRoleRefusal keeps making or promoting an admin account on a project
// to a browser session. An admin there can mint keys and sign in as anyone
// below them, so a leaked key here would leave that behind on every app.
func adminRoleRefusal( c fiber.Ctx , role string ) ( refused error ) {
	if role != remoteRoleAdmin { return }
	if actor := security.ActorFrom( c ); actor == nil || actor.ViaAPIKey() == false { return }
	refused = refuse( fiber.StatusForbidden , "only a browser session makes an admin account on a project, or makes someone one" ).coded( apidoc.CodeSessionOnly )
	return
}

// ladderOn is the owner rule here (models.MayActOn) for an action on what
// someone has on the projects: their account there, their link to it, their
// own key. Nobody does it to the owner, and only the owner to another admin,
// the same as to their account here. what follows "cannot".
func ( handlers *Handlers ) ladderOn( c fiber.Ctx , user_id uint64 , what string ) ( refused error ) {
	if found := actOnRefusal( models.MayActOn( handlers.Store , security.UserFrom( c ).ID , user_id ) , what ); found != nil { refused = found }
	return
}

// expiryFrom is days from now as a time, nil for 0 (never).
func expiryFrom( days int ) ( expires_at *time.Time ) {
	if days <= 0 { return }
	at := time.Now().UTC().Add( time.Duration( days ) * 24 * time.Hour ).Truncate( time.Second )
	expires_at = &at
	return
}

type remoteUserPath struct {
	projectIDPath
	RemoteUserID uint64 `json:"remote_user_id" path:"remote_user_id" jsonschema:"minimum=1" jsonschema_description:"The account's id on the project, from admin_list_remote_users." error:"invalid remote user id"`
}

// --- accounts ---------------------------------------------------------------

type createRemoteUserRequest struct {
	projectIDPath
	DisplayName   string `json:"display_name" jsonschema:"minLength=1,maxLength=80" error:"display_name must be 1-80 characters"`
	Role          string `json:"role" jsonschema:"enum=user,enum=viewer,enum=admin" jsonschema_description:"viewer needs an app with view-only accounts." error:"role must be admin, user or viewer"`
	ExpiresInDays int    `json:"expires_in_days,omitempty" jsonschema:"minimum=0,maximum=3650" jsonschema_description:"The account stops working this many days from now. 0 or left out: never. Needs an app with accounts that expire." error:"expires_in_days must be 0-3650"`
	UserID        uint64 `json:"user_id,omitempty" jsonschema:"minimum=0" jsonschema_description:"Someone here to put on the project with it, so their Open signs into it. Left out: a guest with no account here." error:"invalid user id"`
}

type createdRemoteUserReply struct {
	RemoteUserID uint64 `json:"remote_user_id"`
	LoginURL     string `json:"login_url" jsonschema_description:"Signs the new account in. It works once, for 15 minutes. A credential: treat it like a password."`
}

// createRemoteUser makes an account on the project directly: a guest, a
// view-only account, or one for someone here.
//
// @route("POST /api/admin/projects/{project_id}/remote-users", summary="Make an account on a project",
//     tag="Manage project", role=ADMIN, status=201, errors=(404, 409, 424),
//     description="Make an account on the project itself, with a role and an optional expiry, and get "
//         "its login link. With user_id, that person here is put on the project with it. An admin account "
//         "takes a browser session: an API key gets 403 session_only.",
//     mcp="admin_create_remote_user",
//     mcp_note="Answers login_url: hand it to the person it is for (it works once, for 15 minutes).",
// )
func ( handlers *Handlers ) createRemoteUser( c fiber.Ctx , in *createRemoteUserRequest ) ( reply *createdRemoteUserReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	role := strings.TrimSpace( in.Role )
	if role == "" { role = remoteRoleUser }
	if refused = adminRoleRefusal( c , role ); refused != nil { return nil , refused }
	if refused = handlers.roleAllowed( project , role ); refused != nil { return nil , refused }
	expires_at := expiryFrom( in.ExpiresInDays )
	if expires_at != nil {
		if refused = handlers.needs( project , remote.CapAccountExpiry ); refused != nil { return nil , refused }
	}
	var member *models.User
	if in.UserID != 0 {
		user , user_err := models.GetUser( handlers.Store , in.UserID )
		if user_err != nil {
			err = refuse( fiber.StatusNotFound , "no such user" )
			return
		}
		if existing , get_err := models.GetMembership( handlers.Store , user.ID , project.ID ); get_err == nil && existing.Active() {
			err = refuse( fiber.StatusConflict , user.DisplayName + " is already on this project" )
			return
		}
		member = user
	}
	ctx , cancel := remoteContext()
	defer cancel()
	name := strings.TrimSpace( in.DisplayName )
	link , remote_err := remote.CreateUserAs( ctx , project.BaseURL , key , remote.NewAccount{ Name: name , Role: role , ExpiresAt: expires_at } )
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
	if member != nil {
		membership := models.Membership{
			UserID: member.ID , ProjectID: project.ID , RemoteUserID: link.RemoteUserID , RemoteName: name ,
			RemoteRole: role , RemoteExpiresAt: expires_at , Via: models.ViaJoined ,
		}
		if _ , err = models.AddMembership( handlers.Store , membership , true ); err != nil { return }
		if err = handlers.allowMember( c , project.ID , member.ID ); err != nil { return }
	}
	handlers.audit( c , logs.KindAdmin , "project account made" , fiber.Map{
		"project_id": project.ID , "remote_user_id": link.RemoteUserID , "role": role , "expires_in_days": in.ExpiresInDays , "user_id": in.UserID ,
	} )
	reply = &createdRemoteUserReply{ RemoteUserID: link.RemoteUserID , LoginURL: link.URL }
	return
}

// remoteLoginLink is a fresh login link for any account the portal's key may
// make one for, to hand over by hand. Browser only: through a key it would be
// signing in as anyone below admin on every app, from wherever the key went.
//
// @route("POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/login-link", summary="A login link for an account on a project",
//     tag="Manage project", role=ADMIN, access=SESSION_ONLY, errors=(404, 409, 424),
//     description="A fresh login link for one account on the project. It works once, for 15 minutes. "
//         "Browser only: an API key gets 403 session_only.",
// )
func ( handlers *Handlers ) remoteLoginLink( c fiber.Ctx , in *remoteUserPath ) ( reply *projectLinkReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	link , remote_err := remote.ReissueLogin( ctx , project.BaseURL , key , in.RemoteUserID )
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
	handlers.audit( c , logs.KindAdmin , "project login link made" , fiber.Map{ "project_id": project.ID , "remote_user_id": in.RemoteUserID } )
	reply = &projectLinkReply{ ProjectID: project.ID , LoginURL: link.URL }
	return
}

type remoteRoleRequest struct {
	remoteUserPath
	Role string `json:"role" jsonschema:"enum=user,enum=viewer,enum=admin" error:"role must be admin, user or viewer"`
}

// setRemoteRole changes an account's role on the project.
//
// @route("POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/role", summary="Change an account's role on a project",
//     tag="Manage project", role=ADMIN, errors=(400, 404, 409, 424),
//     description="admin, user, or viewer (view only, on an app that has it). The project's owner cannot be changed. "
//         "Making someone admin takes a browser session: an API key gets 403 session_only.",
//     mcp="admin_set_remote_role",
// )
func ( handlers *Handlers ) setRemoteRole( c fiber.Ctx , in *remoteRoleRequest ) ( reply *okReply , err error ) {
	if refused := adminRoleRefusal( c , in.Role ); refused != nil { return nil , refused }
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	if refused = handlers.roleAllowed( project , in.Role ); refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	if remote_err := remote.SetRole( ctx , project.BaseURL , key , in.RemoteUserID , in.Role ); remote_err != nil {
		return nil , handlers.remoteRefusal( project , remote_err , true )
	}
	err = models.SyncRemoteAccount( handlers.Store , project.ID , in.RemoteUserID , func( membership *models.Membership ) { membership.RemoteRole = in.Role } )
	if err != nil { return }
	handlers.audit( c , logs.KindAdmin , "project account role changed" , fiber.Map{ "project_id": project.ID , "remote_user_id": in.RemoteUserID , "role": in.Role } )
	reply = &okReply{ Ok: true }
	return
}

type remoteExpiryRequest struct {
	remoteUserPath
	ExpiresInDays int `json:"expires_in_days" jsonschema:"minimum=0,maximum=3650" jsonschema_description:"Stops working this many days from now. 0: never." error:"expires_in_days must be 0-3650"`
}

type remoteExpiryReply struct {
	Ok        bool       `json:"ok"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// setRemoteExpiry is when an account on the project stops working.
//
// @route("POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/expiry", summary="Set when an account on a project expires",
//     tag="Manage project", role=ADMIN, errors=(404, 409, 424),
//     description="The account stops working expires_in_days from now; 0 means never. Needs an app with "
//         "accounts that expire.",
//     mcp="admin_set_remote_expiry",
// )
func ( handlers *Handlers ) setRemoteExpiry( c fiber.Ctx , in *remoteExpiryRequest ) ( reply *remoteExpiryReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	if refused = handlers.needs( project , remote.CapAccountExpiry ); refused != nil { return nil , refused }
	expires_at := expiryFrom( in.ExpiresInDays )
	ctx , cancel := remoteContext()
	defer cancel()
	if remote_err := remote.SetExpiry( ctx , project.BaseURL , key , in.RemoteUserID , expires_at ); remote_err != nil {
		return nil , handlers.remoteRefusal( project , remote_err , true )
	}
	err = models.SyncRemoteAccount( handlers.Store , project.ID , in.RemoteUserID , func( membership *models.Membership ) { membership.RemoteExpiresAt = expires_at } )
	if err != nil { return }
	handlers.audit( c , logs.KindAdmin , "project account expiry set" , fiber.Map{ "project_id": project.ID , "remote_user_id": in.RemoteUserID , "expires_in_days": in.ExpiresInDays } )
	reply = &remoteExpiryReply{ Ok: true , ExpiresAt: expires_at }
	return
}

type remoteDisabledRequest struct {
	remoteUserPath
	Disabled bool `json:"disabled"`
}

// setRemoteDisabled switches an account on the project off, or back on.
//
// @route("POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/disabled", summary="Disable or enable an account on a project",
//     tag="Manage project", role=ADMIN, errors=(404, 409, 424),
//     description="disabled true stops the account signing in there, and ends its sessions.",
//     mcp="admin_set_remote_disabled",
// )
func ( handlers *Handlers ) setRemoteDisabled( c fiber.Ctx , in *remoteDisabledRequest ) ( reply *okReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	if remote_err := remote.SetDisabled( ctx , project.BaseURL , key , in.RemoteUserID , in.Disabled ); remote_err != nil {
		return nil , handlers.remoteRefusal( project , remote_err , true )
	}
	handlers.audit( c , logs.KindAdmin , "project account disabled changed" , fiber.Map{ "project_id": project.ID , "remote_user_id": in.RemoteUserID , "disabled": in.Disabled } )
	reply = &okReply{ Ok: true }
	return
}

// revokeRemoteAccess ends every way an account on the project is signed in.
//
// @route("POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/revoke-access", summary="Sign an account out everywhere on a project",
//     tag="Manage project", role=ADMIN, errors=(404, 409, 424),
//     description="End every session, login link and API key the account has there. The account stays "
//         "enabled; a new login link signs it back in.",
//     mcp="admin_revoke_remote_access", destructive=True,
// )
func ( handlers *Handlers ) revokeRemoteAccess( c fiber.Ctx , in *remoteUserPath ) ( reply *okReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	if remote_err := remote.RevokeAccess( ctx , project.BaseURL , key , in.RemoteUserID ); remote_err != nil {
		return nil , handlers.remoteRefusal( project , remote_err , true )
	}
	handlers.audit( c , logs.KindAdmin , "project account access revoked" , fiber.Map{ "project_id": project.ID , "remote_user_id": in.RemoteUserID } )
	reply = &okReply{ Ok: true }
	return
}

// deleteRemoteUser removes an account from the project for good, and
// forgets it here.
//
// @route("POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/delete", summary="Delete an account on a project",
//     tag="Manage project", role=ADMIN, errors=(404, 409, 424),
//     description="Delete the account on the project itself, with everything it held there. Anyone here it "
//         "belonged to comes off the project. It cannot be undone.",
//     mcp="admin_delete_remote_user", destructive=True,
// )
func ( handlers *Handlers ) deleteRemoteUser( c fiber.Ctx , in *remoteUserPath ) ( reply *okReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	if remote_err := remote.DeleteUser( ctx , project.BaseURL , key , in.RemoteUserID ); remote_err != nil {
		return nil , handlers.remoteRefusal( project , remote_err , true )
	}
	if _ , err = models.ForgetRemoteAccount( handlers.Store , project.ID , in.RemoteUserID ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "project account deleted" , fiber.Map{ "project_id": project.ID , "remote_user_id": in.RemoteUserID } )
	reply = &okReply{ Ok: true }
	return
}

// --- invite links -------------------------------------------------------------

type remoteInviteLinksReply struct {
	InviteLinks []remote.InviteLink `json:"invite_links"`
}

// remoteInviteLinks is the project's invite links.
//
// @route("GET /api/admin/projects/{project_id}/invite-links", summary="A project's invite links",
//     tag="Manage project", role=ADMIN, errors=(404, 409, 424),
//     description="The invite links on the project itself. Needs an app with invite links switched on.",
//     mcp="admin_list_remote_invite_links", read_only=True,
// )
func ( handlers *Handlers ) remoteInviteLinks( c fiber.Ctx , in *projectIDPath ) ( reply *remoteInviteLinksReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	if refused = handlers.needs( project , remote.CapInviteLinks ); refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	links , remote_err := remote.ListInviteLinks( ctx , project.BaseURL , key )
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
	reply = &remoteInviteLinksReply{ InviteLinks: links }
	return
}

type createRemoteInviteRequest struct {
	projectIDPath
	Label   string `json:"label" jsonschema:"minLength=1,maxLength=80" error:"label must be 1-80 characters, uses 1-1000, days 1-365"`
	MaxUses int    `json:"max_uses" jsonschema:"minimum=1,maximum=1000" error:"label must be 1-80 characters, uses 1-1000, days 1-365"`
	Days    int    `json:"days" jsonschema:"minimum=1,maximum=365" error:"label must be 1-80 characters, uses 1-1000, days 1-365"`
	Role    string `json:"role,omitempty" jsonschema:"enum=user,enum=viewer" jsonschema_description:"What people who join through it become. Default user." error:"role must be user or viewer"`
}

type createdRemoteInviteReply struct {
	JoinURL string `json:"join_url" jsonschema:"format=uri" jsonschema_description:"The link to hand out, shown in this response only. A credential: treat it like a password."`
}

// createRemoteInvite makes an invite link on the project.
//
// @route("POST /api/admin/projects/{project_id}/invite-links", summary="Make an invite link on a project",
//     tag="Manage project", role=ADMIN, status=201, errors=(400, 404, 409, 424),
//     description="One link a group can join the project itself through: up to max_uses accounts, for "
//         "days days. role viewer makes them view only, on an app that has it.",
//     mcp="admin_create_remote_invite_link",
//     mcp_note="Answers join_url: hand it out; it is shown only once.",
// )
func ( handlers *Handlers ) createRemoteInvite( c fiber.Ctx , in *createRemoteInviteRequest ) ( reply *createdRemoteInviteReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	if refused = handlers.needs( project , remote.CapInviteLinks ); refused != nil { return nil , refused }
	role := in.Role
	if role == "" { role = remoteRoleUser }
	if role == remoteRoleViewer {
		if refused = handlers.needs( project , remote.CapViewerRole ); refused != nil { return nil , refused }
	}
	ctx , cancel := remoteContext()
	defer cancel()
	url , remote_err := remote.CreateInviteLink( ctx , project.BaseURL , key , remote.NewInviteLink{
		Label: strings.TrimSpace( in.Label ) , MaxUses: in.MaxUses , Days: in.Days , Role: role ,
	} )
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
	handlers.audit( c , logs.KindAdmin , "project invite link made" , fiber.Map{ "project_id": project.ID , "max_uses": in.MaxUses , "days": in.Days , "role": role } )
	reply = &createdRemoteInviteReply{ JoinURL: url }
	return
}

type remoteInvitePath struct {
	projectIDPath
	LinkID string `json:"link_id" path:"link_id" jsonschema:"minLength=1,maxLength=100" jsonschema_description:"The link's id, from admin_list_remote_invite_links." error:"invalid link id"`
}

// revokeRemoteInvite stops an invite link on the project working.
//
// @route("POST /api/admin/projects/{project_id}/invite-links/{link_id}/revoke", summary="Stop an invite link on a project",
//     tag="Manage project", role=ADMIN, errors=(404, 409, 424),
//     description="The link stops working. Accounts already made through it stay.",
//     mcp="admin_revoke_remote_invite_link", destructive=True,
// )
func ( handlers *Handlers ) revokeRemoteInvite( c fiber.Ctx , in *remoteInvitePath ) ( reply *okReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	if remote_err := remote.RevokeInviteLink( ctx , project.BaseURL , key , in.LinkID ); remote_err != nil {
		return nil , handlers.remoteRefusal( project , remote_err , true )
	}
	handlers.audit( c , logs.KindAdmin , "project invite link revoked" , fiber.Map{ "project_id": project.ID , "link_id": in.LinkID } )
	reply = &okReply{ Ok: true }
	return
}

// --- API keys -----------------------------------------------------------------

type remoteKeysReply struct {
	Keys []remote.Key `json:"keys"`
}

// remoteKeys is one account's API keys on the project.
//
// @route("GET /api/admin/projects/{project_id}/remote-users/{remote_user_id}/keys", summary="An account's API keys on a project",
//     tag="Manage project", role=ADMIN, errors=(404, 409, 424),
//     description="The account's API keys there, never their secrets. Needs an app where the portal's key "
//         "may make keys.",
//     mcp="admin_list_remote_keys", read_only=True,
// )
func ( handlers *Handlers ) remoteKeys( c fiber.Ctx , in *remoteUserPath ) ( reply *remoteKeysReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	if refused = handlers.needsMinting( project ); refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	keys , remote_err := remote.ListUserKeys( ctx , project.BaseURL , key , in.RemoteUserID )
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
	reply = &remoteKeysReply{ Keys: keys }
	return
}

type mintRemoteKeyRequest struct {
	remoteUserPath
	Name          string `json:"name" jsonschema:"minLength=1,maxLength=80" error:"name must be 1-80 characters"`
	Role          string `json:"role" jsonschema:"enum=user,enum=viewer,enum=admin" jsonschema_description:"No higher than the account's own role." error:"role must be admin, user or viewer"`
	ExpiresInDays int    `json:"expires_in_days" jsonschema:"minimum=1,maximum=365" jsonschema_description:"A key made from the portal always expires." error:"expires_in_days must be 1-365"`
}

// mintRemoteKey makes an API key for an account on the project. It hangs off
// the portal's key there: revoking that revokes this too.
//
// @route("POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/keys", summary="Make an API key on a project",
//     tag="Manage project", role=ADMIN, access=SESSION_ONLY, status=201, errors=(400, 404, 409, 424),
//     description="An API key for the account on the project, shown once. It always expires, its role is "
//         "no higher than the account's, and revoking the portal's admin key there revokes it too.",
// )
func ( handlers *Handlers ) mintRemoteKey( c fiber.Ctx , in *mintRemoteKeyRequest ) ( reply *remote.MintedKey , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	if refused = handlers.needsMinting( project ); refused != nil { return nil , refused }
	if refused = handlers.roleAllowed( project , in.Role ); refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	minted , remote_err := remote.MintKey( ctx , project.BaseURL , key , in.RemoteUserID , remote.NewKey{
		Name: strings.TrimSpace( in.Name ) , Role: in.Role , ExpiresInDays: in.ExpiresInDays ,
	} )
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
	handlers.audit( c , logs.KindAdmin , "project api key made" , fiber.Map{
		"project_id": project.ID , "remote_user_id": in.RemoteUserID , "key_id": minted.ID , "role": in.Role , "expires_in_days": in.ExpiresInDays ,
	} )
	reply = minted
	return
}

type remoteKeyPath struct {
	projectIDPath
	KeyID string `json:"key_id" path:"key_id" jsonschema:"minLength=1,maxLength=100" jsonschema_description:"The key's id, from admin_list_remote_keys." error:"invalid key id"`
}

// revokeRemoteKey stops an API key on the project working, with every key it
// made.
//
// @route("POST /api/admin/projects/{project_id}/keys/{key_id}/revoke", summary="Revoke an API key on a project",
//     tag="Manage project", role=ADMIN, errors=(404, 409, 424),
//     description="The key stops working there now, and so does every key it made.",
//     mcp="admin_revoke_remote_key", destructive=True,
// )
func ( handlers *Handlers ) revokeRemoteKey( c fiber.Ctx , in *remoteKeyPath ) ( reply *okReply , err error ) {
	project , key , refused := handlers.managedProject( in.ProjectID )
	if refused != nil { return nil , refused }
	if refused = handlers.needsMinting( project ); refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	if remote_err := remote.RevokeKey( ctx , project.BaseURL , key , in.KeyID ); remote_err != nil {
		return nil , handlers.remoteRefusal( project , remote_err , true )
	}
	handlers.audit( c , logs.KindAdmin , "project api key revoked" , fiber.Map{ "project_id": project.ID , "key_id": in.KeyID } )
	reply = &okReply{ Ok: true }
	return
}

// --- removal --------------------------------------------------------------------

const (
	onProjectKeep    = "keep"
	onProjectDisable = "disable"
	onProjectRevoke  = "revoke"
)

// actOnRemoteAccount does what an admin chose for someone's account on a
// project as they come off it: keep it, disable it, or end its sessions,
// links and keys. An own-key membership the portal did not make, or one with
// no account recorded, has nothing for the portal's key to act on.
func ( handlers *Handlers ) actOnRemoteAccount( membership *models.Membership , on_project string ) ( err error ) {
	if on_project == "" || on_project == onProjectKeep || membership.RemoteUserID == 0 { return }
	project , key , refused := handlers.managedProject( membership.ProjectID )
	if refused != nil { return refused }
	ctx , cancel := remoteContext()
	defer cancel()
	var remote_err error
	switch on_project {
	case onProjectDisable:
		remote_err = remote.SetDisabled( ctx , project.BaseURL , key , membership.RemoteUserID , true )
	case onProjectRevoke:
		remote_err = remote.RevokeAccess( ctx , project.BaseURL , key , membership.RemoteUserID )
	default:
		return refuse( fiber.StatusBadRequest , "on_project must be keep, disable or revoke" )
	}
	if remote_err != nil { err = handlers.remoteRefusal( project , remote_err , true ) }
	return
}

type projectAccountsRequest struct {
	UserID    uint64 `json:"user_id" path:"user_id" jsonschema:"minimum=1" jsonschema_description:"The account's id here." error:"invalid user id"`
	OnProject string `json:"on_project" jsonschema:"enum=disable,enum=revoke" jsonschema_description:"Disable their account on every project, or end its sessions, links and keys there." error:"on_project must be disable or revoke"`
}

type projectAccountResult struct {
	ProjectID uint64 `json:"project_id"`
	Name      string `json:"name"`
	Ok        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
}

type projectAccountsReply struct {
	Results []projectAccountResult `json:"results" jsonschema_description:"One per project they are on that the portal made or linked an account for."`
}

// actOnProjectAccounts is coming off every project at once, for someone
// about to be disabled or deleted here: each project they are on is asked to
// disable their account, or sign it out everywhere. It carries on past a
// project that refuses, and says which.
//
// @route("POST /api/admin/users/{user_id}/project-accounts", summary="Disable someone's accounts on every project",
//     tag="Manage project", role=ADMIN, errors=(404,),
//     description="For every project they are on, disable their account there (on_project disable) or end "
//         "its sessions, links and keys (revoke). Use it before disabling or deleting them here. Answers "
//         "how each project took it. Only the owner does it to an admin; nobody does it to the owner.",
//     mcp="admin_act_on_project_accounts", destructive=True,
// )
func ( handlers *Handlers ) actOnProjectAccounts( c fiber.Ctx , in *projectAccountsRequest ) ( reply *projectAccountsReply , err error ) {
	if _ , get_err := models.GetUser( handlers.Store , in.UserID ); get_err != nil {
		err = refuse( fiber.StatusNotFound , "no such user" )
		return
	}
	if refused := handlers.ladderOn( c , in.UserID , "act on the project accounts of" ); refused != nil { return nil , refused }
	if in.OnProject != onProjectDisable && in.OnProject != onProjectRevoke {
		err = refuse( fiber.StatusBadRequest , "on_project must be disable or revoke" )
		return
	}
	memberships , err := models.MembershipsOf( handlers.Store , in.UserID )
	if err != nil { return }
	reply = &projectAccountsReply{ Results: []projectAccountResult{} }
	for project_id , membership := range memberships {
		if membership.RemoteUserID == 0 { continue }
		result := projectAccountResult{ ProjectID: project_id , Ok: true }
		if project , get_err := models.GetProject( handlers.Store , project_id ); get_err == nil { result.Name = project.Name }
		if act_err := handlers.actOnRemoteAccount( membership , in.OnProject ); act_err != nil {
			result.Ok , result.Error = false , act_err.Error()
		}
		reply.Results = append( reply.Results , result )
	}
	handlers.audit( c , logs.KindAdmin , "project accounts acted on" , fiber.Map{ "user_id": in.UserID , "on_project": in.OnProject , "projects": len( reply.Results ) } )
	return
}
