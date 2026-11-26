// Package remote calls the other projects in the portal's catalog: apps built
// from the go-webserver and python-webserver templates, which share one API.
// It holds no state. Every call takes the project's address and the API key
// to send, and answers what that project said, or an *Error carrying its
// words for the person to read.
//
// What it relies on, in those apps:
//
//   - GET  /api/me                                  who a key is, at what role
//   - GET  /api/health                              the container healthcheck
//   - POST /api/admin/users                         create an account (admin key)
//   - POST /api/admin/users/{id}/reissue-login      a login link for it (admin key)
//   - GET  /api/admin/users                         list accounts (admin key)
//   - POST /api/account/login-link                  a key's own account's link,
//     once that key may make sign-in links
//
// None of those mint API keys: a key never can, there. That is why the portal
// holds one admin key per project and asks for a fresh link each time, rather
// than a key per person.
package remote

import (
	bytes "bytes"
	context "context"
	json "encoding/json"
	errors "errors"
	fmt "fmt"
	io "io"
	http "net/http"
	net_url "net/url"
	strconv "strconv"
	time "time"
)

const (
	callTimeout   = 10 * time.Second
	maxReplyBytes = 1 << 20
	// A link the portal asks for works once, for 15 minutes: the person is
	// about to open it.
	linkMinutes = 15
)

// Error is a project refusing, or not answering. Status is its HTTP status,
// 0 when it never answered. Message is safe to show: the project's own error
// text, or what went wrong reaching it -- never the key.
type Error struct {
	Status  int
	Message string
	// Code is the app's machine-readable refusal (contract v3, error_codes),
	// empty on an older app; Field names the one bad field of an "invalid".
	Code  string
	Field string
}

func ( remote_err *Error ) Error() ( result string ) {
	result = remote_err.Message
	return
}

// Unauthorized is a key the project no longer takes.
func ( remote_err *Error ) Unauthorized() ( result bool ) {
	result = remote_err.Status == http.StatusUnauthorized || remote_err.Code == CodeNotAuthenticated
	return
}

// Refusal codes (portal contract v3, "Refusals" in the webserver skills).
const (
	CodeNotAuthenticated = "not_authenticated"
	CodeNotAdmin         = "not_admin"
	CodeViewOnly         = "view_only"
	CodeSessionOnly      = "session_only"
	CodeOwnerRequired    = "owner_required"
	CodeOwnerProtected   = "owner_protected"
	CodeKeyMintingOff    = "key_minting_off"
	CodeServiceAccount   = "service_account"
	CodeNotFound         = "not_found"
	CodeInvalid          = "invalid"
	CodeConflict         = "conflict"
	CodeRateLimited      = "rate_limited"
	CodeForbidden        = "forbidden"
)

// CodeOf is err's refusal code, or "" when it has none.
func CodeOf( err error ) ( code string ) {
	var remote_err *Error
	if errors.As( err , &remote_err ) { code = remote_err.Code }
	return
}

var client = &http.Client{
	Timeout: callTimeout,
	// A redirect would send the key on to wherever it points.
	CheckRedirect: func( *http.Request , []*http.Request ) error { return http.ErrUseLastResponse },
}

// Me is GET /api/me as an API key sees it.
type Me struct {
	ID          uint64 `json:"id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	IsOwner     bool   `json:"is_owner"`
	APIKeyName  string `json:"api_key_name"`
	Version     string `json:"version"`
	// Contract and Capabilities are portal contract v2 (CROSS_PROJECT.md):
	// 0 and empty on an app from before it.
	Contract     int      `json:"contract"`
	Capabilities []string `json:"capabilities"`
	// Contract v3. Owner repeats IsOwner; Kind is "person" or "service";
	// Skill and SkillCommit are the template the app came from (empty when
	// it does not say). Key is what the calling key may do (key_grants),
	// nil for an older app.
	Owner       bool      `json:"owner"`
	Kind        string    `json:"kind"`
	Skill       string    `json:"skill"`
	SkillCommit string    `json:"skill_commit"`
	Key         *KeyGrant `json:"key"`
}

// KeyGrant is /api/me's "key": the calling key and what it may do.
type KeyGrant struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Role        string     `json:"role"`
	MayMintKeys bool       `json:"may_mint_keys"`
	MaySignIn   bool       `json:"may_sign_in"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// KindService is a service account's kind: the portal's own account.
const KindService = "service"

// User is one account as GET /api/admin/users lists it.
type User struct {
	ID          uint64 `json:"id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
	Disabled    bool   `json:"disabled"`
	Owner       bool   `json:"owner"`
	// Contract v2, account_expiry: nil for never, or on an older app.
	ExpiresAt *time.Time `json:"expires_at"`
	Expired   bool       `json:"expired"`
	// Contract v3, service_accounts: "person" or "service".
	Kind string `json:"kind"`
}

type usersPage struct {
	Users []User `json:"users"`
	Next  *int   `json:"next"`
	Total int    `json:"total"`
}

type loginLink struct {
	UserID    uint64 `json:"user_id"`
	LoginPath string `json:"login_path"`
}

// Link is a login link a project made, as a full address to open.
type Link struct {
	URL          string
	RemoteUserID uint64
}

func call( ctx context.Context , base_url string , key string , method string , path string , body any , out any ) ( err error ) {
	var reader io.Reader
	if body != nil {
		encoded , encode_err := json.Marshal( body )
		if encode_err != nil { return encode_err }
		reader = bytes.NewReader( encoded )
	}
	request , err := http.NewRequestWithContext( ctx , method , base_url + path , reader )
	if err != nil {
		err = &Error{ Message: "the project's address is not usable" }
		return
	}
	request.Header.Set( "Accept" , "application/json" )
	if body != nil { request.Header.Set( "Content-Type" , "application/json" ) }
	if key != "" { request.Header.Set( "Authorization" , "Bearer " + key ) }
	response , err := client.Do( request )
	if err != nil {
		err = &Error{ Message: "could not reach the project: " + reachFailure( err ) }
		return
	}
	defer response.Body.Close()
	raw , err := io.ReadAll( io.LimitReader( response.Body , maxReplyBytes ) )
	if err != nil {
		err = &Error{ Status: response.StatusCode , Message: "the project's reply was cut off" }
		return
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		err = refusal( response.StatusCode , raw )
		return
	}
	if out == nil { return }
	if json.Unmarshal( raw , out ) != nil {
		err = &Error{ Status: response.StatusCode , Message: "the project answered something that is not its API (is the address right?)" }
	}
	return
}

// reachFailure is why a call never got an answer, without the URL: Go's own
// error text repeats the whole request line.
func reachFailure( cause error ) ( result string ) {
	var url_err *net_url.Error
	if errors.As( cause , &url_err ) { cause = url_err.Err }
	if errors.Is( cause , context.DeadlineExceeded ) {
		result = "it did not answer in time"
		return
	}
	result = cause.Error()
	return
}

func refusal( status int , raw []byte ) ( remote_err *Error ) {
	reply := struct {
		Error string `json:"error"`
		Code  string `json:"code"`
		Field string `json:"field"`
	}{}
	_ = json.Unmarshal( raw , &reply )
	remote_err = &Error{ Status: status , Code: reply.Code , Field: reply.Field }
	remote_err.Message = codeText( status , reply.Code , reply.Field , reply.Error )
	if remote_err.Message == "" { remote_err.Message = wordText( status , reply.Error ) }
	return
}

// codeText words a contract v3 refusal by its code, or "" for none (an
// older app), to fall back on its words.
func codeText( status int , code string , field string , said string ) ( result string ) {
	switch code {
	case CodeNotAuthenticated:
		result = "the project no longer takes this key"
	case CodeNotAdmin:
		result = "the portal's key there is not an admin's: in that app, Connect ProjectPortal from its owner's API keys screen, and put that key on the project here"
	case CodeViewOnly:
		result = "the key there is view-only, so it cannot change anything"
	case CodeSessionOnly:
		result = "that project does this only from a browser session, never with a key"
	case CodeOwnerRequired:
		result = "only that project's owner may do this to an admin's account (the portal's own account included)"
	case CodeOwnerProtected:
		result = "nobody may do this to that project's owner"
	case CodeKeyMintingOff:
		result = "the portal's key there may not make or list API keys for other accounts: Connect ProjectPortal again from that app's owner screen, or turn Key minting on for the portal's key"
	case CodeServiceAccount:
		result = "that is a service account: it is never signed into, it has keys only"
	case CodeNotFound:
		result = "the project has no such account, key or link (or is behind and lacks this)"
	case CodeInvalid:
		result = "the project turned the request down: " + said
		if field != "" { result += " (" + field + ")" }
	case CodeConflict:
		result = "the project said: " + said
	case CodeRateLimited:
		result = "the project is asking the portal to slow down: try again in a minute"
	case CodeForbidden:
		result = "the project refused this: " + said
	case "":
		return
	default:
		if said == "" { said = code }
		result = "the project said: " + said
	}
	result = fmt.Sprintf( "%s (%d)" , result , status )
	return
}

// wordText words a refusal from an app before contract v3, by its text.
func wordText( status int , said string ) ( result string ) {
	switch {
	case said == "key_minting_off":
		result = "the portal's key there may not make or list API keys for other accounts: in that app, turn Key minting on for the portal's key in its API keys list (403)"
	case said == "forbidden" && status == http.StatusForbidden:
		// An admin key that an app turns away anyway: some apps (Rook) let
		// only their owner manage accounts.
		result = "the portal's key there is an admin's, but that project lets only its owner do this: in the project, sign in as its owner, make an admin API key, and put that key on the project here (403)"
	case said != "":
		result = fmt.Sprintf( "the project said: %s (%d)" , said , status )
	case status == http.StatusUnauthorized:
		result = "the project no longer takes this key (401)"
	case status >= 300 && status < 400:
		result = fmt.Sprintf( "the project redirected (%d): check the address, https included" , status )
	default:
		result = fmt.Sprintf( "the project answered %d" , status )
	}
	return
}

func linkRequest() ( body map[string]any ) {
	body = map[string]any{ "max_uses": 1 , "expires_in_minutes": linkMinutes }
	return
}

// CheckKey is who key is on the project.
func CheckKey( ctx context.Context , base_url string , key string ) ( me *Me , err error ) {
	me = &Me{}
	if err = call( ctx , base_url , key , http.MethodGet , "/api/me" , nil , me ); err != nil { me = nil }
	return
}

// Health is the project's /api/health: nil when it is up.
func Health( ctx context.Context , base_url string ) ( err error ) {
	err = call( ctx , base_url , "" , http.MethodGet , "/api/health" , nil , nil )
	return
}

// CreateUser makes an ordinary account named name, and the link that signs it
// in.
func CreateUser( ctx context.Context , base_url string , admin_key string , name string ) ( link *Link , err error ) {
	link , err = CreateUserAs( ctx , base_url , admin_key , NewAccount{ Name: name , Role: "user" } )
	return
}

// NewAccount is an account to make: its role, and when it stops working
// (nil for never; only sent to an app with account_expiry).
type NewAccount struct {
	Name      string
	Role      string
	ExpiresAt *time.Time
}

// CreateUserAs is CreateUser with the role and expiry chosen.
func CreateUserAs( ctx context.Context , base_url string , admin_key string , account NewAccount ) ( link *Link , err error ) {
	body := linkRequest()
	body[ "display_name" ] , body[ "role" ] = account.Name , account.Role
	if account.ExpiresAt != nil { body[ "account_expires_at" ] = account.ExpiresAt.UTC().Format( time.RFC3339 ) }
	reply := &loginLink{}
	if err = call( ctx , base_url , admin_key , http.MethodPost , "/api/admin/users" , body , reply ); err != nil { return }
	link , err = toLink( base_url , reply )
	return
}

// ReissueLogin is a fresh link for an account, made with an admin key.
func ReissueLogin( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 ) ( link *Link , err error ) {
	reply := &loginLink{}
	path := "/api/admin/users/" + strconv.FormatUint( remote_user_id , 10 ) + "/reissue-login"
	if err = call( ctx , base_url , admin_key , http.MethodPost , path , linkRequest() , reply ); err != nil { return }
	reply.UserID = remote_user_id
	link , err = toLink( base_url , reply )
	return
}

// SetUserTheme puts theme on an account there and picks it (capability
// user_theme), made with an admin key.
func SetUserTheme( ctx context.Context , base_url string , admin_key string , remote_user_id uint64 , theme any ) ( err error ) {
	path := "/api/admin/users/" + strconv.FormatUint( remote_user_id , 10 ) + "/theme"
	err = call( ctx , base_url , admin_key , http.MethodPost , path , map[string]any{ "theme": theme } , nil )
	return
}

// OwnLoginLink is a link for the key's own account. The project refuses
// unless that key was allowed to make sign-in links.
func OwnLoginLink( ctx context.Context , base_url string , key string ) ( link *Link , err error ) {
	reply := &loginLink{}
	if err = call( ctx , base_url , key , http.MethodPost , "/api/account/login-link" , linkRequest() , reply ); err != nil { return }
	link , err = toLink( base_url , reply )
	return
}

// ListUsers is one page of the project's accounts, searched by q.
func ListUsers( ctx context.Context , base_url string , admin_key string , q string , offset int ) ( users []User , next *int , total int , err error ) {
	query := net_url.Values{}
	query.Set( "limit" , "200" )
	query.Set( "offset" , strconv.Itoa( offset ) )
	if q != "" { query.Set( "q" , q ) }
	page := &usersPage{}
	if err = call( ctx , base_url , admin_key , http.MethodGet , "/api/admin/users?" + query.Encode() , nil , page ); err != nil { return }
	users , next , total = page.Users , page.Next , page.Total
	if users == nil { users = []User{} }
	return
}

func toLink( base_url string , reply *loginLink ) ( link *Link , err error ) {
	if len( reply.LoginPath ) < 8 || reply.LoginPath[ :7 ] != "/login/" {
		err = &Error{ Status: http.StatusOK , Message: "the project's reply had no login link in it" }
		return
	}
	link = &Link{ URL: base_url + reply.LoginPath , RemoteUserID: reply.UserID }
	return
}
