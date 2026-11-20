// Package portal is what the portal does with a project beyond storing it:
// checking it is up and its key still works. The routes' Check now and the
// project-health job share it, so both record the same findings.
package portal

import (
	context "context"
	errors "errors"
	time "time"

	db "projectportal/server/db"
	models "projectportal/server/models"
	remote "projectportal/server/remote"
)

const checkTimeout = 15 * time.Second

// Check asks the project whether it is up and whether its admin key still
// works at admin, and records the answer on the project.
func Check( store *db.Store , secret_key string , project *models.Project ) ( checked *models.Project , err error ) {
	ctx , cancel := context.WithTimeout( context.Background() , checkTimeout )
	defer cancel()
	health , key_status , note := models.HealthUp , models.KeyMissing , ""
	if health_err := remote.Health( ctx , project.BaseURL ); health_err != nil {
		health , note = models.HealthDown , health_err.Error()
	}
	if project.AdminKey != "" && health == models.HealthUp {
		key , open_err := models.OpenRemoteKey( secret_key , project.AdminKey )
		if open_err != nil { return nil , open_err }
		me , key_err := remote.CheckKey( ctx , project.BaseURL , key )
		var remote_err *remote.Error
		switch {
		case key_err == nil && me.Role == models.RoleAdmin:
			key_status = models.KeyOK
		case key_err == nil:
			key_status , note = models.KeyNotAdmin , "the key works, but only at role "+me.Role
		case errors.As( key_err , &remote_err ) && remote_err.Unauthorized():
			key_status , note = models.KeyInvalid , key_err.Error()
		default:
			// Up but not answering for keys: say what it said, keep what was known.
			key_status , note = project.KeyStatus , key_err.Error()
		}
	} else if project.AdminKey != "" {
		key_status = project.KeyStatus
	}
	checked , err = models.RecordCheck( store , project.ID , health , key_status , note )
	return
}

// CheckAll checks every project, one at a time, and answers how many are up.
func CheckAll( store *db.Store , secret_key string ) ( up int , total int , err error ) {
	projects , err := models.ListProjects( store )
	if err != nil { return }
	for _ , project := range projects {
		checked , check_err := Check( store , secret_key , project )
		if check_err != nil {
			err = errors.Join( err , check_err )
			continue
		}
		total++
		if checked.Health == models.HealthUp { up++ }
	}
	return
}
