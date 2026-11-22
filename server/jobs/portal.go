package jobs

import (
	errors "errors"
	fmt "fmt"
	time "time"

	config "projectportal/server/config"
	db "projectportal/server/db"
	models "projectportal/server/models"
	portal "projectportal/server/portal"
)

const projectHealthInterval = 5 * time.Minute

// Portal registers the portal's own jobs: checking every project in the
// catalog, and clearing out what deleted projects and accounts left behind.
func Portal( registry *Registry , store *db.Store , settings *config.Holder ) {
	registry.Register( "project-health" , projectHealthInterval , func() ( note string , err error ) {
		up , total , err := portal.CheckAll( store , settings.Get().SecretKey )
		note = fmt.Sprintf( "%d of %d projects up" , up , total )
		return
	} )
	// Hourly, so a key due today rotates within the hour; each key itself
	// rotates every catalog.rotate_key_days.
	registry.Register( "key-rotation" , housekeepingInterval , func() ( note string , err error ) {
		rotated , err := portal.RotateDue( store , settings.Get().SecretKey , settings.Get().Catalog.RotateKeyAfter )
		note = fmt.Sprintf( "rotated %d keys" , rotated )
		return
	} )
	registry.Register( "people-sync" , housekeepingInterval , func() ( note string , err error ) {
		accounts , err := portal.SyncPeople( store , settings.Get().SecretKey )
		note = fmt.Sprintf( "%d accounts seen" , accounts )
		return
	} )
	registry.Register( "portal-purge" , housekeepingInterval , func() ( note string , err error ) {
		projects , projects_err := models.PurgeDeletedProjects( store )
		orphans , orphans_err := models.PurgeOrphanMemberships( store )
		keys_err := models.PurgeLeftKeys( store )
		note = fmt.Sprintf( "removed %d deleted projects, %d memberships of deleted accounts" , projects , orphans )
		err = errors.Join( projects_err , orphans_err , keys_err )
		return
	} )
}
