package control

import (
	context "context"
	fmt "fmt"
	http "net/http"
	strings "strings"
	time "time"

	db "projectportal/server/db"
	models "projectportal/server/models"
	remote "projectportal/server/remote"
)

// The catalog from the shell: `manage list-projects` and `manage
// set-project-key`. set-project-key is the Catalog screen's Admin key for an
// install where a fresh key is piped in from another app's `manage
// portal-connect`, so it never shows on a screen or in a shell history.

// ProjectResponse is one catalog row, without any key.
type ProjectResponse struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	BaseURL   string `json:"base_url"`
	HasKey    bool   `json:"has_key"`
	KeyStatus string `json:"key_status,omitempty"`
}

// SetProjectKeyRequest is the body of POST /v1/projects/key.
type SetProjectKeyRequest struct {
	Project  string `json:"project"`
	AdminKey string `json:"admin_key"`
}

func projectResponse( project *models.Project ) ( row ProjectResponse ) {
	row = ProjectResponse{ ID: project.ID , Name: project.Name , BaseURL: project.BaseURL , HasKey: project.AdminKey != "" , KeyStatus: project.KeyStatus }
	return
}

// ListProjectRows is every project in the catalog, in catalog order.
func ListProjectRows( store *db.Store ) ( rows []ProjectResponse , err error ) {
	projects , err := models.ListProjects( store )
	if err != nil { return }
	rows = []ProjectResponse{}
	for _ , project := range projects { rows = append( rows , projectResponse( project ) ) }
	return
}

// SetProjectKey finds a project by id, name or address and gives it a new
// admin key, after asking the project that the key is an admin key there --
// what POST /api/admin/projects/{id}/key does. The key is never in an error.
func SetProjectKey( store *db.Store , secret_key string , which string , key string ) ( row ProjectResponse , err error ) {
	key = strings.TrimSpace( key )
	if len( key ) < 8 || len( key ) > 400 {
		err = fmt.Errorf( "the admin key must be 8-400 characters" )
		return
	}
	project , err := findProject( store , which )
	if err != nil { return }
	ctx , cancel := context.WithTimeout( context.Background() , 20 * time.Second )
	defer cancel()
	me , check_err := remote.CheckKey( ctx , project.BaseURL , key )
	if check_err != nil {
		err = fmt.Errorf( "%s: %v" , project.Name , check_err )
		return
	}
	if me.Role != models.RoleAdmin {
		err = fmt.Errorf( "%s: that key is a %s key there; the portal needs an admin key" , project.Name , me.Role )
		return
	}
	sealed , err := models.SealRemoteKey( secret_key , key )
	if err != nil { return }
	if project , err = models.SetProjectKey( store , project.ID , sealed ); err != nil { return }
	row = projectResponse( project )
	return
}

// findProject matches an id, then a name (any case), then an address.
func findProject( store *db.Store , which string ) ( found *models.Project , err error ) {
	which = strings.TrimSpace( which )
	projects , err := models.ListProjects( store )
	if err != nil { return }
	for _ , project := range projects {
		if fmt.Sprint( project.ID ) == which { found = project; return }
	}
	for _ , project := range projects {
		if strings.EqualFold( project.Name , which ) || strings.TrimRight( project.BaseURL , "/" ) == strings.TrimRight( which , "/" ) { found = project; return }
	}
	err = fmt.Errorf( "no project %q in the catalog (manage list-projects)" , which )
	return
}

func ( server *Server ) handleListProjects( w http.ResponseWriter , r *http.Request ) {
	rows , err := ListProjectRows( server.store )
	if err != nil {
		writeError( w , http.StatusInternalServerError , err.Error() )
		return
	}
	writeJSON( w , http.StatusOK , rows )
}

func ( server *Server ) handleSetProjectKey( w http.ResponseWriter , r *http.Request ) {
	body := SetProjectKeyRequest{}
	if decodeBody( w , r , &body ) == false { return }
	row , err := SetProjectKey( server.store , server.cfg().SecretKey , body.Project , body.AdminKey )
	if err != nil {
		writeError( w , http.StatusBadRequest , err.Error() )
		return
	}
	writeJSON( w , http.StatusOK , row )
}

// ListProjects is `manage list-projects` over the socket.
func ( client *Client ) ListProjects() ( rows []ProjectResponse , err error ) {
	err = client.do( http.MethodGet , "/v1/projects" , nil , &rows )
	return
}

// SetProjectKey is `manage set-project-key` over the socket.
func ( client *Client ) SetProjectKey( which string , key string ) ( row ProjectResponse , err error ) {
	err = client.do( http.MethodPost , "/v1/projects/key" , SetProjectKeyRequest{ Project: which , AdminKey: key } , &row )
	return
}
