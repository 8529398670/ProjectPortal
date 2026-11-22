package remote

import (
	context "context"
	http "net/http"
	net_url "net/url"
	slices "slices"
	strconv "strconv"
	time "time"
)

// Running an app from outside: portal contract v2 (CROSS_PROJECT.md, "Portal
// contract v2"). Each call needs the capability named on it in the app's
// /api/me; the portal checks Has before offering it, and an app that is
// behind answers 404 or 403, which reaches the admin in the app's own words.

const (
	CapViewerRole    = "viewer_role"
	CapAccountExpiry = "account_expiry"
	CapKeyMinting    = "key_minting"
	CapInviteLinks   = "invite_links"
	CapKeySignIn     = "key_sign_in"
	// Contract v3.
	CapServiceAccounts = "service_accounts"
	CapKeyGrants       = "key_grants"
	CapErrorCodes      = "error_codes"
	CapKeyRotation     = "key_rotation"
	CapAdminSessions   = "admin_sessions"
)

// Has is whether capabilities lists capability.
func Has( capabilities []string , capability string ) ( result bool ) {
	result = slices.Contains( capabilities , capability )
	return
}

func userPath( remote_user_id uint64 , action string ) ( path string ) {
	path = "/api/admin/users/" + strconv.FormatUint( remote_user_id , 10 ) + "/" + action
	return
}

// SetRole changes an account's role there: admin, user, or (viewer_role) viewer.
func SetRole( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 , role string ) ( err error ) {
	err = call( ctx , base_url , admin_key , http.MethodPost , userPath( remote_user_id , "role" ) , map[string]any{ "role": role } , nil )
	return
}

// SetExpiry is when an account stops working; nil is never. (account_expiry)
func SetExpiry( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 , expires_at *time.Time ) ( err error ) {
	body := map[string]any{ "expires_at": nil }
	if expires_at != nil { body[ "expires_at" ] = expires_at.UTC().Format( time.RFC3339 ) }
	err = call( ctx , base_url , admin_key , http.MethodPost , userPath( remote_user_id , "expiry" ) , body , nil )
	return
}

// SetDisabled switches an account off there, or back on.
func SetDisabled( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 , disabled bool ) ( err error ) {
	err = call( ctx , base_url , admin_key , http.MethodPost , userPath( remote_user_id , "disabled" ) , map[string]any{ "disabled": disabled } , nil )
	return
}

// RevokeAccess ends every session, login link and API key the account holds
// there; the account itself stays enabled.
func RevokeAccess( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 ) ( err error ) {
	err = call( ctx , base_url , admin_key , http.MethodPost , userPath( remote_user_id , "revoke-access" ) , map[string]any{} , nil )
	return
}

// DeleteUser removes the account there for good.
func DeleteUser( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 ) ( err error ) {
	err = call( ctx , base_url , admin_key , http.MethodPost , userPath( remote_user_id , "delete" ) , map[string]any{} , nil )
	return
}

// InviteLink is one of the app's invite links, as its list shows it.
type InviteLink struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	MaxUses   int       `json:"max_uses"`
	Used      int       `json:"used"`
	Remaining int       `json:"remaining"`
	Live      bool      `json:"live"`
	Revoked   bool      `json:"revoked"`
	ExpiresAt time.Time `json:"expires_at"`
	Role      string    `json:"role"`
}

// NewInviteLink is an invite link to make.
type NewInviteLink struct {
	Label   string `json:"label"`
	MaxUses int    `json:"max_uses"`
	Days    int    `json:"days"`
	Role    string `json:"role,omitempty"`
}

// ListInviteLinks is the app's invite links, newest first. (invite_links)
func ListInviteLinks( ctx context.Context , base_url string , admin_key string ) ( links []InviteLink , err error ) {
	page := struct{ InviteLinks []InviteLink `json:"invite_links"` }{}
	if err = call( ctx , base_url , admin_key , http.MethodGet , "/api/admin/invite-links?limit=200" , nil , &page ); err != nil { return }
	links = page.InviteLinks
	if links == nil { links = []InviteLink{} }
	return
}

// CreateInviteLink makes one and answers its full address, which is shown
// once. The role is sent only when it is not the default.
func CreateInviteLink( ctx context.Context , base_url string , admin_key string , link NewInviteLink ) ( url string , err error ) {
	if link.Role == "user" { link.Role = "" }
	reply := struct{ JoinPath string `json:"join_path"` }{}
	if err = call( ctx , base_url , admin_key , http.MethodPost , "/api/admin/invite-links" , link , &reply ); err != nil { return }
	if len( reply.JoinPath ) < 7 || reply.JoinPath[ :6 ] != "/join/" {
		err = &Error{ Status: http.StatusOK , Message: "the project's reply had no invite link in it" }
		return
	}
	url = base_url + reply.JoinPath
	return
}

// RevokeInviteLink stops one working.
func RevokeInviteLink( ctx context.Context , base_url string , admin_key string , link_id string ) ( err error ) {
	err = call( ctx , base_url , admin_key , http.MethodPost , "/api/admin/invite-links/" + net_url.PathEscape( link_id ) + "/revoke" , map[string]any{} , nil )
	return
}

// Key is one API key on the app, never its secret.
type Key struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	Live        bool      `json:"live"`
	Revoked     bool      `json:"revoked"`
	Expired     bool      `json:"expired"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	MintedByKey string     `json:"minted_by_key,omitempty"`
}

// NewKey is a key to mint for an account.
type NewKey struct {
	Name          string `json:"name"`
	Role          string `json:"role"`
	ExpiresInDays int    `json:"expires_in_days"`
}

// MintedKey is a key just made, with its credential, shown once.
type MintedKey struct {
	Key
	Credential string `json:"credential"`
}

// ListUserKeys is an account's API keys there. (key_minting)
func ListUserKeys( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 ) ( keys []Key , err error ) {
	page := struct{ Keys []Key `json:"keys"` }{}
	if err = call( ctx , base_url , admin_key , http.MethodGet , userPath( remote_user_id , "keys" ) , nil , &page ); err != nil { return }
	keys = page.Keys
	if keys == nil { keys = []Key{} }
	return
}

// MintKey makes an API key for an account there, on the portal's key's
// lineage: revoking the portal's key revokes it too. (key_minting)
func MintKey( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 , key NewKey ) ( minted *MintedKey , err error ) {
	minted = &MintedKey{}
	if err = call( ctx , base_url , admin_key , http.MethodPost , userPath( remote_user_id , "keys" ) , key , minted ); err != nil {
		minted = nil
		return
	}
	if minted.Credential == "" {
		minted , err = nil , &Error{ Status: http.StatusOK , Message: "the project's reply had no key in it" }
	}
	return
}

// RevokeKey stops a key there working, and every key it minted.
func RevokeKey( ctx context.Context , base_url string , admin_key string , key_id string ) ( err error ) {
	err = call( ctx , base_url , admin_key , http.MethodPost , "/api/admin/keys/" + net_url.PathEscape( key_id ) + "/revoke" , map[string]any{} , nil )
	return
}

// RotateSelf swaps the calling key's secret for a new one (key_rotation) and
// answers it. The old secret stops working at once; nothing else about the
// key changes, its expiry included.
func RotateSelf( ctx context.Context , base_url string , key string ) ( credential string , err error ) {
	reply := struct{ Credential string `json:"credential"` }{}
	if err = call( ctx , base_url , key , http.MethodPost , "/api/keys/self/rotate" , nil , &reply ); err != nil { return }
	if reply.Credential == "" {
		err = &Error{ Status: http.StatusOK , Message: "the project's reply had no new key in it" }
		return
	}
	credential = reply.Credential
	return
}

// Session is one signed-in device of an account (admin_sessions).
type Session struct {
	ID         string     `json:"id"`
	CreatedAt  time.Time  `json:"created_at"`
	LastSeenAt *time.Time `json:"last_seen_at"`
	IP         string     `json:"ip"`
	UserAgent  string     `json:"user_agent"`
	Label      string     `json:"label"`
}

// UserSessions is an account's signed-in devices there.
func UserSessions( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 ) ( sessions []Session , err error ) {
	reply := struct{ Sessions []Session `json:"sessions"` }{}
	if err = call( ctx , base_url , admin_key , http.MethodGet , userPath( remote_user_id , "sessions" ) , nil , &reply ); err != nil { return }
	sessions = reply.Sessions
	if sessions == nil { sessions = []Session{} }
	return
}
