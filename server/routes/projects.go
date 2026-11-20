package routes

import (
	context "context"
	errors "errors"
	sort "sort"
	strings "strings"
	sync "sync"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	db "projectportal/server/db"
	logs "projectportal/server/logs"
	models "projectportal/server/models"
	remote "projectportal/server/remote"
	security "projectportal/server/security"
)

// The portal's own screen: every project you may use, the ones you are on,
// and a way into each. Joining makes you an account on the project with the
// admin key the portal holds for it; opening asks the project for a fresh
// login link for that account (or for your own key's account, when you
// attached one) and hands back its full address.

const (
	remoteTimeout = 20 * time.Second
	// A project added this recently has a "new" badge on the screen.
	newProjectDays = 14
)

// joinLocks stops two presses of Join making two accounts on the project.
var joinLocks sync.Map

type projectIDPath struct {
	ProjectID uint64 `json:"project_id" path:"project_id" jsonschema:"minimum=1" jsonschema_description:"The project's id, from GET /api/projects (list_projects over MCP)." error:"invalid project id"`
}

// projectCard is one project as a person sees it.
type projectCard struct {
	ID           uint64     `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	BaseURL      string     `json:"base_url"`
	Icon         string     `json:"icon"`
	Color        string     `json:"color"`
	Health       string     `json:"health" jsonschema_description:"up, down, or empty before the first check."`
	CheckedAt    *time.Time `json:"checked_at"`
	CreatedAt    time.Time  `json:"created_at"`
	New          bool       `json:"new" jsonschema_description:"Added in the last 14 days."`
	Ready        bool       `json:"ready" jsonschema_description:"The portal holds a working admin key, so joining works."`
	Member       bool       `json:"member" jsonschema_description:"You are on it: open_project signs you in."`
	Via          string     `json:"via" jsonschema_description:"How: joined, linked (an admin matched your account), own_key. Empty when not a member."`
	RemoteName   string     `json:"remote_name" jsonschema_description:"Your account's name there."`
	OwnKey       bool       `json:"own_key" jsonschema_description:"Opening uses your own key on that project."`
	Favorite     bool       `json:"favorite"`
	Tags         []string   `json:"tags"`
	Position     int        `json:"position"`
	LastOpenedAt *time.Time `json:"last_opened_at"`
	Listed       bool       `json:"listed" jsonschema_description:"false only for a project hidden from everyone but admins."`
	BoardShareURL string    `json:"board_share_url" jsonschema_description:"A view-only link to the project's TodoBoard board, or empty."`
	GitHubURL    string     `json:"github_url" jsonschema_description:"Where the project's code lives, or empty."`
}

func ( projectCard ) APIName() ( name string ) { name = "ProjectCard"; return }

func newProjectCard( secret_key string , project *models.Project , membership *models.Membership ) ( card projectCard ) {
	card = projectCard{
		ID: project.ID , Name: project.Name , Description: project.Description , BaseURL: project.BaseURL ,
		Icon: project.Icon , Color: project.Color , Health: project.Health , CheckedAt: project.CheckedAt ,
		CreatedAt: project.CreatedAt , New: time.Since( project.CreatedAt ) < newProjectDays * 24 * time.Hour ,
		Ready: project.AdminKey != "" && project.KeyStatus == models.KeyOK ,
		Tags: []string{} , Position: project.Position , Listed: project.Listed ,
		BoardShareURL: openShareLink( secret_key , project ) , GitHubURL: project.GitHubURL ,
	}
	if membership.Active() {
		card.Member , card.Via , card.RemoteName = true , membership.Via , membership.RemoteName
		card.OwnKey , card.Favorite , card.LastOpenedAt = membership.PersonalKey != "" , membership.Favorite , membership.LastOpenedAt
		if membership.Tags != nil { card.Tags = membership.Tags }
	}
	return
}

type listProjectsRequest struct {
	Page
	Q string `json:"q,omitempty" query:"q" jsonschema_description:"Keep projects whose name, description or one of your tags contains this, in any case."`
}

type listProjectsReply struct {
	Projects []projectCard `json:"projects"`
	Paged
}

// listProjects is the catalog as the caller may see it.
//
// @route("GET /api/projects", summary="Your projects",
//     tag="Projects",
//     description="Every project you may use, a page at a time: the ones you are on first (favorites "
//         "first), then the ones you could join. member says which.",
//     mcp_note="Use open_project with a project you are a member of to get a login link for it, and "
//         "join_project for one you are not.",
//     mcp="list_projects", read_only=True,
// )
func ( handlers *Handlers ) listProjects( c fiber.Ctx , in *listProjectsRequest ) ( reply *listProjectsReply , err error ) {
	actor := security.ActorFrom( c )
	projects , err := models.ListProjects( handlers.Store )
	if err != nil { return }
	memberships , err := models.MembershipsOf( handlers.Store , actor.User.ID )
	if err != nil { return }
	query := strings.ToLower( strings.TrimSpace( in.Q ) )
	cards := []projectCard{}
	for _ , project := range projects {
		if models.CanSee( actor.User.ID , actor.IsAdmin() , project ) == false { continue }
		card := newProjectCard( handlers.cfg().SecretKey , project , memberships[ project.ID ] )
		if query != "" && cardMatches( card , query ) == false { continue }
		cards = append( cards , card )
	}
	sort.SliceStable( cards , func( i , j int ) bool {
		if cards[ i ].Member != cards[ j ].Member { return cards[ i ].Member }
		if cards[ i ].Favorite != cards[ j ].Favorite { return cards[ i ].Favorite }
		return false
	} )
	reply = &listProjectsReply{}
	reply.Projects , reply.Paged = cutPage( cards , in.Page )
	return
}

func cardMatches( card projectCard , query string ) ( result bool ) {
	haystack := strings.ToLower( card.Name + " " + card.Description + " " + strings.Join( card.Tags , " " ) )
	result = strings.Contains( haystack , query )
	return
}

// visibleProject is the project, if the caller may see it: a 404 otherwise,
// the same answer as one that does not exist.
func ( handlers *Handlers ) visibleProject( c fiber.Ctx , project_id uint64 ) ( project *models.Project , refused error ) {
	actor := security.ActorFrom( c )
	project , err := models.GetProject( handlers.Store , project_id )
	if err != nil || models.CanSee( actor.User.ID , actor.IsAdmin() , project ) == false {
		project , refused = nil , refuse( fiber.StatusNotFound , "no such project" )
	}
	return
}

// adminKeyOf opens the project's admin key, or refuses saying why there is
// none to use.
func ( handlers *Handlers ) adminKeyOf( project *models.Project ) ( key string , refused error ) {
	if project.AdminKey == "" {
		refused = refuse( fiber.StatusConflict , "this project is not set up yet: an admin has to give the portal its admin key" )
		return
	}
	key , err := models.OpenRemoteKey( handlers.cfg().SecretKey , project.AdminKey )
	if err != nil { refused = err }
	return
}

// remoteRefusal turns a project's failure into the caller's 502, saying what
// the project said, and notes a key it no longer takes.
func ( handlers *Handlers ) remoteRefusal( project *models.Project , cause error , admin_key bool ) ( refused error ) {
	var remote_err *remote.Error
	if errors.As( cause , &remote_err ) == false {
		refused = cause
		return
	}
	if admin_key && remote_err.Unauthorized() {
		_ , _ = models.RecordCheck( handlers.Store , project.ID , models.HealthUp , models.KeyInvalid , remote_err.Message )
	}
	refused = refuse( fiber.StatusBadGateway , project.Name + ": " + remote_err.Message )
	return
}

func remoteContext() ( ctx context.Context , cancel context.CancelFunc ) {
	ctx , cancel = context.WithTimeout( context.Background() , remoteTimeout )
	return
}

type projectLinkReply struct {
	ProjectID uint64 `json:"project_id"`
	LoginURL  string `json:"login_url" jsonschema_description:"Open it to sign in to the project. It works once, for 15 minutes. A credential: treat it like a password."`
}

// joinProject makes the caller an account on the project and signs them in.
// A project they left earlier gets its old account back rather than a second
// one.
//
// @route("POST /api/projects/{project_id}/join", summary="Join a project",
//     tag="Projects", errors=(404, 409, 502),
//     description="Make yourself an account on a project you may use, named as you are here, and get a "
//         "login link for it. 409 when you are already on it or it is not set up yet; 502 when the "
//         "project refused or did not answer, with what it said.",
//     mcp_note="Answers login_url: give it to the person to open (it works once, for 15 minutes).",
//     mcp="join_project",
// )
func ( handlers *Handlers ) joinProject( c fiber.Ctx , in *projectIDPath ) ( reply *projectLinkReply , err error ) {
	project , refused := handlers.visibleProject( c , in.ProjectID )
	if refused != nil { return nil , refused }
	user := security.UserFrom( c )
	reply , err = handlers.joinAs( c , project , user , true )
	return
}

// joinAs is joining for user: by themself (with_link), or an admin granting
// it (no link: that is the person's to ask for).
func ( handlers *Handlers ) joinAs( c fiber.Ctx , project *models.Project , user *models.User , with_link bool ) ( reply *projectLinkReply , err error ) {
	lock_key := [2]uint64{ user.ID , project.ID }
	if _ , busy := joinLocks.LoadOrStore( lock_key , true ); busy {
		err = refuse( fiber.StatusConflict , "already joining this project" )
		return
	}
	defer joinLocks.Delete( lock_key )

	existing , get_err := models.GetMembership( handlers.Store , user.ID , project.ID )
	if get_err == nil && existing.Active() {
		err = refuse( fiber.StatusConflict , "already on this project" )
		return
	}
	key , refused := handlers.adminKeyOf( project )
	if refused != nil { return nil , refused }
	ctx , cancel := remoteContext()
	defer cancel()

	var link *remote.Link
	var remote_err error
	membership := models.Membership{ UserID: user.ID , ProjectID: project.ID , Via: models.ViaJoined , RemoteName: user.DisplayName , RemoteRole: models.RoleUser }
	if existing != nil && existing.RemoteUserID != 0 {
		// Left earlier: the account there is still theirs.
		membership = *existing
		membership.LeftAt = nil
		link , remote_err = remote.ReissueLogin( ctx , project.BaseURL , key , existing.RemoteUserID )
	} else {
		link , remote_err = remote.CreateUser( ctx , project.BaseURL , key , user.DisplayName )
		if link != nil { membership.RemoteUserID = link.RemoteUserID }
	}
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , true ) }
	if _ , err = models.AddMembership( handlers.Store , membership , true ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "project joined" , fiber.Map{ "project_id": project.ID , "user_id": user.ID , "remote_user_id": membership.RemoteUserID } )
	reply = &projectLinkReply{ ProjectID: project.ID }
	if with_link {
		reply.LoginURL = link.URL
		_ = models.MarkOpened( handlers.Store , user.ID , project.ID )
	}
	return
}

// openProject is the reason the portal exists: a fresh way into a project
// you are on, whatever happened to your cookie there.
//
// @route("POST /api/projects/{project_id}/open", summary="Sign in to a project",
//     tag="Projects", errors=(404, 409, 502),
//     description="A fresh login link for your account on a project you are on: with your own key when "
//         "you attached one, otherwise made by the portal's admin key there. 409 when you are not on it "
//         "yet (join it), 502 when the project refused, with what it said.",
//     mcp_note="Answers login_url: give it to the person to open (it works once, for 15 minutes).",
//     mcp="open_project",
// )
func ( handlers *Handlers ) openProject( c fiber.Ctx , in *projectIDPath ) ( reply *projectLinkReply , err error ) {
	project , refused := handlers.visibleProject( c , in.ProjectID )
	if refused != nil { return nil , refused }
	user := security.UserFrom( c )
	membership , get_err := models.GetMembership( handlers.Store , user.ID , project.ID )
	if get_err != nil || membership.Active() == false {
		err = refuse( fiber.StatusConflict , "you are not on this project yet: join it first" )
		return
	}
	ctx , cancel := remoteContext()
	defer cancel()
	var link *remote.Link
	var remote_err error
	used_admin_key := membership.PersonalKey == ""
	if used_admin_key {
		key , key_refused := handlers.adminKeyOf( project )
		if key_refused != nil { return nil , key_refused }
		link , remote_err = remote.ReissueLogin( ctx , project.BaseURL , key , membership.RemoteUserID )
	} else {
		key , open_err := models.OpenRemoteKey( handlers.cfg().SecretKey , membership.PersonalKey )
		if open_err != nil { return nil , open_err }
		link , remote_err = remote.OwnLoginLink( ctx , project.BaseURL , key )
	}
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , used_admin_key ) }
	_ = models.MarkOpened( handlers.Store , user.ID , project.ID )
	handlers.audit( c , logs.KindAuth , "project sign-in link made" , fiber.Map{ "project_id": project.ID , "own_key": used_admin_key == false } )
	reply = &projectLinkReply{ ProjectID: project.ID , LoginURL: link.URL }
	return
}

type ownKeyRequest struct {
	projectIDPath
	Key string `json:"key" jsonschema:"minLength=8,maxLength=400" jsonschema_description:"An API key of your own on that project, allowed to make sign-in links there." error:"key must be 8-400 characters"`
}

type ownKeyReply struct {
	Ok         bool   `json:"ok"`
	RemoteName string `json:"remote_name" jsonschema_description:"The account the key acts as there."`
	RemoteRole string `json:"remote_role"`
}

// setOwnKey attaches the caller's own key on a project: the way in for an
// account the portal's admin key may not make links for (that project's
// owner, or another admin), or one the person already had. The key is
// checked against the project before it is kept.
//
// Browser only: a key that could set keys could point the portal at an
// account of its choosing.
//
// @route("POST /api/projects/{project_id}/own-key", summary="Use your own key on a project",
//     tag="Projects", access=SESSION_ONLY, errors=(404, 502),
//     description="Keep an API key of yours on that project, and sign in with it from now on. On the "
//         "project, switch on Sign-in links for that key (API keys card), or opening it will be refused.",
// )
func ( handlers *Handlers ) setOwnKey( c fiber.Ctx , in *ownKeyRequest ) ( reply *ownKeyReply , err error ) {
	project , refused := handlers.visibleProject( c , in.ProjectID )
	if refused != nil { return nil , refused }
	user := security.UserFrom( c )
	ctx , cancel := remoteContext()
	defer cancel()
	key := strings.TrimSpace( in.Key )
	me , remote_err := remote.CheckKey( ctx , project.BaseURL , key )
	if remote_err != nil { return nil , handlers.remoteRefusal( project , remote_err , false ) }
	sealed , err := models.SealRemoteKey( handlers.cfg().SecretKey , key )
	if err != nil { return }
	existing , get_err := models.GetMembership( handlers.Store , user.ID , project.ID )
	membership := models.Membership{ UserID: user.ID , ProjectID: project.ID , Via: models.ViaOwnKey }
	if get_err == nil && existing.Active() { membership = *existing }
	membership.PersonalKey , membership.RemoteName , membership.RemoteRole = sealed , me.DisplayName , me.Role
	if _ , err = models.AddMembership( handlers.Store , membership , true ); err != nil { return }
	handlers.audit( c , logs.KindAdmin , "own project key attached" , fiber.Map{ "project_id": project.ID , "remote_user_id": me.ID } )
	reply = &ownKeyReply{ Ok: true , RemoteName: me.DisplayName , RemoteRole: me.Role }
	return
}

// removeOwnKey stops using the caller's own key on a project. Opening goes
// back to the portal's admin key, if the portal made their account there.
//
// @route("POST /api/projects/{project_id}/own-key/remove", summary="Stop using your own key on a project",
//     tag="Projects", errors=(404,),
//     description="Forget your own key on that project. If the key was all that put you on it, you are "
//         "no longer on it here.",
//     mcp="remove_own_project_key", destructive=True,
// )
func ( handlers *Handlers ) removeOwnKey( c fiber.Ctx , in *projectIDPath ) ( reply *okReply , err error ) {
	if _ , refused := handlers.visibleProject( c , in.ProjectID ); refused != nil { return nil , refused }
	user := security.UserFrom( c )
	if _ , err = models.SetPersonalKey( handlers.Store , user.ID , in.ProjectID , "" ); errors.Is( err , db.ErrNotFound ) {
		err = refuse( fiber.StatusNotFound , "you have no key on this project" )
		return
	}
	if err != nil { return }
	handlers.audit( c , logs.KindAdmin , "own project key removed" , fiber.Map{ "project_id": in.ProjectID } )
	reply = &okReply{ Ok: true }
	return
}

type favoriteRequest struct {
	projectIDPath
	Favorite bool `json:"favorite"`
}

type membershipReply struct {
	Project projectCard `json:"project"`
}

// membershipChanged answers a change to the caller's membership: the card as
// it is now, or the refusal the change ran into.
func ( handlers *Handlers ) membershipChanged( project_id uint64 , membership *models.Membership , cause error ) ( reply *membershipReply , err error ) {
	switch {
	case errors.Is( cause , db.ErrNotFound ):
		err = refuse( fiber.StatusConflict , "you are not on this project" )
		return
	case errors.Is( cause , models.ErrBadTags ):
		err = refuse( fiber.StatusBadRequest , cause.Error() )
		return
	case errors.Is( cause , models.ErrRestoreGone ):
		err = refuse( fiber.StatusNotFound , cause.Error() )
		return
	case errors.Is( cause , models.ErrRestoreExpired ):
		err = refuse( fiber.StatusConflict , cause.Error() )
		return
	case cause != nil:
		err = cause
		return
	}
	project , err := models.GetProject( handlers.Store , project_id )
	if err != nil { return }
	reply = &membershipReply{ Project: newProjectCard( handlers.cfg().SecretKey , project , membership ) }
	return
}

// favoriteProject stars a project, which keeps it at the top of the list.
//
// @route("POST /api/projects/{project_id}/favorite", summary="Star or unstar a project",
//     tag="Projects", errors=(404, 409),
//     description="favorite true keeps the project at the top of your list.",
//     mcp="favorite_project",
// )
func ( handlers *Handlers ) favoriteProject( c fiber.Ctx , in *favoriteRequest ) ( reply *membershipReply , err error ) {
	if _ , refused := handlers.visibleProject( c , in.ProjectID ); refused != nil { return nil , refused }
	membership , cause := models.SetFavorite( handlers.Store , security.UserFrom( c ).ID , in.ProjectID , in.Favorite )
	reply , err = handlers.membershipChanged( in.ProjectID , membership , cause )
	return
}

type tagsRequest struct {
	projectIDPath
	Tags []string `json:"tags" jsonschema:"maxItems=10" jsonschema_description:"Up to 10, each 1-30 characters. They are lowercased. An empty list clears them." error:"up to 10 tags, each 1-30 characters"`
}

// tagProject is the caller's own labels for a project, to filter by.
//
// @route("POST /api/projects/{project_id}/tags", summary="Tag a project",
//     tag="Projects", errors=(404, 409),
//     description="Your own labels for a project you are on, replacing the ones it had.",
//     mcp="tag_project",
// )
func ( handlers *Handlers ) tagProject( c fiber.Ctx , in *tagsRequest ) ( reply *membershipReply , err error ) {
	if _ , refused := handlers.visibleProject( c , in.ProjectID ); refused != nil { return nil , refused }
	membership , cause := models.SetTags( handlers.Store , security.UserFrom( c ).ID , in.ProjectID , in.Tags )
	reply , err = handlers.membershipChanged( in.ProjectID , membership , cause )
	return
}

// leaveProject takes a project off the caller's list. Their account on the
// project stays, and joining again gives it back.
//
// @route("POST /api/projects/{project_id}/leave", summary="Leave a project",
//     tag="Projects", errors=(404, 409),
//     description="Take a project off your list. Your account on the project stays as it is; joining "
//         "again signs you back into it. Undo with restore_project_membership within 10 minutes keeps "
//         "your stars, tags and key.",
//     mcp="leave_project", destructive=True,
// )
func ( handlers *Handlers ) leaveProject( c fiber.Ctx , in *projectIDPath ) ( reply *okReply , err error ) {
	if _ , refused := handlers.visibleProject( c , in.ProjectID ); refused != nil { return nil , refused }
	if err = models.LeaveProject( handlers.Store , security.UserFrom( c ).ID , in.ProjectID ); errors.Is( err , db.ErrNotFound ) {
		err = refuse( fiber.StatusConflict , "you are not on this project" )
		return
	}
	if err != nil { return }
	handlers.audit( c , logs.KindAdmin , "project left" , fiber.Map{ "project_id": in.ProjectID } )
	reply = &okReply{ Ok: true }
	return
}

// restoreMembership is Undo for leaveProject.
//
// @route("POST /api/projects/{project_id}/restore", summary="Undo leaving a project",
//     tag="Projects", errors=(404, 409),
//     description="Put back a project you left in the last 10 minutes, as it was.",
//     mcp="restore_project_membership",
// )
func ( handlers *Handlers ) restoreMembership( c fiber.Ctx , in *projectIDPath ) ( reply *membershipReply , err error ) {
	if _ , refused := handlers.visibleProject( c , in.ProjectID ); refused != nil { return nil , refused }
	membership , cause := models.UndoLeave( handlers.Store , security.UserFrom( c ).ID , in.ProjectID )
	reply , err = handlers.membershipChanged( in.ProjectID , membership , cause )
	return
}
