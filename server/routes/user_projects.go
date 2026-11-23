package routes

import (
	strings "strings"

	fiber "github.com/gofiber/fiber/v3"

	models "projectportal/server/models"
)

// The per-person side of the Access screen: every project against one
// person, so an admin ticks the apps someone may join in one place.
// Switching one on or off is the same setAccess call the Access screen uses.

type userProjectRow struct {
	ProjectID  uint64 `json:"project_id"`
	Name       string `json:"name"`
	Icon       string `json:"icon"`
	Color      string `json:"color"`
	Listed     bool   `json:"listed" jsonschema_description:"false: hidden from everyone but admins, even people allowed it."`
	Allowed    bool   `json:"allowed" jsonschema_description:"Whether they may see and join it. Admins always may."`
	Member     bool   `json:"member"`
	Via          string `json:"via"`
	RemoteUserID uint64 `json:"remote_user_id"`
	RemoteName   string `json:"remote_name"`
	OwnKey     bool   `json:"own_key"`
	Position   int    `json:"position"`
}

func ( userProjectRow ) APIName() ( name string ) { name = "UserProject"; return }

type userProjectsRequest struct {
	userIDPath
	Page
	Q    string `json:"q,omitempty" query:"q" jsonschema_description:"Keep projects whose name contains this, in any case."`
	Sort string `json:"sort,omitempty" query:"sort" jsonschema:"enum=name,enum=allowed,enum=member,enum=position" jsonschema_description:"Field to sort by. Default the catalog's order."`
	Direction
}

type userProjectsReply struct {
	Rows []userProjectRow `json:"rows"`
	Paged
}

// userProjects is every project against one person.
//
// @route("GET /api/admin/users/{user_id}/projects", summary="Which projects someone may join",
//     tag="Catalog", role=ADMIN, errors=(404,),
//     description="Every project, a page at a time, with whether this person may see and join it and "
//         "whether they are on it. Projects are off for everyone by default; switch one on with "
//         "admin_set_project_access.",
//     mcp="admin_list_user_projects", read_only=True,
// )
func ( handlers *Handlers ) userProjects( c fiber.Ctx , in *userProjectsRequest ) ( reply *userProjectsReply , err error ) {
	user , get_err := models.GetUser( handlers.Store , in.UserID )
	if get_err != nil {
		err = refuse( fiber.StatusNotFound , "no such user" )
		return
	}
	projects , err := models.ListProjects( handlers.Store )
	if err != nil { return }
	memberships , err := models.MembershipsOf( handlers.Store , user.ID )
	if err != nil { return }
	query := strings.ToLower( strings.TrimSpace( in.Q ) )
	rows := []userProjectRow{}
	for _ , project := range projects {
		if query != "" && strings.Contains( strings.ToLower( project.Name ) , query ) == false { continue }
		row := userProjectRow{
			ProjectID: project.ID , Name: project.Name , Icon: project.Icon , Color: project.Color ,
			Listed: project.Listed , Allowed: user.IsAdmin() || project.Allows( user.ID ) , Position: project.Position ,
		}
		if membership := memberships[ project.ID ]; membership != nil {
			row.Member , row.Via , row.RemoteUserID = true , membership.Via , membership.RemoteUserID
			row.RemoteName , row.OwnKey = membership.RemoteName , membership.PersonalKey != ""
		}
		rows = append( rows , row )
	}
	reply = &userProjectsReply{}
	reply.Rows , reply.Paged = sortedPage( rows , in.Sort , in.Direction , in.Page )
	return
}
