package routes

import (
	errors "errors"
	sort "sort"
	time "time"

	fiber "github.com/gofiber/fiber/v3"

	apidoc "projectportal/server/apidoc"
	db "projectportal/server/db"
	jobs "projectportal/server/jobs"
	logs "projectportal/server/logs"
	models "projectportal/server/models"
	portal "projectportal/server/portal"
	security "projectportal/server/security"
)

// People across the catalog (phase 3b): one real person and their accounts
// in every app, this portal included (project 0). What each app has comes
// from the people-sync job's last look; nothing here calls an app. Apps never
// learn about people.

// PeopleSyncJob is the job that refreshes what the portal knows of accounts.
const PeopleSyncJob = "people-sync"

type personAccount struct {
	ProjectID    uint64          `json:"project_id" jsonschema_description:"The project, or 0 for this portal's own accounts."`
	ProjectName  string          `json:"project_name"`
	RemoteUserID uint64          `json:"remote_user_id" jsonschema_description:"The account's id there (a user id here for project 0)."`
	DisplayName  string          `json:"display_name"`
	Role         string          `json:"role"`
	Kind         string          `json:"kind" jsonschema_description:"person or service; empty from an app before contract 3."`
	Owner        bool            `json:"owner"`
	Portal       bool            `json:"portal" jsonschema_description:"The account the portal's own key belongs to there."`
	HandMade     bool            `json:"hand_made_portal" jsonschema_description:"A person's account named like the portal's: one made by hand for the portal before Connect. Delete it in that app."`
	Disabled     bool            `json:"disabled"`
	Expired      bool            `json:"expired"`
	Devices      []models.Device `json:"devices" jsonschema_description:"Its signed-in devices, when devices_listed."`
	Listed       bool            `json:"devices_listed" jsonschema_description:"The app listed this account's devices (it has admin_sessions and the portal's key may read them)."`
	SyncedAt     time.Time       `json:"synced_at"`
}

func ( personAccount ) APIName() ( name string ) { name = "PersonAccount"; return }

type personRow struct {
	ID          uint64          `json:"id"`
	DisplayName string          `json:"display_name"`
	Notes       string          `json:"notes"`
	Accounts    []personAccount `json:"accounts"`
}

func ( personRow ) APIName() ( name string ) { name = "Person"; return }

type suggestionRow struct {
	A       []personAccount `json:"a" jsonschema_description:"One side: a person's accounts, or one account nobody has placed."`
	APerson uint64          `json:"a_person_id,omitempty"`
	B       []personAccount `json:"b"`
	BPerson uint64          `json:"b_person_id,omitempty"`
	Score   int             `json:"score"`
	Reasons []string        `json:"reasons" jsonschema_description:"same_name, shared_ip, same_agent, close_in."`
}

func ( suggestionRow ) APIName() ( name string ) { name = "PeopleSuggestion"; return }

type peopleReply struct {
	People      []personRow     `json:"people"`
	Unlinked    []personAccount `json:"unlinked" jsonschema_description:"Accounts of people nobody has placed yet."`
	Service     []personAccount `json:"service" jsonschema_description:"Service accounts (the portal's own among them) and hand-made portal accounts: never a person's."`
	Suggestions []suggestionRow `json:"suggestions"`
	SyncedAt    *time.Time      `json:"synced_at" jsonschema_description:"When the oldest account was last seen; null before the first sync."`
}

// listPeople is the People screen.
//
// @route("GET /api/admin/people", summary="People across the catalog",
//     tag="People", role=ADMIN,
//     description="Every person and their accounts in each app (project 0 is this portal), the accounts "
//         "nobody has placed, service accounts, and suggested matches with why. From the last people-sync.",
//     mcp="admin_list_people", read_only=True,
// )
func ( handlers *Handlers ) listPeople( c fiber.Ctx , in *apidoc.None ) ( reply *peopleReply , err error ) {
	book , err := models.ReadPeople( handlers.Store )
	if err != nil { return }
	names , err := handlers.projectNames()
	if err != nil { return }
	row := func( account models.Account ) ( result personAccount ) {
		result = personAccount{ ProjectID: account.ProjectID , ProjectName: names[ account.ProjectID ] , RemoteUserID: account.RemoteUserID }
		if snapshot := book.Snapshots[ account.Key() ]; snapshot != nil {
			result.DisplayName , result.Role , result.Kind , result.Owner = snapshot.DisplayName , snapshot.Role , snapshot.Kind , snapshot.Owner
			result.Portal , result.HandMade , result.Disabled , result.Expired = snapshot.Portal , snapshot.HandMadePortal() , snapshot.Disabled , snapshot.Expired
			result.Devices , result.Listed , result.SyncedAt = snapshot.Devices , snapshot.Devices != nil , snapshot.SyncedAt
		}
		if result.Devices == nil { result.Devices = []models.Device{} }
		return
	}
	rows := func( accounts []models.Account ) ( result []personAccount ) {
		result = []personAccount{}
		for _ , account := range accounts { result = append( result , row( account ) ) }
		return
	}
	reply = &peopleReply{ People: []personRow{} , Unlinked: []personAccount{} , Service: []personAccount{} , Suggestions: []suggestionRow{} }
	for _ , person := range book.People {
		reply.People = append( reply.People , personRow{ ID: person.ID , DisplayName: person.DisplayName , Notes: person.Notes , Accounts: rows( book.AccountsOf( person.ID ) ) } )
	}
	sort.Slice( reply.People , func( a int , b int ) bool { return reply.People[ a ].DisplayName < reply.People[ b ].DisplayName } )
	loose := []models.Account{}
	service := []models.Account{}
	for key , snapshot := range book.Snapshots {
		if reply.SyncedAt == nil || snapshot.SyncedAt.Before( *reply.SyncedAt ) {
			at := snapshot.SyncedAt
			reply.SyncedAt = &at
		}
		switch {
		case snapshot.Service() || snapshot.HandMadePortal():
			service = append( service , snapshot.Account )
		case book.Links[ key ] == nil:
			loose = append( loose , snapshot.Account )
		}
	}
	sortByProject := func( accounts []models.Account ) {
		sort.Slice( accounts , func( a int , b int ) bool {
			if accounts[ a ].ProjectID != accounts[ b ].ProjectID { return accounts[ a ].ProjectID < accounts[ b ].ProjectID }
			return accounts[ a ].RemoteUserID < accounts[ b ].RemoteUserID
		} )
	}
	sortByProject( loose )
	sortByProject( service )
	reply.Unlinked , reply.Service = rows( loose ) , rows( service )
	for _ , suggestion := range models.Suggest( book ) {
		reply.Suggestions = append( reply.Suggestions , suggestionRow{
			A: rows( suggestion.A.Accounts ) , APerson: suggestion.A.PersonID ,
			B: rows( suggestion.B.Accounts ) , BPerson: suggestion.B.PersonID ,
			Score: suggestion.Score , Reasons: suggestion.Reasons ,
		} )
	}
	return
}

func ( handlers *Handlers ) projectNames() ( names map[uint64]string , err error ) {
	projects , err := models.ListProjects( handlers.Store )
	if err != nil { return }
	names = map[uint64]string{ models.PortalProject: handlers.cfg().ProjectName }
	for _ , project := range projects { names[ project.ID ] = project.Name }
	return
}

// syncPeople runs people-sync now.
//
// @route("POST /api/admin/people/sync", summary="Refresh people now",
//     tag="People", role=ADMIN, status=202, errors=(409,),
//     description="Start the people-sync job now: every app's accounts and devices, and this portal's. "
//         "It runs in the background; list people again for the result. It also runs every hour.",
//     mcp="admin_sync_people",
// )
func ( handlers *Handlers ) syncPeople( c fiber.Ctx , in *apidoc.None ) ( reply *okReply , err error ) {
	if handlers.Jobs == nil {
		if _ , err = portal.SyncPeople( handlers.Store , handlers.cfg().SecretKey ); err != nil { return nil , refuse( fiber.StatusFailedDependency , err.Error() ) }
		return &okReply{ Ok: true } , nil
	}
	switch run_err := handlers.Jobs.RunNow( PeopleSyncJob ); {
	case errors.Is( run_err , jobs.ErrUnknown ):
		err = refuse( fiber.StatusNotFound , "not found" )
	case run_err != nil:
		err = refuse( fiber.StatusConflict , "people-sync is already running" )
	default:
		reply = &okReply{ Ok: true }
	}
	return
}

type accountRef struct {
	ProjectID    uint64 `json:"project_id" jsonschema:"minimum=0" jsonschema_description:"The project, or 0 for this portal."`
	RemoteUserID uint64 `json:"remote_user_id" jsonschema:"minimum=1" jsonschema_description:"The account's id there."`
}

func refsToAccounts( refs []accountRef ) ( accounts []models.Account ) {
	accounts = []models.Account{}
	for _ , ref := range refs { accounts = append( accounts , models.Account{ ProjectID: ref.ProjectID , RemoteUserID: ref.RemoteUserID } ) }
	return
}

type linkPeopleRequest struct {
	Accounts []accountRef `json:"accounts" jsonschema:"minItems=1,maxItems=200" jsonschema_description:"The accounts that are one person."`
	PersonID uint64       `json:"person_id,omitempty" jsonschema_description:"Put them under this person; 0 to use the person one of them has, or a new one."`
}

type personReply struct {
	Person *models.Person `json:"person"`
}

// linkPeople says accounts are one person: confirming a suggestion, or adding
// an account to someone.
//
// @route("POST /api/admin/people/link", summary="Link accounts to one person",
//     tag="People", role=ADMIN, errors=(400, 404),
//     description="Put accounts under one person. People they already belong to are merged into it. "
//         "Service accounts are refused: they belong to nobody.",
//     mcp="admin_link_people",
// )
func ( handlers *Handlers ) linkPeople( c fiber.Ctx , in *linkPeopleRequest ) ( reply *personReply , err error ) {
	person , link_err := models.LinkAccounts( handlers.Store , security.UserFrom( c ).ID , in.PersonID , refsToAccounts( in.Accounts ) )
	if link_err != nil { return nil , peopleRefusal( link_err ) }
	handlers.audit( c , logs.KindAdmin , "people linked" , fiber.Map{ "person_id": person.ID , "accounts": len( in.Accounts ) } )
	reply = &personReply{ Person: person }
	return
}

// unlinkPerson takes one account off its person: a split.
//
// @route("POST /api/admin/people/unlink", summary="Take an account off its person",
//     tag="People", role=ADMIN, errors=(404,),
//     description="The account no longer counts as that person's, and is not suggested for them again.",
//     mcp="admin_unlink_person",
// )
func ( handlers *Handlers ) unlinkPerson( c fiber.Ctx , in *accountRef ) ( reply *okReply , err error ) {
	if unlink_err := models.UnlinkAccount( handlers.Store , models.Account{ ProjectID: in.ProjectID , RemoteUserID: in.RemoteUserID } ); unlink_err != nil {
		return nil , peopleRefusal( unlink_err )
	}
	handlers.audit( c , logs.KindAdmin , "person account unlinked" , fiber.Map{ "project_id": in.ProjectID , "remote_user_id": in.RemoteUserID } )
	reply = &okReply{ Ok: true }
	return
}

type dismissPeopleRequest struct {
	A []accountRef `json:"a" jsonschema:"minItems=1,maxItems=200"`
	B []accountRef `json:"b" jsonschema:"minItems=1,maxItems=200"`
}

// dismissPeople says a suggestion is wrong.
//
// @route("POST /api/admin/people/dismiss", summary="Dismiss a suggested match",
//     tag="People", role=ADMIN, errors=(400,),
//     description="The accounts in a are not the person of those in b: that match is never suggested again.",
//     mcp="admin_dismiss_people",
// )
func ( handlers *Handlers ) dismissPeople( c fiber.Ctx , in *dismissPeopleRequest ) ( reply *okReply , err error ) {
	if err = models.DismissPairs( handlers.Store , refsToAccounts( in.A ) , refsToAccounts( in.B ) ); err != nil { return }
	reply = &okReply{ Ok: true }
	return
}

type mergePeopleRequest struct {
	Into uint64 `json:"into" jsonschema:"minimum=1" jsonschema_description:"The person who stays."`
	From uint64 `json:"from" jsonschema:"minimum=1" jsonschema_description:"The person folded into them, then gone."`
}

// mergePeople folds one person into another.
//
// @route("POST /api/admin/people/merge", summary="Merge two people",
//     tag="People", role=ADMIN, errors=(400, 404),
//     description="Every account of from goes to into, with its notes, and from is deleted.",
//     mcp="admin_merge_people",
// )
func ( handlers *Handlers ) mergePeople( c fiber.Ctx , in *mergePeopleRequest ) ( reply *personReply , err error ) {
	person , merge_err := models.MergePeople( handlers.Store , in.Into , in.From )
	if merge_err != nil { return nil , peopleRefusal( merge_err ) }
	handlers.audit( c , logs.KindAdmin , "people merged" , fiber.Map{ "into": in.Into , "from": in.From } )
	reply = &personReply{ Person: person }
	return
}

type editPersonRequest struct {
	PersonID    uint64 `json:"person_id" path:"person_id" jsonschema:"minimum=1" error:"invalid person id"`
	DisplayName string `json:"display_name" jsonschema:"minLength=1,maxLength=80"`
	Notes       string `json:"notes" jsonschema:"maxLength=2000"`
}

// editPerson renames a person and sets their notes.
//
// @route("POST /api/admin/people/{person_id}", summary="Rename a person",
//     tag="People", role=ADMIN, errors=(400, 404),
//     description="The name and notes the portal keeps for a person. Their accounts keep their own names.",
//     mcp="admin_edit_person",
// )
func ( handlers *Handlers ) editPerson( c fiber.Ctx , in *editPersonRequest ) ( reply *personReply , err error ) {
	person , edit_err := models.EditPerson( handlers.Store , in.PersonID , in.DisplayName , in.Notes )
	if edit_err != nil { return nil , peopleRefusal( edit_err ) }
	reply = &personReply{ Person: person }
	return
}

func peopleRefusal( cause error ) ( refused error ) {
	switch {
	case errors.Is( cause , db.ErrNotFound ):
		refused = refuse( fiber.StatusNotFound , "no such person or account" )
	case errors.Is( cause , models.ErrServiceLink ) , errors.Is( cause , models.ErrBadPersonName ) ,
		errors.Is( cause , models.ErrBadNotes ) , errors.Is( cause , models.ErrSamePerson ):
		refused = refuse( fiber.StatusBadRequest , cause.Error() )
	default:
		refused = cause
	}
	return
}
