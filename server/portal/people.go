package portal

import (
	context "context"
	errors "errors"
	fmt "fmt"
	time "time"

	db "projectportal/server/db"
	models "projectportal/server/models"
	remote "projectportal/server/remote"
)

// A whole project's accounts and their devices, page by page and one account
// at a time, can take a while; each call still has callTimeout.
const syncTimeout = 2 * time.Minute

// maxUserPages caps how far a sync pages through one app (200 a page).
const maxUserPages = 50

// SyncPeople refreshes what the portal knows of every account: its own users,
// and each project's accounts and (with admin_sessions) their devices. A
// project it cannot reach keeps what the last sync saw. Then accounts that
// memberships tie to a portal user are linked to them.
func SyncPeople( store *db.Store , secret_key string ) ( accounts int , err error ) {
	here , err := portalSnapshots( store )
	if err != nil { return }
	if err = models.ReplaceSnapshots( store , models.PortalProject , here ); err != nil { return }
	accounts = len( here )
	projects , err := models.ListProjects( store )
	if err != nil { return }
	live := map[uint64]bool{}
	for _ , project := range projects {
		live[ project.ID ] = true
		// No working key: nothing to ask with. What the last sync saw stays.
		if project.AdminKey == "" || project.KeyStatus != models.KeyOK { continue }
		snapshots , sync_err := projectSnapshots( secret_key , project )
		if sync_err != nil {
			err = errors.Join( err , fmt.Errorf( "%s: %w" , project.Name , sync_err ) )
			continue
		}
		if replace_err := models.ReplaceSnapshots( store , project.ID , snapshots ); replace_err != nil {
			err = errors.Join( err , replace_err )
			continue
		}
		accounts += len( snapshots )
	}
	// Projects deleted since: their accounts are gone with them.
	if book , read_err := models.ReadPeople( store ); read_err == nil {
		forgotten := map[uint64]bool{}
		for _ , snapshot := range book.Snapshots {
			project_id := snapshot.Account.ProjectID
			if project_id == models.PortalProject || live[ project_id ] || forgotten[ project_id ] { continue }
			forgotten[ project_id ] = true
			err = errors.Join( err , models.ForgetProjectAccounts( store , project_id ) )
		}
	}
	_ , link_err := models.LinkMemberships( store )
	err = errors.Join( err , link_err )
	return
}

// portalSnapshots is the portal's own users, as accounts of project 0.
func portalSnapshots( store *db.Store ) ( snapshots []models.AccountSnapshot , err error ) {
	users , err := models.ListUsers( store )
	if err != nil { return }
	now := time.Now().UTC()
	for _ , user := range users {
		snapshot := models.AccountSnapshot{
			Account: models.Account{ ProjectID: models.PortalProject , RemoteUserID: user.ID } ,
			DisplayName: user.DisplayName , Role: user.Role , Kind: user.Kind , Disabled: user.DisabledAt != nil ,
			Devices: []models.Device{} , SyncedAt: now ,
		}
		if snapshot.Kind == "" { snapshot.Kind = models.KindPerson }
		sessions , sessions_err := models.ListSessionsForUser( store , user.ID )
		if sessions_err != nil { return nil , sessions_err }
		for _ , session := range sessions {
			snapshot.Devices = append( snapshot.Devices , models.Device{ IP: session.IP , UserAgent: session.UserAgent , Label: session.Label , CreatedAt: session.CreatedAt , LastSeenAt: session.LastSeenAt } )
		}
		snapshots = append( snapshots , snapshot )
	}
	return
}

// projectSnapshots is one project's accounts, through the portal's key.
func projectSnapshots( secret_key string , project *models.Project ) ( snapshots []models.AccountSnapshot , err error ) {
	if project.AdminKey == "" || project.KeyStatus != models.KeyOK { return nil , errors.New( "no working admin key" ) }
	key , err := models.OpenRemoteKey( secret_key , project.AdminKey )
	if err != nil { return }
	ctx , cancel := context.WithTimeout( context.Background() , syncTimeout )
	defer cancel()
	users := []remote.User{}
	offset := 0
	for page := 0 ; page < maxUserPages ; page++ {
		found , next , _ , list_err := remote.ListUsers( ctx , project.BaseURL , key , "" , offset )
		if list_err != nil { return nil , list_err }
		users = append( users , found... )
		if next == nil { break }
		offset = *next
	}
	devices := remote.Has( project.Capabilities , remote.CapAdminSessions )
	now := time.Now().UTC()
	for _ , user := range users {
		snapshot := models.AccountSnapshot{
			Account: models.Account{ ProjectID: project.ID , RemoteUserID: user.ID } ,
			DisplayName: user.DisplayName , Role: user.Role , Kind: user.Kind , Owner: user.Owner ,
			Disabled: user.Disabled , Expired: user.Expired , SyncedAt: now ,
			Portal: project.KeyFacts.AccountID != 0 && user.ID == project.KeyFacts.AccountID ,
		}
		if devices && MayActOn( project , user ) {
			if sessions , sessions_err := remote.UserSessions( ctx , project.BaseURL , key , user.ID ) ; sessions_err == nil {
				snapshot.Devices = []models.Device{}
				for _ , session := range sessions {
					snapshot.Devices = append( snapshot.Devices , models.Device{ IP: session.IP , UserAgent: session.UserAgent , Label: session.Label , CreatedAt: session.CreatedAt , LastSeenAt: session.LastSeenAt } )
				}
			}
		}
		snapshots = append( snapshots , snapshot )
	}
	return
}

// MayActOn is the app's ladder as the portal's key meets it: nobody runs the
// owner's account, and an admin's (a service account counts as one) only
// with the owner's own key. Before contract 3 the portal cannot tell, so it
// offers and lets the app answer.
func MayActOn( project *models.Project , user remote.User ) ( result bool ) {
	switch {
	case user.Owner:
		result = false
	case project.Contract < 3:
		result = true
	case user.Kind == remote.KindService || user.Role == models.RoleAdmin:
		result = project.KeyFacts.Owner
	default:
		result = true
	}
	return
}
