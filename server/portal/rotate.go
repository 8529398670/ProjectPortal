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

// ErrCannotRotate is a project whose app cannot rotate keys, or that has no
// working key to rotate.
var ErrCannotRotate = errors.New( "this project cannot rotate its key from here: it needs a working admin key and an app with key rotation (portal contract v3)" )

// lostKeyNote is what the Catalog row says when a rotation's answer never
// arrived and the old secret no longer works.
const lostKeyNote = "rotating the key lost it (the app took the old one but its answer never arrived): in that app, Connect ProjectPortal again and put the new key on the project here"

// MayRotate is whether project's key can be rotated from here.
func MayRotate( project *models.Project ) ( result bool ) {
	result = project.AdminKey != "" && project.KeyStatus == models.KeyOK && remote.Has( project.Capabilities , remote.CapKeyRotation )
	return
}

// RotationDue is whether the key-rotation job should rotate project's key
// now: it can, and it was stored or last rotated at least after ago. A key
// stored before the portal recorded when counts as due.
func RotationDue( project *models.Project , after time.Duration , now time.Time ) ( due bool ) {
	if after <= 0 || MayRotate( project ) == false { return }
	due = project.KeyRotatedAt == nil || now.Sub( *project.KeyRotatedAt ) >= after
	return
}

// Rotate swaps the portal's admin key on project for a new secret and stores
// it. The app keeps the key's id, grants and expiry, so what the key made
// keeps working. If the answer never arrives, the old secret may be dead
// already: the project is checked again, and the row says so.
func Rotate( store *db.Store , secret_key string , project *models.Project ) ( rotated *models.Project , err error ) {
	if MayRotate( project ) == false { return nil , ErrCannotRotate }
	key , err := models.OpenRemoteKey( secret_key , project.AdminKey )
	if err != nil { return }
	ctx , cancel := context.WithTimeout( context.Background() , checkTimeout )
	defer cancel()
	credential , rotate_err := remote.RotateSelf( ctx , project.BaseURL , key )
	if rotate_err != nil {
		checked , check_err := Check( store , secret_key , project )
		if check_err == nil && checked.KeyStatus == models.KeyInvalid {
			_ , _ = models.RecordCheck( store , project.ID , checked.Health , models.KeyInvalid , lostKeyNote )
		}
		err = fmt.Errorf( "%s: %w" , project.Name , rotate_err )
		return
	}
	sealed , err := models.SealRemoteKey( secret_key , credential )
	if err != nil { return }
	if rotated , err = models.RecordRotatedKey( store , project.ID , project.AdminKey , sealed ); err != nil {
		if errors.Is( err , models.ErrKeyChanged ) {
			// Someone put a new key on it meanwhile: theirs stands, and the
			// rotated one is the old key's, which they replaced anyway.
			err = fmt.Errorf( "%s: the key was changed while it was rotating, so the new one was kept" , project.Name )
		}
		return
	}
	if checked , check_err := Check( store , secret_key , rotated ); check_err == nil { rotated = checked }
	return
}

// RotateDue rotates every project whose key is due, one at a time.
func RotateDue( store *db.Store , secret_key string , after time.Duration ) ( rotated int , err error ) {
	if after <= 0 { return }
	projects , err := models.ListProjects( store )
	if err != nil { return }
	now := time.Now()
	for _ , project := range projects {
		if RotationDue( project , after , now ) == false { continue }
		if _ , rotate_err := Rotate( store , secret_key , project ); rotate_err != nil {
			err = errors.Join( err , rotate_err )
			continue
		}
		rotated++
	}
	return
}
