package routes

import (
	errors "errors"
	strings "strings"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	config "projectportal/server/config"
	db "projectportal/server/db"
	logs "projectportal/server/logs"
	models "projectportal/server/models"
	portal "projectportal/server/portal"
	remote "projectportal/server/remote"
	security "projectportal/server/security"
)

// The admin side of the catalog: adding projects with the admin key the
// portal uses on each, checking them, deciding who may use which, and
// matching portal accounts to accounts a project already had.
//
// Routes that store a key are browser-only (SESSION_ONLY), like every route
// here that makes a credential: a leaked portal key must not be able to swap
// in keys, or change an address so that stored keys are sent somewhere else
// (EditProject drops them when the address changes, for the same reason).

// catalogRow is a project as the Catalog screen lists it. Never the key.
type catalogRow struct {
	ID              uint64     `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	BaseURL         string     `json:"base_url"`
	Icon            string     `json:"icon"`
	Color           string     `json:"color"`
	Listed          bool       `json:"listed"`
	GitHubURL       string     `json:"github_url" jsonschema_description:"The project's code, for admins."`
	BoardURL        string     `json:"board_url" jsonschema_description:"Its TodoBoard board, for admins."`
	BoardShareURL   string     `json:"board_share_url" jsonschema_description:"A view-only TodoBoard share link, which everyone who may see the project gets."`
	KeySet          bool       `json:"key_set" jsonschema_description:"Whether the portal holds an admin key for it."`
	KeyStatus       string     `json:"key_status" jsonschema_description:"ok, invalid, not_admin or missing, from the last check."`
	Health          string     `json:"health" jsonschema_description:"up, down, or empty before the first check."`
	CheckNote       string     `json:"check_note"`
	CheckedAt       *time.Time `json:"checked_at"`
	Members         int        `json:"members"`
	ExcludedUserIDs []uint64   `json:"excluded_user_ids" jsonschema_description:"People it is kept from."`
	Position        int        `json:"position"`
	CreatedAt       time.Time  `json:"created_at"`
	Version         string     `json:"version" jsonschema_description:"The app's version, from its last check."`
	Contract        int        `json:"contract" jsonschema_description:"Which portal contract it speaks: 2 lets Manage project do everything its capabilities list; 0 is older."`
	Capabilities    []string   `json:"capabilities" jsonschema_description:"What Manage project may do there: viewer_role, account_expiry, key_minting, invite_links, key_sign_in, and on contract 3 service_accounts, key_grants, error_codes, key_rotation, admin_sessions."`
	Key             catalogKey `json:"key" jsonschema_description:"Who the portal's admin key is there and what it may do, from the last check (contract 3)."`
	Skill           string     `json:"skill" jsonschema_description:"The template the app came from: go-webserver or python-webserver. Empty when it does not say."`
	SkillCommit     string     `json:"skill_commit" jsonschema_description:"The template commit the app was last built from."`
	Behind          bool       `json:"behind" jsonschema_description:"skill_commit differs from the template commit set in Settings."`
	MayRotate       bool       `json:"may_rotate" jsonschema_description:"Rotate now works: a working key on an app with key rotation."`
}

// catalogKey is what a Catalog row says about the portal's key, never the key.
type catalogKey struct {
	AccountID   uint64     `json:"account_id" jsonschema_description:"The account the key belongs to there: the portal's own."`
	Owner       bool       `json:"owner" jsonschema_description:"The key is the app owner's own: it works, but should be a Connect key."`
	Kind        string     `json:"kind" jsonschema_description:"person or service; empty before contract 3."`
	Grants      bool       `json:"grants" jsonschema_description:"The app reported what this key may do."`
	Name        string     `json:"name"`
	MayMintKeys bool       `json:"may_mint_keys"`
	MaySignIn   bool       `json:"may_sign_in"`
	ExpiresAt   *time.Time `json:"expires_at" jsonschema_description:"When the key stops working there; null for never or unknown."`
	RotatedAt   *time.Time `json:"rotated_at" jsonschema_description:"When the portal last stored or rotated it."`
	Warnings    []string   `json:"warnings" jsonschema_description:"owner_key, not_service, cannot_mint, expires_soon, expired."`
}

func ( catalogRow ) APIName() ( name string ) { name = "CatalogProject"; return }

func newCatalogRow( cfg *config.Config , project *models.Project , members int ) ( row catalogRow ) {
	secret_key , template_commit := cfg.SecretKey , cfg.Catalog.TemplateCommit
	row = catalogRow{
		ID: project.ID , Name: project.Name , Description: project.Description , BaseURL: project.BaseURL ,
		Icon: project.Icon , Color: project.Color , Listed: project.Listed , KeySet: project.AdminKey != "" ,
		KeyStatus: project.KeyStatus , Health: project.Health , CheckNote: project.CheckNote ,
		CheckedAt: project.CheckedAt , Members: members , ExcludedUserIDs: project.ExcludedUserIDs ,
		Position: project.Position , CreatedAt: project.CreatedAt ,
		GitHubURL: project.GitHubURL , BoardURL: project.BoardURL , BoardShareURL: openShareLink( secret_key , project ) ,
	}
	if row.ExcludedUserIDs == nil { row.ExcludedUserIDs = []uint64{} }
	row.Version , row.Contract , row.Capabilities = project.Version , project.Contract , project.Capabilities
	if row.Capabilities == nil { row.Capabilities = []string{} }
	facts := project.KeyFacts
	row.Key = catalogKey{
		AccountID: facts.AccountID , Owner: facts.Owner , Kind: facts.Kind , Grants: facts.Grants , Name: facts.KeyName ,
		MayMintKeys: facts.MayMintKeys , MaySignIn: facts.MaySignIn , ExpiresAt: facts.ExpiresAt ,
		RotatedAt: project.KeyRotatedAt , Warnings: portal.KeyWarnings( project , time.Now() ) ,
	}
	row.Skill , row.SkillCommit = facts.Skill , facts.SkillCommit
	row.Behind = behindTemplate( template_commit , facts.SkillCommit )
	row.MayRotate = portal.MayRotate( project )
	return
}

// behindTemplate is whether an app's template commit differs from the one
// set in Settings. Either may be a short sha; empty means unknown, not behind.
func behindTemplate( template_commit string , skill_commit string ) ( behind bool ) {
	template_commit , skill_commit = strings.ToLower( strings.TrimSpace( template_commit ) ) , strings.ToLower( strings.TrimSpace( skill_commit ) )
	if template_commit == "" || skill_commit == "" { return }
	behind = strings.HasPrefix( template_commit , skill_commit ) == false && strings.HasPrefix( skill_commit , template_commit ) == false
	return
}

type listCatalogRequest struct {
	Page
	Q    string `json:"q,omitempty" query:"q" jsonschema_description:"Keep projects whose name or address contains this, in any case."`
	Sort string `json:"sort,omitempty" query:"sort" jsonschema:"enum=name,enum=base_url,enum=health,enum=members,enum=position" jsonschema_description:"Field to sort by. Default the catalog's order."`
	Direction
}

type listCatalogReply struct {
	Projects []catalogRow `json:"projects"`
	Paged
}

// listCatalog is every project, hidden ones included.
//
// @route("GET /api/admin/projects", summary="The catalog",
//     tag="Catalog", role=ADMIN,
//     description="Every project in the portal, a page at a time, with whether its admin key works, "
//         "whether it is up, how many people are on it, and its GitHub and TodoBoard links.",
//     mcp="admin_list_catalog", read_only=True,
// )
func ( handlers *Handlers ) listCatalog( c fiber.Ctx , in *listCatalogRequest ) ( reply *listCatalogReply , err error ) {
	projects , err := models.ListProjects( handlers.Store )
	if err != nil { return }
	counts , err := models.MemberCounts( handlers.Store )
	if err != nil { return }
	query := strings.ToLower( strings.TrimSpace( in.Q ) )
	rows := []catalogRow{}
	for _ , project := range projects {
		if query != "" && strings.Contains( strings.ToLower( project.Name + " " + project.BaseURL ) , query ) == false { continue }
		rows = append( rows , newCatalogRow( handlers.cfg() , project , counts[ project.ID ] ) )
	}
	reply = &listCatalogReply{}
	reply.Projects , reply.Paged = sortedPage( rows , in.Sort , in.Direction , in.Page )
	return
}

type projectFieldsRequest struct {
	Name        string `json:"name" jsonschema:"minLength=1,maxLength=80" error:"name must be 1-80 characters"`
	Description string `json:"description,omitempty" jsonschema:"maxLength=500" error:"description must be at most 500 characters"`
	BaseURL     string `json:"base_url" jsonschema:"minLength=8,maxLength=300" jsonschema_description:"Where the project is served, e.g. https://todo.example.com: what people's browsers open." error:"base_url must be 8-300 characters"`
	Icon        string `json:"icon,omitempty" jsonschema:"maxLength=40,pattern=^[a-z0-9-]*$" jsonschema_description:"A Bootstrap Icons name without bi-, e.g. kanban." error:"icon must be a Bootstrap Icons name"`
	Color       string `json:"color,omitempty" jsonschema:"maxLength=7,pattern=^(#[0-9a-fA-F]{6})?$" jsonschema_description:"#rrggbb." error:"color must be #rrggbb"`
	Listed      bool   `json:"listed" jsonschema_description:"Whether people see it. Off, only admins do."`
	GitHubURL     string `json:"github_url,omitempty" jsonschema:"maxLength=300" jsonschema_description:"Where its code is, e.g. https://github.com/me/app. Only admins see it." error:"github_url must be at most 300 characters"`
	BoardURL      string `json:"board_url,omitempty" jsonschema:"maxLength=300" jsonschema_description:"Its TodoBoard board, e.g. https://todo.example.com/b/app. Only admins see it." error:"board_url must be at most 300 characters"`
	BoardShareURL string `json:"board_share_url,omitempty" jsonschema:"maxLength=500" jsonschema_description:"A view-only TodoBoard share link (/s/...), handed to everyone who may see the project." error:"board_share_url must be at most 500 characters"`
}

func ( fields *projectFieldsRequest ) model() ( result models.ProjectFields ) {
	result = models.ProjectFields{
		Name: fields.Name , Description: fields.Description , BaseURL: fields.BaseURL ,
		Icon: fields.Icon , Color: fields.Color , Listed: fields.Listed ,
		GitHubURL: fields.GitHubURL , BoardURL: fields.BoardURL , BoardShareURL: fields.BoardShareURL ,
	}
	return
}

// sealedFields is what an admin sent, checked, with the share link sealed.
func ( handlers *Handlers ) sealedFields( in *projectFieldsRequest ) ( fields models.ProjectFields , refused error ) {
	fields = in.model()
	if check_err := fields.Check(); check_err != nil {
		refused = projectFieldsRefusal( check_err )
		return
	}
	fields.SealedShare , refused = models.SealRemoteKey( handlers.cfg().SecretKey , fields.BoardShareURL )
	return
}

// openShareLink is a project's view-only board link, or empty.
func openShareLink( secret_key string , project *models.Project ) ( link string ) {
	link , err := models.OpenRemoteKey( secret_key , project.BoardShareURL )
	if err != nil { link = "" }
	return
}

type createProjectRequest struct {
	projectFieldsRequest
	AdminKey string `json:"admin_key,omitempty" jsonschema:"maxLength=400" jsonschema_description:"An admin API key on that project. Checked before it is kept." error:"admin_key must be at most 400 characters"`
}

type catalogReply struct {
	Project catalogRow `json:"project"`
}

func projectFieldsRefusal( cause error ) ( refused error ) {
	switch {
	case errors.Is( cause , models.ErrBadBaseURL ) , errors.Is( cause , models.ErrBadProjectName ) , errors.Is( cause , models.ErrBadLinkURL ):
		refused = refuse( fiber.StatusBadRequest , cause.Error() )
	case errors.Is( cause , db.ErrNotFound ):
		refused = refuse( fiber.StatusNotFound , "no such project" )
	default:
		refused = cause
	}
	return
}

// checkAdminKey asks the project who key is, and refuses unless it is an
// admin there. The key itself never goes into an error.
func ( handlers *Handlers ) checkAdminKey( name string , base_url string , key string ) ( sealed string , refused error ) {
	ctx , cancel := remoteContext()
	defer cancel()
	me , remote_err := remote.CheckKey( ctx , base_url , key )
	if remote_err != nil {
		refused = refuse( fiber.StatusFailedDependency , name + ": " + remote_err.Error() )
		return
	}
	if me.Role != models.RoleAdmin {
		refused = refuse( fiber.StatusBadRequest , name + ": that key is a user key there; the portal needs an admin key" )
		return
	}
	sealed , refused = models.SealRemoteKey( handlers.cfg().SecretKey , key )
	return
}

// createProject adds a project to the catalog.
//
// @route("POST /api/admin/projects", summary="Add a project",
//     tag="Catalog", role=ADMIN, access=SESSION_ONLY, status=201, errors=(424,),
//     description="Add a project to the catalog. With admin_key (an admin API key there), people can "
//         "join it straight away; the key is checked first and never shown again.",
// )
func ( handlers *Handlers ) createProject( c fiber.Ctx , in *createProjectRequest ) ( reply *catalogReply , err error ) {
	fields , refused := handlers.sealedFields( &in.projectFieldsRequest )
	if refused != nil { return nil , refused }
	base_url := fields.BaseURL
	sealed := ""
	if key := strings.TrimSpace( in.AdminKey ); key != "" {
		if sealed , err = handlers.checkAdminKey( fields.Name , base_url , key ); err != nil { return }
	}
	project , create_err := models.CreateProject( handlers.Store , fields , sealed )
	if create_err != nil { return nil , projectFieldsRefusal( create_err ) }
	handlers.audit( c , logs.KindAdmin , "project added" , fiber.Map{ "project_id": project.ID , "name": project.Name , "base_url": project.BaseURL } )
	if checked , check_err := portal.Check( handlers.Store , handlers.cfg().SecretKey , project ); check_err == nil { project = checked }
	reply = &catalogReply{ Project: newCatalogRow( handlers.cfg() , project , 0 ) }
	return
}

type editProjectRequest struct {
	projectIDPath
	projectFieldsRequest
}

type editProjectReply struct {
	Project catalogRow `json:"project"`
	KeysDropped bool   `json:"keys_dropped" jsonschema_description:"The address changed, so the admin key and everyone's own keys for the old one were dropped."`
}

// editProject changes a project's details. A new address drops every key the
// portal held for the old one.
//
// @route("POST /api/admin/projects/{project_id}/edit", summary="Change a project",
//     tag="Catalog", role=ADMIN, access=SESSION_ONLY, errors=(404,),
//     description="Change a project's name, description, address, icon, color, links or whether it is listed. "
//         "A new address drops the admin key and everyone's own keys for the old one: give the key again.",
// )
func ( handlers *Handlers ) editProject( c fiber.Ctx , in *editProjectRequest ) ( reply *editProjectReply , err error ) {
	fields , refused := handlers.sealedFields( &in.projectFieldsRequest )
	if refused != nil { return nil , refused }
	project , moved , edit_err := models.EditProject( handlers.Store , in.ProjectID , fields )
	if edit_err != nil { return nil , projectFieldsRefusal( edit_err ) }
	handlers.audit( c , logs.KindAdmin , "project changed" , fiber.Map{ "project_id": project.ID , "name": project.Name , "keys_dropped": moved } )
	counts , err := models.MemberCounts( handlers.Store )
	if err != nil { return }
	reply = &editProjectReply{ Project: newCatalogRow( handlers.cfg() , project , counts[ project.ID ] ) , KeysDropped: moved }
	return
}

type projectKeyRequest struct {
	projectIDPath
	AdminKey string `json:"admin_key" jsonschema:"minLength=8,maxLength=400" jsonschema_description:"An admin API key on that project." error:"admin_key must be 8-400 characters"`
}

// setProjectKey gives the portal a new admin key for a project, checked first.
//
// @route("POST /api/admin/projects/{project_id}/key", summary="Set a project's admin key",
//     tag="Catalog", role=ADMIN, access=SESSION_ONLY, errors=(404, 424),
//     description="Replace the admin API key the portal uses on a project. It must be an admin key "
//         "there. Revoking the old one on the project signs out everyone it signed in.",
// )
func ( handlers *Handlers ) setProjectKey( c fiber.Ctx , in *projectKeyRequest ) ( reply *catalogReply , err error ) {
	project , get_err := models.GetProject( handlers.Store , in.ProjectID )
	if get_err != nil { return nil , projectFieldsRefusal( get_err ) }
	sealed , err := handlers.checkAdminKey( project.Name , project.BaseURL , strings.TrimSpace( in.AdminKey ) )
	if err != nil { return }
	if project , err = models.SetProjectKey( handlers.Store , in.ProjectID , sealed ); err != nil { return nil , projectFieldsRefusal( err ) }
	handlers.audit( c , logs.KindAdmin , "project key set" , fiber.Map{ "project_id": project.ID } )
	if checked , check_err := portal.Check( handlers.Store , handlers.cfg().SecretKey , project ); check_err == nil { project = checked }
	reply , err = handlers.catalogReplyFor( project )
	return
}

func ( handlers *Handlers ) catalogReplyFor( project *models.Project ) ( reply *catalogReply , err error ) {
	counts , err := models.MemberCounts( handlers.Store )
	if err != nil { return }
	reply = &catalogReply{ Project: newCatalogRow( handlers.cfg() , project , counts[ project.ID ] ) }
	return
}

// rotateProjectKey is Rotate now: a new secret for the portal's key there.
//
// @route("POST /api/admin/projects/{project_id}/rotate-key", summary="Rotate a project's key",
//     tag="Catalog", role=ADMIN, errors=(404, 409, 424),
//     description="Swap the portal's admin key on the project for a new secret, now. The key keeps its "
//         "grants and its expiry there, and what it made keeps working. Needs an app with key rotation; "
//         "the key-rotation job also does this every catalog.rotate_key_days.",
//     mcp="admin_rotate_project_key",
// )
func ( handlers *Handlers ) rotateProjectKey( c fiber.Ctx , in *projectIDPath ) ( reply *catalogReply , err error ) {
	project , get_err := models.GetProject( handlers.Store , in.ProjectID )
	if get_err != nil { return nil , projectFieldsRefusal( get_err ) }
	if portal.MayRotate( project ) == false { return nil , refuse( fiber.StatusConflict , project.Name + ": " + portal.ErrCannotRotate.Error() ) }
	rotated , rotate_err := portal.Rotate( handlers.Store , handlers.cfg().SecretKey , project )
	if rotate_err != nil { return nil , refuse( fiber.StatusFailedDependency , rotate_err.Error() ) }
	handlers.audit( c , logs.KindAdmin , "project key rotated" , fiber.Map{ "project_id": project.ID } )
	reply , err = handlers.catalogReplyFor( rotated )
	return
}

// checkProject is Check now: is it up, does its key still work.
//
// @route("POST /api/admin/projects/{project_id}/check", summary="Check a project",
//     tag="Catalog", role=ADMIN, errors=(404,),
//     description="Ask the project now whether it is up and whether the portal's admin key still works "
//         "there. The project-health job does this every 5 minutes.",
//     mcp="admin_check_project",
// )
func ( handlers *Handlers ) checkProject( c fiber.Ctx , in *projectIDPath ) ( reply *catalogReply , err error ) {
	project , get_err := models.GetProject( handlers.Store , in.ProjectID )
	if get_err != nil { return nil , projectFieldsRefusal( get_err ) }
	if project , err = portal.Check( handlers.Store , handlers.cfg().SecretKey , project ); err != nil { return }
	reply , err = handlers.catalogReplyFor( project )
	return
}

// deleteProject takes a project out of the catalog. Undoable for 10
// minutes; nothing on the project itself changes.
//
// @route("POST /api/admin/projects/{project_id}/delete", summary="Delete a project",
//     tag="Catalog", role=ADMIN, errors=(404,),
//     description="Take a project out of the portal, with its key and everyone's membership of it. "
//         "Accounts on the project itself stay. admin_restore_project undoes it within 10 minutes.",
//     mcp="admin_delete_project", destructive=True,
// )
func ( handlers *Handlers ) deleteProject( c fiber.Ctx , in *projectIDPath ) ( reply *okReply , err error ) {
	if delete_err := models.DeleteProject( handlers.Store , in.ProjectID , security.UserFrom( c ).ID ); delete_err != nil {
		return nil , projectFieldsRefusal( delete_err )
	}
	handlers.audit( c , logs.KindAdmin , "project deleted" , fiber.Map{ "project_id": in.ProjectID } )
	reply = &okReply{ Ok: true }
	return
}

// restoreProject is Undo for deleteProject.
//
// @route("POST /api/admin/projects/{project_id}/restore", summary="Undo deleting a project",
//     tag="Catalog", role=ADMIN, errors=(404, 409),
//     description="Put back a project deleted in the last 10 minutes.",
//     mcp="admin_restore_project",
// )
func ( handlers *Handlers ) restoreProject( c fiber.Ctx , in *projectIDPath ) ( reply *okReply , err error ) {
	if restore_err := models.RestoreProject( handlers.Store , in.ProjectID , security.UserFrom( c ).ID ); restore_err != nil {
		return nil , restoreRefusal( restore_err )
	}
	handlers.audit( c , logs.KindAdmin , "project restored" , fiber.Map{ "project_id": in.ProjectID } )
	reply = &okReply{ Ok: true }
	return
}

type orderProjectsRequest struct {
	ProjectIDs []uint64 `json:"project_ids" jsonschema:"minItems=1,maxItems=1000" jsonschema_description:"Project ids in the order people should see them." error:"project_ids must list 1-1000 ids"`
}

// orderProjects is the catalog's order, which everyone's list follows.
//
// @route("POST /api/admin/projects/order", summary="Order the catalog",
//     tag="Catalog", role=ADMIN,
//     description="Put projects in this order everywhere they are listed. Ids left out follow.",
//     mcp="admin_order_catalog",
// )
func ( handlers *Handlers ) orderProjects( c fiber.Ctx , in *orderProjectsRequest ) ( reply *okReply , err error ) {
	if err = models.OrderProjects( handlers.Store , in.ProjectIDs ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "catalog reordered" , fiber.Map{ "projects": len( in.ProjectIDs ) } )
	reply = &okReply{ Ok: true }
	return
}

// accessRow is one person and one project, as the Access screen shows it.
type accessRow struct {
	UserID       uint64 `json:"user_id"`
	DisplayName  string `json:"display_name"`
	Role         string `json:"role"`
	Disabled     bool   `json:"disabled"`
	Allowed      bool   `json:"allowed" jsonschema_description:"Whether they may see and join it. Admins always may."`
	Member       bool   `json:"member"`
	Via          string `json:"via"`
	RemoteUserID uint64 `json:"remote_user_id"`
	RemoteName   string `json:"remote_name"`
	OwnKey       bool   `json:"own_key"`
}

func ( accessRow ) APIName() ( name string ) { name = "ProjectAccess"; return }

type projectAccessRequest struct {
	projectIDPath
	Page
	Q    string `json:"q,omitempty" query:"q" jsonschema_description:"Keep people whose name contains this, in any case."`
	Sort string `json:"sort,omitempty" query:"sort" jsonschema:"enum=display_name,enum=allowed,enum=member" jsonschema_description:"Field to sort by."`
	Direction
}

type projectAccessReply struct {
	Rows []accessRow `json:"rows"`
	Paged
}

// projectAccess is everyone in the portal against one project.
//
// @route("GET /api/admin/projects/{project_id}/access", summary="Who may use a project",
//     tag="Catalog", role=ADMIN, errors=(404,),
//     description="Every account here, a page at a time, with whether it may use the project and "
//         "whether it is on it.",
//     mcp="admin_project_access", read_only=True,
// )
func ( handlers *Handlers ) projectAccess( c fiber.Ctx , in *projectAccessRequest ) ( reply *projectAccessReply , err error ) {
	project , get_err := models.GetProject( handlers.Store , in.ProjectID )
	if get_err != nil { return nil , projectFieldsRefusal( get_err ) }
	users , err := models.ListUsers( handlers.Store )
	if err != nil { return }
	members , err := models.MembersOf( handlers.Store , project.ID )
	if err != nil { return }
	query := strings.ToLower( strings.TrimSpace( in.Q ) )
	rows := []accessRow{}
	for _ , user := range users {
		if query != "" && strings.Contains( strings.ToLower( user.DisplayName ) , query ) == false { continue }
		row := accessRow{
			UserID: user.ID , DisplayName: user.DisplayName , Role: user.Role , Disabled: user.Disabled() ,
			Allowed: user.IsAdmin() || project.Excludes( user.ID ) == false ,
		}
		if membership := members[ user.ID ]; membership != nil {
			row.Member , row.Via , row.RemoteUserID = true , membership.Via , membership.RemoteUserID
			row.RemoteName , row.OwnKey = membership.RemoteName , membership.PersonalKey != ""
		}
		rows = append( rows , row )
	}
	reply = &projectAccessReply{}
	reply.Rows , reply.Paged = sortedPage( rows , in.Sort , in.Direction , in.Page )
	return
}

type setAccessRequest struct {
	projectIDPath
	UserID  uint64 `json:"user_id" path:"user_id" jsonschema:"minimum=1" jsonschema_description:"The account's id here, from admin_list_users." error:"invalid user id"`
	Allowed bool   `json:"allowed" jsonschema_description:"false keeps the project from them: it leaves their list, and the portal will not sign them into it."`
	OnProject string `json:"on_project,omitempty" jsonschema:"enum=keep,enum=disable,enum=revoke" jsonschema_description:"With allowed false, what happens to an account they have on the project: keep it (the default), disable it, or end its sessions, links and keys." error:"on_project must be keep, disable or revoke"`
}

type setAccessReply struct {
	Ok      bool `json:"ok"`
	Allowed bool `json:"allowed"`
}

// setAccess lets one person use a project, or keeps it from them. Projects
// are open to everyone by default; this is the exception list.
//
// @route("POST /api/admin/projects/{project_id}/access/{user_id}", summary="Let someone use a project, or not",
//     bulk=True,
//     tag="Catalog", role=ADMIN, errors=(404,),
//     description="allowed false keeps a listed project from one person: it leaves their list and the "
//         "portal stops signing them in. Their account on the project itself stays. Admins always see "
//         "everything.",
//     mcp="admin_set_project_access",
// )
func ( handlers *Handlers ) setAccess( c fiber.Ctx , in *setAccessRequest ) ( reply *setAccessReply , err error ) {
	if _ , get_err := models.GetUser( handlers.Store , in.UserID ); get_err != nil {
		err = refuse( fiber.StatusNotFound , "no such user" )
		return
	}
	if in.Allowed == false {
		if membership , get_err := models.GetMembership( handlers.Store , in.UserID , in.ProjectID ); get_err == nil && membership.Active() {
			if in.OnProject != "" && in.OnProject != onProjectKeep {
				if refused := handlers.ladderOn( c , in.UserID , "act on the project accounts of" ); refused != nil { return nil , refused }
			}
			if err = handlers.actOnRemoteAccount( membership , in.OnProject ); err != nil { return }
		}
	}
	if _ , set_err := models.SetUserExcluded( handlers.Store , in.ProjectID , in.UserID , in.Allowed == false ); set_err != nil {
		return nil , projectFieldsRefusal( set_err )
	}
	handlers.audit( c , logs.KindAdmin , "project access changed" , fiber.Map{ "project_id": in.ProjectID , "user_id": in.UserID , "allowed": in.Allowed , "on_project": in.OnProject } )
	reply = &setAccessReply{ Ok: true , Allowed: in.Allowed }
	return
}

type grantRequest struct {
	projectIDPath
	UserID        uint64 `json:"user_id" path:"user_id" jsonschema:"minimum=1" jsonschema_description:"The account's id here." error:"invalid user id"`
	Role          string `json:"role,omitempty" jsonschema:"enum=user,enum=viewer" jsonschema_description:"Their role on the project. Default user; viewer (view only) needs an app that has it." error:"role must be user or viewer"`
	ExpiresInDays int    `json:"expires_in_days,omitempty" jsonschema:"minimum=0,maximum=3650" jsonschema_description:"Their account there stops working this many days from now. 0 or left out: never." error:"expires_in_days must be 0-3650"`
}

// grantProject is joining on someone's behalf: their account on the project
// is made now, and they sign in from their own list.
//
// @route("POST /api/admin/projects/{project_id}/grant/{user_id}", summary="Put someone on a project",
//     tag="Catalog", role=ADMIN, errors=(404, 409, 424),
//     description="Make someone an account on the project now, as if they had joined, optionally view only "
//         "or expiring. They sign in from their own list; no link is handed to you.",
//     mcp="admin_grant_project",
// )
func ( handlers *Handlers ) grantProject( c fiber.Ctx , in *grantRequest ) ( reply *okReply , err error ) {
	project , get_err := models.GetProject( handlers.Store , in.ProjectID )
	if get_err != nil { return nil , projectFieldsRefusal( get_err ) }
	user , user_err := models.GetUser( handlers.Store , in.UserID )
	if user_err != nil {
		err = refuse( fiber.StatusNotFound , "no such user" )
		return
	}
	if user.IsAdmin() == false && project.Excludes( user.ID ) {
		err = refuse( fiber.StatusConflict , "this project is kept from them: allow it first" )
		return
	}
	role := in.Role
	if role == "" { role = remoteRoleUser }
	if refused := handlers.roleAllowed( project , role ); refused != nil { return nil , refused }
	expires_at := expiryFrom( in.ExpiresInDays )
	if expires_at != nil {
		if refused := handlers.needs( project , remote.CapAccountExpiry ); refused != nil { return nil , refused }
	}
	if _ , err = handlers.joinAs( c , project , user , false , role , expires_at ); err != nil { return }
	reply = &okReply{ Ok: true }
	return
}

type remoteUsersRequest struct {
	projectIDPath
	Q      string `json:"q,omitempty" query:"q" jsonschema_description:"Keep accounts whose name contains this."`
	Offset int    `json:"offset,omitempty" query:"offset" jsonschema:"minimum=0" jsonschema_description:"From the previous page's next." error:"offset must be 0 or more"`
}

type remoteUser struct {
	ID          uint64 `json:"id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Disabled    bool   `json:"disabled"`
	Owner       bool   `json:"owner"`
	ExpiresAt   *time.Time `json:"expires_at" jsonschema_description:"When it stops working there, or null for never (or an app without expiring accounts)."`
	Expired     bool   `json:"expired"`
	PortalUserID uint64 `json:"portal_user_id,omitempty" jsonschema_description:"Who it belongs to here, when someone is on the project with it."`
	PortalName  string `json:"portal_name,omitempty"`
	Kind        string `json:"kind" jsonschema_description:"person or service (the app's own word; empty before contract 3)."`
	Portal      bool   `json:"portal" jsonschema_description:"The account the portal's own admin key belongs to there."`
	// Acts is whether the portal's key may run this account at all: the
	// app's ladder keeps an admin's (a service account's too) to the owner.
	Acts bool `json:"acts" jsonschema_description:"Whether the portal's key may act on this account there: never the owner, and another admin only with the owner's key."`
}

func ( remoteUser ) APIName() ( name string ) { name = "RemoteUser"; return }

type remoteUsersReply struct {
	Users []remoteUser `json:"users"`
	Paged
}

// remoteUsers is the accounts the project itself has, to match someone here
// to the one they already use there.
//
// @route("GET /api/admin/projects/{project_id}/remote-users", summary="A project's own accounts",
//     tag="Catalog", role=ADMIN, errors=(404, 409, 424),
//     description="The accounts on the project itself, as its admin key sees them, 200 at a time.",
//     mcp="admin_list_remote_users", read_only=True,
// )
func ( handlers *Handlers ) remoteUsers( c fiber.Ctx , in *remoteUsersRequest ) ( reply *remoteUsersReply , err error ) {
	project , get_err := models.GetProject( handlers.Store , in.ProjectID )
	if get_err != nil { return nil , projectFieldsRefusal( get_err ) }
	key , refused := handlers.adminKeyOf( project )
	if refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()
	users , next , total , remote_err := remote.ListUsers( ctx , project.BaseURL , key , in.Q , in.Offset )
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
	owners , err := handlers.remoteOwners( project.ID )
	if err != nil { return }
	reply = &remoteUsersReply{ Users: []remoteUser{} , Paged: Paged{ Next: next , Total: total } }
	for _ , user := range users {
		row := remoteUser{ ID: user.ID , DisplayName: user.DisplayName , Role: user.Role , Disabled: user.Disabled , Owner: user.Owner , ExpiresAt: user.ExpiresAt , Expired: user.Expired , Kind: user.Kind }
		row.Portal = project.KeyFacts.AccountID != 0 && user.ID == project.KeyFacts.AccountID
		row.Acts = portal.MayActOn( project , user )
		if person := owners[ user.ID ]; person != nil { row.PortalUserID , row.PortalName = person.ID , person.DisplayName }
		reply.Users = append( reply.Users , row )
	}
	return
}

type linkRemoteRequest struct {
	projectIDPath
	UserID       uint64 `json:"user_id" path:"user_id" jsonschema:"minimum=1" jsonschema_description:"The account's id here." error:"invalid user id"`
	RemoteUserID uint64 `json:"remote_user_id" jsonschema:"minimum=1" jsonschema_description:"Their account's id on the project, from admin_list_remote_users." error:"invalid remote user id"`
}

// remoteOwners is who here each account on the project belongs to, by its
// id there.
func ( handlers *Handlers ) remoteOwners( project_id uint64 ) ( owners map[uint64]*models.User , err error ) {
	owners = map[uint64]*models.User{}
	members , err := models.MembersOf( handlers.Store , project_id )
	if err != nil { return }
	for user_id , membership := range members {
		if membership.RemoteUserID == 0 { continue }
		if user , get_err := models.GetUser( handlers.Store , user_id ); get_err == nil { owners[ membership.RemoteUserID ] = user }
	}
	return
}

// linkRemote matches someone here to an account the project already had, so
// opening it signs them into that one instead of a new one.
//
// @route("POST /api/admin/projects/{project_id}/link/{user_id}", summary="Match someone to their account on a project",
//     tag="Catalog", role=ADMIN, errors=(404, 409, 424),
//     description="Say which account on the project is theirs. Opening the project then signs them into "
//         "it. The project's owner, and its admins unless the portal's key is the owner's, cannot be "
//         "signed in this way: they attach their own key instead. An account already someone else's here "
//         "is 409. Only the owner links another admin; nobody links the owner.",
//     mcp="admin_link_remote_user",
// )
func ( handlers *Handlers ) linkRemote( c fiber.Ctx , in *linkRemoteRequest ) ( reply *okReply , err error ) {
	project , get_err := models.GetProject( handlers.Store , in.ProjectID )
	if get_err != nil { return nil , projectFieldsRefusal( get_err ) }
	if _ , user_err := models.GetUser( handlers.Store , in.UserID ); user_err != nil {
		err = refuse( fiber.StatusNotFound , "no such user" )
		return
	}
	// Where someone's Open lands is theirs: only the owner points another
	// admin's, and nobody points the owner's.
	if refused := handlers.ladderOn( c , in.UserID , "link a project account for" ); refused != nil { return nil , refused }
	key , refused := handlers.adminKeyOf( project )
	if refused != nil { return nil , refused }
	found , err := handlers.findRemoteUser( project , key , in.RemoteUserID )
	if err != nil { return }
	// One account there is one person's: linking a second person to it
	// would sign them in as the first.
	owners , err := handlers.remoteOwners( project.ID )
	if err != nil { return }
	if holder := owners[ found.ID ]; holder != nil && holder.ID != in.UserID {
		err = refuse( fiber.StatusConflict , "that account is already " + holder.DisplayName + "'s: take them off the project first" )
		return
	}
	if found.Owner {
		err = refuse( fiber.StatusConflict , "that is the project's owner: nobody can make them a link from outside. They attach their own key instead" )
		return
	}
	if found.Kind == remote.KindService || ( project.KeyFacts.AccountID != 0 && found.ID == project.KeyFacts.AccountID ) {
		err = refuse( fiber.StatusConflict , "that is a service account (the portal's own, or another one): nobody signs into it, so it is no one's account" )
		return
	}
	membership := models.Membership{
		UserID: in.UserID , ProjectID: project.ID , RemoteUserID: found.ID , RemoteName: found.DisplayName ,
		RemoteRole: found.Role , Via: models.ViaLinked ,
	}
	if existing , existing_err := models.GetMembership( handlers.Store , in.UserID , project.ID ); existing_err == nil && existing.Active() {
		membership.PersonalKey = existing.PersonalKey
	}
	if _ , err = models.AddMembership( handlers.Store , membership , true ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "project account linked" , fiber.Map{ "project_id": project.ID , "user_id": in.UserID , "remote_user_id": found.ID } )
	reply = &okReply{ Ok: true }
	return
}

// findRemoteUser looks an account up on the project, a page at a time.
func ( handlers *Handlers ) findRemoteUser( project *models.Project , key string , remote_user_id uint64 ) ( found *remote.User , err error ) {
	ctx , cancel := remoteContext()
	defer cancel()
	offset := 0
	for range 50 {
		users , next , _ , remote_err := remote.ListUsers( ctx , project.BaseURL , key , "" , offset )
		if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
		for index := range users {
			if users[ index ].ID == remote_user_id { return &users[ index ] , nil }
		}
		if next == nil { break }
		offset = *next
	}
	err = refuse( fiber.StatusNotFound , "the project has no such account" )
	return
}

type removeMemberRequest struct {
	projectIDPath
	UserID    uint64 `json:"user_id" path:"user_id" jsonschema:"minimum=1" jsonschema_description:"The account's id here." error:"invalid user id"`
	OnProject string `json:"on_project,omitempty" jsonschema:"enum=keep,enum=disable,enum=revoke" jsonschema_description:"What happens to their account on the project itself: keep it (the default), disable it, or end its sessions, links and keys." error:"on_project must be keep, disable or revoke"`
}

// removeMember forgets which account on the project is someone's. Nothing
// changes on the project itself.
//
// @route("POST /api/admin/projects/{project_id}/members/{user_id}/remove", summary="Take someone off a project here",
//     tag="Catalog", role=ADMIN, errors=(404,),
//     description="Forget someone's account on the project, and their own key for it. on_project says "
//         "what happens to the account there: keep (default), disable, or revoke its sessions, links and keys. "
//         "Only the owner takes another admin off; nobody takes the owner off.",
//     mcp="admin_remove_project_member", destructive=True,
// )
func ( handlers *Handlers ) removeMember( c fiber.Ctx , in *removeMemberRequest ) ( reply *okReply , err error ) {
	membership , get_err := models.GetMembership( handlers.Store , in.UserID , in.ProjectID )
	if get_err != nil {
		err = refuse( fiber.StatusNotFound , "they are not on this project" )
		return
	}
	if refused := handlers.ladderOn( c , in.UserID , "take off a project" ); refused != nil { return nil , refused }
	if err = handlers.actOnRemoteAccount( membership , in.OnProject ); err != nil { return }
	if err = models.RemoveMembership( handlers.Store , in.UserID , in.ProjectID ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "project member removed" , fiber.Map{ "project_id": in.ProjectID , "user_id": in.UserID , "on_project": in.OnProject } )
	reply = &okReply{ Ok: true }
	return
}
