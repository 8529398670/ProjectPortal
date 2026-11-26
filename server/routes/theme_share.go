package routes

import (
	slices "slices"
	sort "sort"
	sync "sync"

	fiber "github.com/gofiber/fiber/v3"

	logs "projectportal/server/logs"
	models "projectportal/server/models"
	remote "projectportal/server/remote"
	security "projectportal/server/security"
)

// One theme everywhere: the theme someone uses here is sent to each app they
// are on, through that app's POST /api/admin/users/{id}/theme (capability
// user_theme in the portal contract), with the portal's key there. The
// browser sends the theme itself, since the built-in ones live in
// static/js/theme.js and nowhere on the server. It sends it whenever the
// pick or the theme in use changes, while theme_share is on (the default).

// userThemeCapability is what an app's /api/me lists when it takes a theme
// from the portal.
const userThemeCapability = "user_theme"

type setThemeShareRequest struct {
	On bool `json:"on" jsonschema_description:"true sends the theme you use to your apps whenever it changes."`
}

type setThemeShareReply struct {
	Ok         bool `json:"ok"`
	ThemeShare bool `json:"theme_share"`
}

// setThemeShare switches sending the caller's theme to their apps.
//
// @route("POST /api/account/theme-share", summary="Use your theme in all your apps",
//     tag="Account", feature="appearance",
//     description="on true: whenever the theme you use changes, it is sent to every app you are on that "
//         "takes it. The sending itself is POST /api/account/theme/send.",
//     mcp="set_theme_share",
// )
func ( handlers *Handlers ) setThemeShare( c fiber.Ctx , in *setThemeShareRequest ) ( reply *setThemeShareReply , err error ) {
	if err = models.SetThemeShare( handlers.Store , security.UserFrom( c ).ID , in.On ); err != nil { return }
	reply = &setThemeShareReply{ Ok: true , ThemeShare: in.On }
	return
}

type sendThemeRequest struct {
	Theme models.Theme `json:"theme" jsonschema_description:"The theme to use in your apps, in the Theme schema."`
}

type sentTheme struct {
	ProjectID uint64 `json:"project_id"`
	Name      string `json:"name"`
	Status    string `json:"status" jsonschema:"enum=sent,enum=skipped,enum=refused" jsonschema_description:"sent; skipped: the app cannot take a theme yet, or is not set up; refused: the app said no, see note."`
	Note      string `json:"note,omitempty" jsonschema_description:"Why it was skipped or refused."`
}

type sendThemeReply struct {
	Apps []sentTheme `json:"apps" jsonschema_description:"Every app you are on, by name."`
}

// sendTheme puts a theme on the caller's account in each app they are on.
//
// @route("POST /api/account/theme/send", summary="Send a theme to all your apps",
//     tag="Account", feature="appearance", errors=(400,),
//     description="Puts the theme on your account in every app you are on and picks it there. Apps that "
//         "cannot take a theme yet (no user_theme capability) are skipped; one app saying no does not stop "
//         "the rest.",
//     mcp="send_theme_to_apps",
// )
func ( handlers *Handlers ) sendTheme( c fiber.Ctx , in *sendThemeRequest ) ( reply *sendThemeReply , err error ) {
	cleaned , clean_err := models.CleanThemes( []models.Theme{ in.Theme } )
	if clean_err != nil {
		err = refuse( fiber.StatusBadRequest , clean_err.Error() )
		return
	}
	theme := cleaned[ 0 ]
	actor := security.ActorFrom( c )
	projects , err := models.ListProjects( handlers.Store )
	if err != nil { return }
	memberships , err := models.MembershipsOf( handlers.Store , actor.User.ID )
	if err != nil { return }

	reply = &sendThemeReply{ Apps: []sentTheme{} }
	var lock sync.Mutex
	var wait sync.WaitGroup
	add := func( row sentTheme ) { lock.Lock(); reply.Apps = append( reply.Apps , row ); lock.Unlock() }
	for _ , project := range projects {
		membership := memberships[ project.ID ]
		if membership == nil || membership.Active() == false { continue }
		if models.CanSee( actor.User.ID , actor.IsAdmin() , project ) == false { continue }
		row := sentTheme{ ProjectID: project.ID , Name: project.Name }
		if slices.Contains( project.Capabilities , userThemeCapability ) == false {
			row.Status , row.Note = "skipped" , "this app cannot take a theme yet: it needs the webserver-upgrade"
			add( row )
			continue
		}
		key , refused := handlers.adminKeyOf( project )
		if refused != nil || membership.RemoteUserID == 0 {
			row.Status , row.Note = "skipped" , "the portal cannot act on your account there"
			add( row )
			continue
		}
		wait.Add( 1 )
		go func( project *models.Project , remote_user_id uint64 , row sentTheme ) {
			defer wait.Done()
			ctx , cancel := remoteContext()
			defer cancel()
			if send_err := remote.SetUserTheme( ctx , project.BaseURL , key , remote_user_id , theme ); send_err != nil {
				row.Status , row.Note = "refused" , send_err.Error()
			} else {
				row.Status = "sent"
			}
			add( row )
		}( project , membership.RemoteUserID , row )
	}
	wait.Wait()
	sort.Slice( reply.Apps , func( i , j int ) bool { return reply.Apps[ i ].Name < reply.Apps[ j ].Name } )

	sent := 0
	for _ , row := range reply.Apps { if row.Status == "sent" { sent++ } }
	handlers.audit( c , logs.KindAuth , "sent theme to apps" , fiber.Map{ "theme_id": theme.ID , "apps": len( reply.Apps ) , "sent": sent } )
	return
}
