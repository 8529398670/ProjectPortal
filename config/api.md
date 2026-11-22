# ProjectPortal API

Every route this server answers, apart from the pages at /login/*, /join/* and /docs/*. Generated from the code by scripts/api_docs.sh: an app's own routes get their docs the same way, and a test fails until they do.

This file is generated: do not edit it. Change the docs next to the handlers and run `scripts/api_docs.sh`.

## Contents

- [Calling the API](#calling-the-api)
- [Errors](#errors)
- [Account](#account)
- [Sign-in](#sign-in)
- [API keys](#api-keys)
- [Admin](#admin)
- [Invite links](#invite-links)
- [Owner](#owner)
- [Monitoring](#monitoring)
- [Catalog](#catalog)
- [Manage project](#manage-project)
- [Projects](#projects)
- [Schemas](#schemas)

## Calling the API

| Way in | How |
|---|---|
| Browser session | A browser session. Its POSTs carry csrf_token in the body. The cookie's name is server.cookie_name, with the __Host- prefix while server.secure_cookies is on. |
| API key | An API key, as Authorization: Bearer <id>.<secret>. |

## Errors

An error reply is a JSON object whose `error` says what went wrong.

| Status | Meaning | Body |
|---|---|---|
| 400 | The request was refused as it stands. | [`Error`](#error) |
| 401 | Not signed in. | [`Error`](#error) |
| 403 | Not allowed, or a bad csrf_token. | [`Error`](#error) |
| 404 | No such thing, or this feature is switched off. | [`Error`](#error) |
| 409 | The record has moved on: an undo too late, or a stale edit. | [`Error`](#error) |
| 424 | Another app this one calls refused or did not answer; error says what it said. | [`Error`](#error) |
| 429 | Too many requests. | [`Error`](#error) |
| 503 | In maintenance mode. Retry-After says when to try again. | [`Maintenance`](#maintenance) |

## Account

| Method | Path | Summary | Who may call |
|---|---|---|---|
| POST | [`/api/account/login-link`](#post-apiaccountlogin-link) | A login link for yourself | browser session or API key, anyone signed in |
| POST | [`/api/account/login-link/cancel`](#post-apiaccountlogin-linkcancel) | Stop your own login link working | browser session or API key, anyone signed in |
| POST | [`/api/account/password`](#post-apiaccountpassword) | Set or clear your password | browser session only, anyone signed in |
| POST | [`/api/account/pin`](#post-apiaccountpin) | Set or clear your PIN | browser session only, anyone signed in |
| POST | [`/api/account/rename`](#post-apiaccountrename) | Rename your account | browser session or API key, anyone signed in |
| GET | [`/api/account/sessions`](#get-apiaccountsessions) | Your signed-in devices | browser session or API key, anyone signed in |
| POST | [`/api/account/sessions/revoke-others`](#post-apiaccountsessionsrevoke-others) | Sign out every other device | browser session or API key, anyone signed in |
| POST | [`/api/account/sessions/{session_id}/label`](#post-apiaccountsessionssession_idlabel) | Name one of your devices | browser session or API key, anyone signed in |
| POST | [`/api/account/sessions/{session_id}/revoke`](#post-apiaccountsessionssession_idrevoke) | Sign out one of your devices | browser session or API key, anyone signed in |
| POST | [`/api/account/shortcuts`](#post-apiaccountshortcuts) | Change your keyboard shortcuts | browser session or API key, anyone signed in |
| POST | [`/api/account/theme`](#post-apiaccounttheme) | Pick your theme | browser session or API key, anyone signed in |
| POST | [`/api/account/themes`](#post-apiaccountthemes) | Replace your own themes | browser session or API key, anyone signed in |
| POST | [`/api/account/two-factor/confirm`](#post-apiaccounttwo-factorconfirm) | Turn two-factor on | browser session only, anyone signed in |
| POST | [`/api/account/two-factor/disable`](#post-apiaccounttwo-factordisable) | Turn two-factor off | browser session only, anyone signed in |
| POST | [`/api/account/two-factor/start`](#post-apiaccounttwo-factorstart) | Begin two-factor setup | browser session only, anyone signed in |
| POST | [`/api/bulk/account/sessions/revoke`](#post-apibulkaccountsessionsrevoke) | Sign out one of your devices, for many at once | browser session or API key, anyone signed in |
| GET | [`/api/language`](#get-apilanguage) | The UI's text | anyone |
| POST | [`/api/logout`](#post-apilogout) | Sign out | browser session or API key, anyone signed in |
| POST | [`/api/logs/events`](#post-apilogsevents) | Record what happened on the page | browser session or API key, anyone signed in |
| GET | [`/api/me`](#get-apime) | Who am I | browser session or API key, anyone signed in |
| GET | [`/api/team`](#get-apiteam) | List team | browser session or API key, anyone signed in |

### POST /api/account/login-link

A login link for yourself

A link that signs your own account in on a browser or another device. An API key may ask only if your account let it, from a browser session (POST /api/keys/{key_id}/sign-in); its link works once, for at most 15 minutes.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`create_my_login_link`](mcp.md#create_my_login_link) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_at` | string (date-time) or null | yes |  | null when it never expires. |
| `expires_in_seconds` | integer | yes |  | 0 when it never expires. |
| `login_path` | string | yes |  | Path of the login link. A credential: treat it like a password. |
| `max_uses` | integer | yes |  | 0 when unlimited. |
| `unlimited` | boolean | yes |  |  |

### POST /api/account/login-link/cancel

Stop your own login link working

Void your account's live login link without making a new one. Devices it already signed in stay signed in. Answers with removed: 0 or 1.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`cancel_my_login_link`](mcp.md#cancel_my_login_link) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `removed` | integer | yes |  | Links removed: 0 or 1. |

### POST /api/account/password

Set or clear your password

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | no |  |  |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `password_digest` | string | yes |  |  |
| `username` | string | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `has_password` | boolean | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/account/pin

Set or clear your PIN

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `pin` | string | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `has_pin` | boolean | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/account/rename

Rename your account

Change your own display name (1-80 characters). expect, when given, is the name you think it has now; if someone changed it meanwhile the rename is refused with the current name, so nothing is overwritten unseen.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`rename_account`](mcp.md#rename_account) |
| Errors | 400, 401, 403, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `display_name` | string | yes | 1 to 80 characters |  |
| `expect` | string or null | no |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### GET /api/account/sessions

Your signed-in devices

Where your account is signed in, a page at a time: id, when it started and was last seen, the browser, and the name you gave it. Through an API key no device is current. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_my_sessions`](mcp.md#list_my_sessions) |
| Errors | 400, 401, 403, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `sessions` | array of [`Session`](#session) | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/account/sessions/revoke-others

Sign out every other device

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`revoke_my_other_sessions`](mcp.md#revoke_my_other_sessions) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `removed` | integer | yes |  |  |

### POST /api/account/sessions/{session_id}/label

Name one of your devices

Give one of your devices a name (at most 80 characters; empty clears it). expect, when given, is the name you think it has now.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`label_my_session`](mcp.md#label_my_session) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `session_id` | path | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `label` | string | yes | at most 80 characters | At most 80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `label` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/account/sessions/{session_id}/revoke

Sign out one of your devices

End one of your signed-in sessions; that browser has to sign in again.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`revoke_my_session`](mcp.md#revoke_my_session) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `session_id` | path | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/account/shortcuts

Change your keyboard shortcuts

Replace your changed keyboard shortcuts: action name to binding, such as "search": "Mod+K" (Mod is Cmd on a Mac, Ctrl elsewhere). Send the whole set; an empty binding switches a key off, and {} puts every key back to its default.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_shortcuts`](mcp.md#set_shortcuts) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `shortcuts` | map of string | yes |  | Action name to binding, such as "search": "Mod+K". Mod is Cmd on a Mac and Ctrl elsewhere. Only changed shortcuts are listed; an empty binding switches that key off. Replaces the whole set; {} puts every key back to its default. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `shortcuts` | map of string | yes |  |  |

### POST /api/account/theme

Pick your theme

Switch to another theme from your list (or the defaults), by its id.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_theme`](mcp.md#set_theme) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `theme_id` | string | yes |  | A theme's id. Empty for the first in the list. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `theme_id` | string | yes |  |  |

### POST /api/account/themes

Replace your own themes

Replace your whole list of color themes (1 to 40) and pick the one you use. themes: null goes back to following the app's defaults. A theme is the Theme schema in the API reference (config/api.json): id, name, mode light|dark, accent, second, and optional colors as #rrggbb.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`set_themes`](mcp.md#set_themes) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | no |  | Which built-in themes this list has seen: the highest since in Theme.builtIn (static/js/theme.js) when it was saved. Built-in themes added after that join the list the browser shows. Left out: 0, the first ones. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `theme_id` | string | no |  | The theme you use. |
| `themes` | array of [`Theme`](#theme) or null | no |  | Your whole list, 1 to 40 themes. null goes back to following the app's defaults. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `theme_id` | string | yes |  |  |
| `themes` | array of [`Theme`](#theme) or null | yes |  | What you now see; null while the built-in themes apply. |
| `themes_builtins` | integer | yes |  | Which built-in themes that list has seen (see builtins on POST /api/account/themes). |
| `themes_custom` | boolean | yes |  |  |

### POST /api/account/two-factor/confirm

Turn two-factor on

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | yes |  |  |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `recovery_codes` | array of string | yes |  |  |

### POST /api/account/two-factor/disable

Turn two-factor off

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `code` | string | yes |  |  |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/account/two-factor/start

Begin two-factor setup

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `secret` | string | yes |  |  |

### POST /api/bulk/account/sessions/revoke

Sign out one of your devices, for many at once

Runs POST /api/account/sessions/{session_id}/revoke once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`revoke_my_session_bulk`](mcp.md#revoke_my_session_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### GET /api/language

The UI's text

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 429, 503 |

**Reply 200**: Success

An empty object.

### POST /api/logout

Sign out

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: Ends a browser session; a key has none to end. |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/logs/events

Record what happened on the page

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The browser's own activity reports (track.js), not an action. |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `events` | array of object | yes |  |  |
| `tab` | string | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `accepted` | integer | yes |  |  |

### GET /api/me

Who am I

The account this API key belongs to, the role it acts with (role -- never higher than the key's own scope), and the account's own role.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`whoami`](mcp.md#whoami) |
| Errors | 401, 429, 503 |

**Reply 200**: Success

[`Me`](#me).

### GET /api/team

List team

People with an active account, a page at a time: id and display name.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_team`](mcp.md#list_team) |
| Errors | 400, 401, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `members` | array of [`Member`](#member) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

## Sign-in

| Method | Path | Summary | Who may call |
|---|---|---|---|
| POST | [`/api/login`](#post-apilogin) | Redeem a login link | anyone |
| GET | [`/api/login/options`](#get-apiloginoptions) | What sign-in offers | anyone |
| POST | [`/api/login/password`](#post-apiloginpassword) | Sign in with a password | anyone |
| POST | [`/api/login/two-factor`](#post-apilogintwo-factor) | Finish a sign-in with a code | anyone |

### POST /api/login

Redeem a login link

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |

**Reply 200**: Success

[`SignIn`](#signin).

### GET /api/login/options

What sign-in offers

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `passwords` | boolean | yes |  |  |
| `salt` | string | no |  |  |
| `two_factor` | boolean | yes |  |  |

### POST /api/login/password

Sign in with a password

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `password_digest` | string | yes |  |  |
| `username` | string | yes |  |  |

**Reply 200**: Success

[`SignIn`](#signin).

### POST /api/login/two-factor

Finish a sign-in with a code

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `challenge` | string | yes |  |  |
| `code` | string | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

## API keys

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/admin/keys`](#get-apiadminkeys) | Every key in the app | browser session only, admins |
| POST | [`/api/admin/keys`](#post-apiadminkeys) | Create a key that acts as someone else | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/delete`](#post-apiadminkeyskey_iddelete) | Delete anyone's key | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/expiry`](#post-apiadminkeyskey_idexpiry) | Change when anyone's key expires | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/rename`](#post-apiadminkeyskey_idrename) | Rename anyone's key | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/restore`](#post-apiadminkeyskey_idrestore) | Undo deleting anyone's key | browser session only, admins |
| POST | [`/api/admin/keys/{key_id}/revoke`](#post-apiadminkeyskey_idrevoke) | Revoke anyone's key | browser session only, admins |
| GET | [`/api/keys`](#get-apikeys) | Your API keys | browser session only, anyone signed in |
| POST | [`/api/keys`](#post-apikeys) | Create an API key | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/delete`](#post-apikeyskey_iddelete) | Delete one of your keys | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/expiry`](#post-apikeyskey_idexpiry) | Change when one of your keys expires | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/rename`](#post-apikeyskey_idrename) | Rename one of your keys | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/restore`](#post-apikeyskey_idrestore) | Undo deleting one of your keys | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/revoke`](#post-apikeyskey_idrevoke) | Revoke one of your keys | browser session only, anyone signed in |
| POST | [`/api/keys/{key_id}/sign-in`](#post-apikeyskey_idsign-in) | Let one of your keys make sign-in links | browser session only, anyone signed in |

### GET /api/admin/keys

Every key in the app

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep keys whose name or owner contains this, in any case. |
| `live` | query | string | no | one of `true` | Leave out revoked and expired keys. |
| `sort` | query | string | no | one of `display_name`, `name`, `role`, `last_used_at`, `expires_at`, `created_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `keys` | array of [`APIKey`](#apikey) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/admin/keys

Create a key that acts as someone else

Any admin may do this for an ordinary user, and only the owner for another admin. Nobody may do it for the owner. The key records who made it and shows that to the person it acts as.

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer or null | no | 0 to 3650 | Absent or null: the configured default. 0: never expires. |
| `name` | string | yes | 1 to 80 characters |  |
| `role` | string | no | one of `admin`, `user`; default `user` |  |
| `user_id` | integer | yes | at least 1 |  |

**Reply 201**: Success

[`NewAPIKey`](#newapikey).

### POST /api/admin/keys/{key_id}/delete

Delete anyone's key

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/keys/{key_id}/expiry

Change when anyone's key expires

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer | yes | 0 to 3650 | Counted from now. 0 means it never expires. |

**Reply 200**: Success

[`APIKey`](#apikey).

### POST /api/admin/keys/{key_id}/rename

Rename anyone's key

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `name` | string | yes | 1 to 80 characters | 1-80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/admin/keys/{key_id}/restore

Undo deleting anyone's key

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`APIKey`](#apikey).

### POST /api/admin/keys/{key_id}/revoke

Revoke anyone's key

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/keys

Your API keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep keys whose name contains this, in any case. |
| `live` | query | string | no | one of `true` | Leave out revoked and expired keys. |
| `sort` | query | string | no | one of `name`, `role`, `last_used_at`, `expires_at`, `created_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `keys` | array of [`APIKey`](#apikey) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/keys

Create an API key

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer or null | no | 0 to 3650 | Absent or null: the configured default. 0: never expires. |
| `name` | string | yes | 1 to 80 characters |  |
| `role` | string | no | one of `admin`, `user`; default `user` |  |

**Reply 201**: Success

[`NewAPIKey`](#newapikey).

### POST /api/keys/{key_id}/delete

Delete one of your keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/keys/{key_id}/expiry

Change when one of your keys expires

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer | yes | 0 to 3650 | Counted from now. 0 means it never expires. |

**Reply 200**: Success

[`APIKey`](#apikey).

### POST /api/keys/{key_id}/rename

Rename one of your keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `name` | string | yes | 1 to 80 characters | 1-80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/keys/{key_id}/restore

Undo deleting one of your keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`APIKey`](#apikey).

### POST /api/keys/{key_id}/revoke

Revoke one of your keys

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/keys/{key_id}/sign-in

Let one of your keys make sign-in links

Off by default. When on, the key may ask POST /api/account/login-link for a link that signs your account in on a browser, which then can do everything you can, including what keys may not. The link works once, for at most 15 minutes.

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `key_id` | path | string | yes |  | The key's id, from GET /api/keys (GET /api/admin/keys for anyone's). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `may_sign_in` | boolean | yes |  | true lets the key make a login link for your account: one that works once, for at most 15 minutes. |

**Reply 200**: Success

[`APIKey`](#apikey).

## Admin

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/admin/theme-defaults`](#get-apiadmintheme-defaults) | Default themes | browser session or API key, admins |
| POST | [`/api/admin/theme-defaults`](#post-apiadmintheme-defaults) | Replace the default themes | browser session or API key, admins |
| GET | [`/api/admin/users`](#get-apiadminusers) | List users | browser session or API key, admins |
| POST | [`/api/admin/users`](#post-apiadminusers) | Create an account | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/cancel-login-link`](#post-apiadminusersuser_idcancel-login-link) | Stop an account's login link working | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/delete`](#post-apiadminusersuser_iddelete) | Delete an account | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/disabled`](#post-apiadminusersuser_iddisabled) | Disable or enable a user | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/reissue-login`](#post-apiadminusersuser_idreissue-login) | A new login link for an account | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/rename`](#post-apiadminusersuser_idrename) | Rename someone | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/reset-two-factor`](#post-apiadminusersuser_idreset-two-factor) | Turn off an account's two-factor | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/revoke-access`](#post-apiadminusersuser_idrevoke-access) | Revoke a user's access | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/role`](#post-apiadminusersuser_idrole) | Change an account's role | browser session or API key, admins |
| GET | [`/api/admin/users/{user_id}/sessions`](#get-apiadminusersuser_idsessions) | An account's signed-in devices | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/sessions/{session_id}/revoke`](#post-apiadminusersuser_idsessionssession_idrevoke) | Sign out one of an account's devices | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/cancel-login-link`](#post-apibulkadminuserscancel-login-link) | Stop an account's login link working, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/delete`](#post-apibulkadminusersdelete) | Delete an account, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/disabled`](#post-apibulkadminusersdisabled) | Disable or enable a user, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/reset-two-factor`](#post-apibulkadminusersreset-two-factor) | Turn off an account's two-factor, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/revoke-access`](#post-apibulkadminusersrevoke-access) | Revoke a user's access, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/role`](#post-apibulkadminusersrole) | Change an account's role, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/users/sessions/revoke`](#post-apibulkadminuserssessionsrevoke) | Sign out one of an account's devices, for many at once | browser session or API key, admins |

### GET /api/admin/theme-defaults

Default themes

The theme list everyone without their own list sees. themes is null while the built-in ones apply.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_get_theme_defaults`](mcp.md#admin_get_theme_defaults) |
| Errors | 401, 403, 404, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | yes |  | Which built-in themes that list has seen (see builtins on POST /api/account/themes). |
| `themes` | array of [`Theme`](#theme) or null | yes |  | null while the built-in themes apply. |

### POST /api/admin/theme-defaults

Replace the default themes

Replace the theme list everyone without their own list sees, 1 to 40 themes in the Theme schema of the API reference. null puts the built-in themes back.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_theme_defaults`](mcp.md#admin_set_theme_defaults) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | no |  | Which built-in themes this list has seen: the highest since in Theme.builtIn (static/js/theme.js) when it was saved. Built-in themes added after that join the list the browser shows. Left out: 0, the first ones. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `themes` | array of [`Theme`](#theme) or null | no |  | 1 to 40 themes. null goes back to the built-in ones. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | yes |  | Which built-in themes that list has seen (see builtins on POST /api/account/themes). |
| `ok` | boolean | yes |  |  |
| `themes` | array of [`Theme`](#theme) or null | yes |  |  |

### GET /api/admin/users

List users

Accounts, a page at a time, including disabled ones, with each one's role and which is the owner. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_users`](mcp.md#admin_list_users) |
| Errors | 400, 401, 403, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep accounts whose name or username contains this, in any case. |
| `sort` | query | string | no | one of `display_name`, `role`, `disabled`, `created_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |
| `users` | array of [`User`](#user) | yes |  |  |

### POST /api/admin/users

Create an account

Make a new account and its first login link, to hand to the person. role is admin or user. Answers with user_id and the link.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_create_user`](mcp.md#admin_create_user) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `display_name` | string | yes | 1 to 80 characters |  |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |
| `role` | string | yes | one of `admin`, `user` |  |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_at` | string (date-time) or null | yes |  | null when it never expires. |
| `expires_in_seconds` | integer | yes |  | 0 when it never expires. |
| `login_path` | string | yes |  | Path of the login link. A credential: treat it like a password. |
| `max_uses` | integer | yes |  | 0 when unlimited. |
| `unlimited` | boolean | yes |  |  |
| `user_id` | integer | yes |  |  |

### POST /api/admin/users/{user_id}/cancel-login-link

Stop an account's login link working

Void an account's live login link without making a new one: a link that went somewhere it should not have.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_cancel_login_link`](mcp.md#admin_cancel_login_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `removed` | integer | yes |  | Links removed: 0 or 1. |

### POST /api/admin/users/{user_id}/delete

Delete an account

Remove an account for good, with its sessions, login link and API keys.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_user`](mcp.md#admin_delete_user) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `removed` | [`Revoked`](#revoked) | yes |  |  |

### POST /api/admin/users/{user_id}/disabled

Disable or enable a user

Switch an account off (disabled: true) or back on (false). Disabling signs it out everywhere and voids its live login link; its API keys stop working while it is off. Reversible. The owner and your own account cannot be disabled.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_user_disabled`](mcp.md#admin_set_user_disabled) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `disabled` | boolean | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/reissue-login

A new login link for an account

For someone who lost their device or their link: a fresh login link for their account.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_reissue_login`](mcp.md#admin_reissue_login) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_at` | string (date-time) or null | yes |  | null when it never expires. |
| `expires_in_seconds` | integer | yes |  | 0 when it never expires. |
| `login_path` | string | yes |  | Path of the login link. A credential: treat it like a password. |
| `max_uses` | integer | yes |  | 0 when unlimited. |
| `unlimited` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/rename

Rename someone

Change another account's display name (1-80 characters). Only the owner renames an admin. expect, when given, is the name you think it has now; if it changed meanwhile the rename is refused with the current name.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_rename_user`](mcp.md#admin_rename_user) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `display_name` | string | yes | 1 to 80 characters | 1-80 characters. |
| `expect` | string or null | no |  | The name you think they have. Answers 409 if it is another. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/reset-two-factor

Turn off an account's two-factor

For someone who lost both their phone and their recovery codes: switches their two-factor sign-in off so a login link lets them in again. Only the owner resets an admin's; the owner's is reset from a shell.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_reset_two_factor`](mcp.md#admin_reset_two_factor) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/revoke-access

Revoke a user's access

For an account that may be compromised: ends every session and voids its live login link and every API key it holds, at once.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_access`](mcp.md#admin_revoke_access) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `revoked` | [`Revoked`](#revoked) | yes |  |  |

### POST /api/admin/users/{user_id}/role

Change an account's role

Make an account an admin or an ordinary user. Only the owner takes an admin's rights away; the owner is always an admin, and nobody changes their own role. Becoming an admin ends the account's sessions and live login link.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_user_role`](mcp.md#admin_set_user_role) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `role` | string | yes | one of `admin`, `user` |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `role` | string | yes |  |  |

### GET /api/admin/users/{user_id}/sessions

An account's signed-in devices

Where an account is signed in: every session, with when it started and was last seen. Addresses are shown to the owner only, and only the owner sees the owner's.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_user_sessions`](mcp.md#admin_list_user_sessions) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `sessions` | array of [`Session`](#session) | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/admin/users/{user_id}/sessions/{session_id}/revoke

Sign out one of an account's devices

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_user_session`](mcp.md#admin_revoke_user_session) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `session_id` | path | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/bulk/admin/users/cancel-login-link

Stop an account's login link working, for many at once

Runs POST /api/admin/users/{user_id}/cancel-login-link once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_cancel_login_link_bulk`](mcp.md#admin_cancel_login_link_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/delete

Delete an account, for many at once

Runs POST /api/admin/users/{user_id}/delete once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_user_bulk`](mcp.md#admin_delete_user_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/disabled

Disable or enable a user, for many at once

Runs POST /api/admin/users/{user_id}/disabled once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_user_disabled_bulk`](mcp.md#admin_set_user_disabled_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].disabled` | boolean | no |  |  |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.disabled` | boolean | no |  |  |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/reset-two-factor

Turn off an account's two-factor, for many at once

Runs POST /api/admin/users/{user_id}/reset-two-factor once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_reset_two_factor_bulk`](mcp.md#admin_reset_two_factor_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/revoke-access

Revoke a user's access, for many at once

Runs POST /api/admin/users/{user_id}/revoke-access once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_access_bulk`](mcp.md#admin_revoke_access_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/role

Change an account's role, for many at once

Runs POST /api/admin/users/{user_id}/role once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_user_role_bulk`](mcp.md#admin_set_user_role_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].role` | string | no | one of `admin`, `user` |  |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.role` | string | no | one of `admin`, `user` |  |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/users/sessions/revoke

Sign out one of an account's devices, for many at once

Runs POST /api/admin/users/{user_id}/sessions/{session_id}/revoke once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_user_session_bulk`](mcp.md#admin_revoke_user_session_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

## Invite links

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/admin/invite-links`](#get-apiadmininvite-links) | List invite links | browser session or API key, admins |
| POST | [`/api/admin/invite-links`](#post-apiadmininvite-links) | Create an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/delete`](#post-apiadmininvite-linkslink_iddelete) | Delete an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/limits`](#post-apiadmininvite-linkslink_idlimits) | Change an invite link's places and expiry | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/rename`](#post-apiadmininvite-linkslink_idrename) | Rename an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/restore`](#post-apiadmininvite-linkslink_idrestore) | Undo deleting an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/revoke`](#post-apiadmininvite-linkslink_idrevoke) | Switch off an invite link | browser session or API key, admins |
| POST | [`/api/admin/invite-links/{link_id}/switch-on`](#post-apiadmininvite-linkslink_idswitch-on) | Switch an invite link back on | browser session or API key, admins |
| POST | [`/api/bulk/admin/invite-links/delete`](#post-apibulkadmininvite-linksdelete) | Delete an invite link, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/invite-links/restore`](#post-apibulkadmininvite-linksrestore) | Undo deleting an invite link, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/invite-links/revoke`](#post-apibulkadmininvite-linksrevoke) | Switch off an invite link, for many at once | browser session or API key, admins |
| POST | [`/api/bulk/admin/invite-links/switch-on`](#post-apibulkadmininvite-linksswitch-on) | Switch an invite link back on, for many at once | browser session or API key, admins |
| POST | [`/api/join/claim`](#post-apijoinclaim) | Sign in as an existing account | anyone |
| POST | [`/api/join/create`](#post-apijoincreate) | Join as a new account | anyone |
| POST | [`/api/join/info`](#post-apijoininfo) | What an invite link is | anyone |
| POST | [`/api/join/match`](#post-apijoinmatch) | Accounts matching a name | anyone |

### GET /api/admin/invite-links

List invite links

Group invite links, a page at a time: label, places used and left, expiry, and whether each still works. The links themselves are never shown again after creation. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_invite_links`](mcp.md#admin_list_invite_links) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep links whose label or maker contains this, in any case. |
| `sort` | query | string | no | one of `label`, `used`, `expires_at` | Field to sort the whole list by. Rows without it go last. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `invite_links` | array of [`InviteLink`](#invitelink) | yes |  |  |
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |

### POST /api/admin/invite-links

Create an invite link

One link for a group: each person who opens it joins as a new user, up to max_uses people, for days days.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_create_invite_link`](mcp.md#admin_create_invite_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `days` | integer | yes | 1 to 365 |  |
| `label` | string | yes | 1 to 80 characters |  |
| `max_uses` | integer | yes | 1 to 1000 |  |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_by` | integer | yes |  |  |
| `created_by_name` | string | no |  |  |
| `expires_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `join_path` | string | yes |  | The link to hand out, shown in this response only. A credential: treat it like a password. |
| `label` | string | yes |  |  |
| `live` | boolean | yes |  |  |
| `max_uses` | integer | yes |  |  |
| `remaining` | integer | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `used` | integer | yes |  |  |

### POST /api/admin/invite-links/{link_id}/delete

Delete an invite link

Take an invite link off the list and stop it working. The accounts it made stay.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_invite_link`](mcp.md#admin_delete_invite_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/invite-links/{link_id}/limits

Change an invite link's places and expiry

Set how many people the link may take in (never fewer than have joined) and, when days is given, a new expiry counted from now. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_invite_link_limits`](mcp.md#admin_set_invite_link_limits) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `days` | integer or null | no | 1 to 365 | Counted from now. Leave out to keep the expiry. |
| `max_uses` | integer | yes | 1 to 1000 | Not fewer than have joined. |

**Reply 200**: Success

[`InviteLink`](#invitelink).

### POST /api/admin/invite-links/{link_id}/rename

Rename an invite link

Change the label people see on an invite link's join page (1-80 characters). expect, when given, is the label you think it has now.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_rename_invite_link`](mcp.md#admin_rename_invite_link) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expect` | string or null | no |  | The label you think it has. Answers 409 if it has another. |
| `label` | string | yes | 1 to 80 characters | 1-80 characters. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `label` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/admin/invite-links/{link_id}/restore

Undo deleting an invite link

Bring back an invite link deleted in the last few minutes, working as before. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_restore_invite_link`](mcp.md#admin_restore_invite_link) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`InviteLink`](#invitelink).

### POST /api/admin/invite-links/{link_id}/revoke

Switch off an invite link

Stop an invite link from letting anyone else join or sign in through it. Accounts it already made keep working. Cannot be switched back on.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_invite_link`](mcp.md#admin_revoke_invite_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/invite-links/{link_id}/switch-on

Switch an invite link back on

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_switch_on_invite_link`](mcp.md#admin_switch_on_invite_link) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `link_id` | path | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`InviteLink`](#invitelink).

### POST /api/bulk/admin/invite-links/delete

Delete an invite link, for many at once

Runs POST /api/admin/invite-links/{link_id}/delete once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_invite_link_bulk`](mcp.md#admin_delete_invite_link_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/invite-links/restore

Undo deleting an invite link, for many at once

Runs POST /api/admin/invite-links/{link_id}/restore once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_restore_invite_link_bulk`](mcp.md#admin_restore_invite_link_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/invite-links/revoke

Switch off an invite link, for many at once

Runs POST /api/admin/invite-links/{link_id}/revoke once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_invite_link_bulk`](mcp.md#admin_revoke_invite_link_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/bulk/admin/invite-links/switch-on

Switch an invite link back on, for many at once

Runs POST /api/admin/invite-links/{link_id}/switch-on once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_switch_on_invite_link_bulk`](mcp.md#admin_switch_on_invite_link_bulk) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

### POST /api/join/claim

Sign in as an existing account

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |
| `display_name` | string | no |  |  |
| `pin` | string | no |  |  |
| `user_id` | integer | no |  |  |

**Reply 200**: Success

[`Joined`](#joined).

### POST /api/join/create

Join as a new account

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |
| `display_name` | string | no |  |  |
| `pin` | string | no |  |  |
| `user_id` | integer | no |  |  |

**Reply 201**: Success

[`Joined`](#joined).

### POST /api/join/info

What an invite link is

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |
| `display_name` | string | no |  |  |
| `pin` | string | no |  |  |
| `user_id` | integer | no |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `label` | string | yes |  |  |
| `remaining` | integer | yes |  |  |

### POST /api/join/match

Accounts matching a name

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `credential` | string | yes |  |  |
| `display_name` | string | no |  |  |
| `pin` | string | no |  |  |
| `user_id` | integer | no |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `can_create` | boolean | yes |  |  |
| `matches` | array of object | yes |  |  |
| `matches[].display_name` | string | yes |  |  |
| `matches[].has_pin` | boolean | yes |  |  |
| `matches[].id` | integer | yes |  |  |

## Owner

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/owner/jobs`](#get-apiownerjobs) | Background jobs | browser session only, the owner |
| POST | [`/api/owner/jobs/{name}/run`](#post-apiownerjobsnamerun) | Run a job now | browser session only, the owner |
| GET | [`/api/owner/logs/days`](#get-apiownerlogsdays) | The days on record | browser session only, the owner |
| GET | [`/api/owner/logs/days/{date}`](#get-apiownerlogsdaysdate) | One day's entries | browser session only, the owner |
| POST | [`/api/owner/logs/days/{date}/delete`](#post-apiownerlogsdaysdatedelete) | Delete one day | browser session only, the owner |
| GET | [`/api/owner/logs/days/{date}/download`](#get-apiownerlogsdaysdatedownload) | One day's entries as a file | browser session only, the owner |
| GET | [`/api/owner/logs/ip-names`](#get-apiownerlogsip-names) | What the owner calls addresses | browser session only, the owner |
| POST | [`/api/owner/logs/ip-names`](#post-apiownerlogsip-names) | Name an address or a range, or rename one | browser session only, the owner |
| POST | [`/api/owner/logs/ip-names/delete`](#post-apiownerlogsip-namesdelete) | Take a name off an address or a range | browser session only, the owner |
| POST | [`/api/owner/logs/prune`](#post-apiownerlogsprune) | Apply the retention rules now | browser session only, the owner |
| GET | [`/api/owner/logs/stream`](#get-apiownerlogsstream) | New log entries as they happen | browser session only, the owner |
| GET | [`/api/owner/settings`](#get-apiownersettings) | config.yaml, as a form | browser session only, the owner |
| POST | [`/api/owner/settings`](#post-apiownersettings) | Save settings | browser session only, the owner |
| POST | [`/api/owner/users/{user_id}/password-access`](#post-apiownerusersuser_idpassword-access) | Allow or block an account's password | browser session only, the owner |

### GET /api/owner/jobs

Background jobs

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 404, 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `jobs` | array of [`Job`](#job) | yes |  |  |

### POST /api/owner/jobs/{name}/run

Run a job now

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `name` | path | string | yes |  | The job's name, from GET /api/owner/jobs. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 202**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/owner/logs/days

The days on record

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `q` | query | string | no |  | The Logs screen's search. |

**Reply 200**: Success

[`LogDays`](#logdays).

### GET /api/owner/logs/days/{date}

One day's entries

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `date` | path | string | yes |  | The day's date, from GET /api/owner/logs/days. |
| `q` | query | string | no |  | The Logs screen's search. |
| `before` | query | integer or null | no | at least 0 |  |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |

**Reply 200**: Success

[`LogDay`](#logday).

### POST /api/owner/logs/days/{date}/delete

Delete one day

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `date` | path | string | yes |  | The day's date, from GET /api/owner/logs/days. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `deleted` | boolean | yes |  |  |

### GET /api/owner/logs/days/{date}/download

One day's entries as a file

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `date` | path | string | yes |  | The day's date, from GET /api/owner/logs/days. |

**Reply 200**: Success

Text, as `application/x-ndjson`.

### GET /api/owner/logs/ip-names

What the owner calls addresses

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 403, 404, 429, 503 |

**Reply 200**: Success

[`IPNames`](#ipnames).

### POST /api/owner/logs/ip-names

Name an address or a range, or rename one

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `name` | string | yes | 1 to 80 characters |  |
| `target` | string | yes | e.g. `10.0.0.0/24` | An address, or a range such as 10.0.0.0/24. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `name` | string | yes |  |  |
| `target` | string | yes |  |  |

### POST /api/owner/logs/ip-names/delete

Take a name off an address or a range

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `target` | string | yes | e.g. `10.0.0.0/24` | An address, or a range such as 10.0.0.0/24. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `deleted` | boolean | yes |  |  |

### POST /api/owner/logs/prune

Apply the retention rules now

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

[`PruneResult`](#pruneresult).

### GET /api/owner/logs/stream

New log entries as they happen

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `q` | query | string | no |  | The Logs screen's search. |

**Reply 200**: Success

Text, as `text/event-stream`.

### GET /api/owner/settings

config.yaml, as a form

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 404, 429, 503 |

**Reply 200**: Success

[`SettingsSnapshot`](#settingssnapshot).

### POST /api/owner/settings

Save settings

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 409, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `changes` | object | yes |  | Setting paths and their new values. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `version` | string | yes |  |  |

**Reply 200**: Success

[`SettingsResult`](#settingsresult).

### POST /api/owner/users/{user_id}/password-access

Allow or block an account's password

|  |  |
|---|---|
| Who may call | browser session only, the owner |
| Errors | 400, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `allowed` | boolean | yes |  |  |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `password_blocked` | boolean | yes |  |  |

## Monitoring

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/api.json`](#get-apiapijson) | This document | browser session or API key, anyone signed in |
| GET | [`/api/api.md`](#get-apiapimd) | The API reference as Markdown tables for people | browser session or API key, anyone signed in |
| GET | [`/api/health`](#get-apihealth) | Is the process up | anyone |
| GET | [`/api/mcp.json`](#get-apimcpjson) | The MCP tools: every tool /mcp has, its arguments, and the route it stands for | browser session or API key, anyone signed in |
| GET | [`/api/mcp.md`](#get-apimcpmd) | The MCP tools as Markdown tables for people | browser session or API key, anyone signed in |
| GET | [`/api/ready`](#get-apiready) | Can the server do real work | anyone |
| GET | [`/mcp`](#get-mcp) | Refused: this server has no stream | anyone |
| POST | [`/mcp`](#post-mcp) | Model Context Protocol, at the key's role | API key only, anyone signed in |
| DELETE | [`/mcp`](#delete-mcp) | Refused: this server keeps no sessions | anyone |

### GET /api/api.json

This document

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The API reference itself; GET /api/mcp.json is the MCP one. |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

An empty object.

### GET /api/api.md

The API reference as Markdown tables for people

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The API reference itself, for people; GET /api/api.json is the same as JSON. |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

Text, as `text/markdown`.

### GET /api/health

Is the process up

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/mcp.json

The MCP tools: every tool /mcp has, its arguments, and the route it stands for

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The MCP reference itself; tools/list gives the same list. |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

An empty object.

### GET /api/mcp.md

The MCP tools as Markdown tables for people

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | None, HTTP only: The MCP reference itself, for people; tools/list gives the same list. |
| Errors | 401, 404, 429, 503 |

**Reply 200**: Success

Text, as `text/markdown`.

### GET /api/ready

Can the server do real work

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 429, 503 |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /mcp

Refused: this server has no stream

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 404, 429, 503 |

**Reply 405**: Success

### POST /mcp

Model Context Protocol, at the key's role

|  |  |
|---|---|
| Who may call | API key only, anyone signed in |
| MCP tool | None, HTTP only: The MCP endpoint itself. |
| Errors | 400, 401, 403, 404, 429, 503 |

**Request body**

An empty object.

**Reply 200**: Success

An empty object.

### DELETE /mcp

Refused: this server keeps no sessions

|  |  |
|---|---|
| Who may call | anyone |
| Errors | 404, 429, 503 |

**Reply 405**: Success

## Catalog

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/admin/projects`](#get-apiadminprojects) | The catalog | browser session or API key, admins |
| POST | [`/api/admin/projects`](#post-apiadminprojects) | Add a project | browser session only, admins |
| POST | [`/api/admin/projects/order`](#post-apiadminprojectsorder) | Order the catalog | browser session or API key, admins |
| GET | [`/api/admin/projects/{project_id}/access`](#get-apiadminprojectsproject_idaccess) | Who may use a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/access/{user_id}`](#post-apiadminprojectsproject_idaccessuser_id) | Let someone use a project, or not | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/check`](#post-apiadminprojectsproject_idcheck) | Check a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/delete`](#post-apiadminprojectsproject_iddelete) | Delete a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/edit`](#post-apiadminprojectsproject_idedit) | Change a project | browser session only, admins |
| POST | [`/api/admin/projects/{project_id}/grant/{user_id}`](#post-apiadminprojectsproject_idgrantuser_id) | Put someone on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/key`](#post-apiadminprojectsproject_idkey) | Set a project's admin key | browser session only, admins |
| POST | [`/api/admin/projects/{project_id}/link/{user_id}`](#post-apiadminprojectsproject_idlinkuser_id) | Match someone to their account on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/members/{user_id}/remove`](#post-apiadminprojectsproject_idmembersuser_idremove) | Take someone off a project here | browser session or API key, admins |
| GET | [`/api/admin/projects/{project_id}/remote-users`](#get-apiadminprojectsproject_idremote-users) | A project's own accounts | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/restore`](#post-apiadminprojectsproject_idrestore) | Undo deleting a project | browser session or API key, admins |
| POST | [`/api/bulk/admin/projects/access`](#post-apibulkadminprojectsaccess) | Let someone use a project, or not, for many at once | browser session or API key, admins |

### GET /api/admin/projects

The catalog

Every project in the portal, a page at a time, with whether its admin key works, whether it is up, how many people are on it, and its GitHub and TodoBoard links.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_catalog`](mcp.md#admin_list_catalog) |
| Errors | 400, 401, 403, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep projects whose name or address contains this, in any case. |
| `sort` | query | string | no | one of `name`, `base_url`, `health`, `members`, `position` | Field to sort by. Default the catalog's order. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `projects` | array of [`CatalogProject`](#catalogproject) | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/admin/projects

Add a project

Add a project to the catalog. With admin_key (an admin API key there), people can join it straight away; the key is checked first and never shown again.

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 424, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `admin_key` | string | no | at most 400 characters | An admin API key on that project. Checked before it is kept. |
| `base_url` | string | yes | 8 to 300 characters | Where the project is served, e.g. https://todo.example.com: what people's browsers open. |
| `board_share_url` | string | no | at most 500 characters | A view-only TodoBoard share link (/s/...), handed to everyone who may see the project. |
| `board_url` | string | no | at most 300 characters | Its TodoBoard board, e.g. https://todo.example.com/b/app. Only admins see it. |
| `color` | string | no | at most 7 characters | #rrggbb. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `description` | string | no | at most 500 characters |  |
| `github_url` | string | no | at most 300 characters | Where its code is, e.g. https://github.com/me/app. Only admins see it. |
| `icon` | string | no | at most 40 characters | A Bootstrap Icons name without bi-, e.g. kanban. |
| `listed` | boolean | yes |  | Whether people see it. Off, only admins do. |
| `name` | string | yes | 1 to 80 characters |  |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project` | [`CatalogProject`](#catalogproject) | yes |  |  |

### POST /api/admin/projects/order

Order the catalog

Put projects in this order everywhere they are listed. Ids left out follow.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_order_catalog`](mcp.md#admin_order_catalog) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `project_ids` | array of integer | yes |  | Project ids in the order people should see them. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/admin/projects/{project_id}/access

Who may use a project

Every account here, a page at a time, with whether it may use the project and whether it is on it.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_project_access`](mcp.md#admin_project_access) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep people whose name contains this, in any case. |
| `sort` | query | string | no | one of `display_name`, `allowed`, `member` | Field to sort by. |
| `dir` | query | string | no | one of `asc`, `desc` | Default asc. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `rows` | array of [`ProjectAccess`](#projectaccess) | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/admin/projects/{project_id}/access/{user_id}

Let someone use a project, or not

allowed false keeps a listed project from one person: it leaves their list and the portal stops signing them in. Their account on the project itself stays. Admins always see everything.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_project_access`](mcp.md#admin_set_project_access) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `user_id` | path | integer | yes | at least 1 | The account's id here, from admin_list_users. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `allowed` | boolean | yes |  | false keeps the project from them: it leaves their list, and the portal will not sign them into it. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `on_project` | string | no | one of `keep`, `disable`, `revoke` | With allowed false, what happens to an account they have on the project: keep it (the default), disable it, or end its sessions, links and keys. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `allowed` | boolean | yes |  |  |
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/check

Check a project

Ask the project now whether it is up and whether the portal's admin key still works there. The project-health job does this every 5 minutes.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_check_project`](mcp.md#admin_check_project) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project` | [`CatalogProject`](#catalogproject) | yes |  |  |

### POST /api/admin/projects/{project_id}/delete

Delete a project

Take a project out of the portal, with its key and everyone's membership of it. Accounts on the project itself stay. admin_restore_project undoes it within 10 minutes.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_project`](mcp.md#admin_delete_project) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/edit

Change a project

Change a project's name, description, address, icon, color, links or whether it is listed. A new address drops the admin key and everyone's own keys for the old one: give the key again.

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `base_url` | string | yes | 8 to 300 characters | Where the project is served, e.g. https://todo.example.com: what people's browsers open. |
| `board_share_url` | string | no | at most 500 characters | A view-only TodoBoard share link (/s/...), handed to everyone who may see the project. |
| `board_url` | string | no | at most 300 characters | Its TodoBoard board, e.g. https://todo.example.com/b/app. Only admins see it. |
| `color` | string | no | at most 7 characters | #rrggbb. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `description` | string | no | at most 500 characters |  |
| `github_url` | string | no | at most 300 characters | Where its code is, e.g. https://github.com/me/app. Only admins see it. |
| `icon` | string | no | at most 40 characters | A Bootstrap Icons name without bi-, e.g. kanban. |
| `listed` | boolean | yes |  | Whether people see it. Off, only admins do. |
| `name` | string | yes | 1 to 80 characters |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `keys_dropped` | boolean | yes |  | The address changed, so the admin key and everyone's own keys for the old one were dropped. |
| `project` | [`CatalogProject`](#catalogproject) | yes |  |  |

### POST /api/admin/projects/{project_id}/grant/{user_id}

Put someone on a project

Make someone an account on the project now, as if they had joined, optionally view only or expiring. They sign in from their own list; no link is handed to you.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_grant_project`](mcp.md#admin_grant_project) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `user_id` | path | integer | yes | at least 1 | The account's id here. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer | no | 0 to 3650 | Their account there stops working this many days from now. 0 or left out: never. |
| `role` | string | no | one of `user`, `viewer` | Their role on the project. Default user; viewer (view only) needs an app that has it. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/key

Set a project's admin key

Replace the admin API key the portal uses on a project. It must be an admin key there. Revoking the old one on the project signs out everyone it signed in.

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `admin_key` | string | yes | 8 to 400 characters | An admin API key on that project. |
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project` | [`CatalogProject`](#catalogproject) | yes |  |  |

### POST /api/admin/projects/{project_id}/link/{user_id}

Match someone to their account on a project

Say which account on the project is theirs. Opening the project then signs them into it. The project's owner, and its admins unless the portal's key is the owner's, cannot be signed in this way: they attach their own key instead.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_link_remote_user`](mcp.md#admin_link_remote_user) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `user_id` | path | integer | yes | at least 1 | The account's id here. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `remote_user_id` | integer | yes | at least 1 | Their account's id on the project, from admin_list_remote_users. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/members/{user_id}/remove

Take someone off a project here

Forget someone's account on the project, and their own key for it. on_project says what happens to the account there: keep (default), disable, or revoke its sessions, links and keys.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_remove_project_member`](mcp.md#admin_remove_project_member) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `user_id` | path | integer | yes | at least 1 | The account's id here. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `on_project` | string | no | one of `keep`, `disable`, `revoke` | What happens to their account on the project itself: keep it (the default), disable it, or end its sessions, links and keys. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### GET /api/admin/projects/{project_id}/remote-users

A project's own accounts

The accounts on the project itself, as its admin key sees them, 200 at a time.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_remote_users`](mcp.md#admin_list_remote_users) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `q` | query | string | no |  | Keep accounts whose name contains this. |
| `offset` | query | integer | no | at least 0 | From the previous page's next. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `total` | integer | yes |  |  |
| `users` | array of [`RemoteUser`](#remoteuser) | yes |  |  |

### POST /api/admin/projects/{project_id}/restore

Undo deleting a project

Put back a project deleted in the last 10 minutes.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_restore_project`](mcp.md#admin_restore_project) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/bulk/admin/projects/access

Let someone use a project, or not, for many at once

Runs POST /api/admin/projects/{project_id}/access/{user_id} once per item: items holds each one's id and fields, shared the fields they have in common (an item's own win). Each runs as that route would, in order, and one failing does not stop the rest; the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_project_access_bulk`](mcp.md#admin_set_project_access_bulk) |
| Errors | 400, 401, 403, 429, 503 |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].allowed` | boolean | no |  | false keeps the project from them: it leaves their list, and the portal will not sign them into it. |
| `items[].on_project` | string | no | one of `keep`, `disable`, `revoke` | With allowed false, what happens to an account they have on the project: keep it (the default), disable it, or end its sessions, links and keys. |
| `items[].project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `items[].user_id` | integer | yes | at least 1 | The account's id here, from admin_list_users. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.allowed` | boolean | no |  | false keeps the project from them: it leaves their list, and the portal will not sign them into it. |
| `shared.on_project` | string | no | one of `keep`, `disable`, `revoke` | With allowed false, what happens to an account they have on the project: keep it (the default), disable it, or end its sessions, links and keys. |

**Reply 200**: Success

[`BulkReply`](#bulkreply).

## Manage project

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/admin/projects/{project_id}/invite-links`](#get-apiadminprojectsproject_idinvite-links) | A project's invite links | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/invite-links`](#post-apiadminprojectsproject_idinvite-links) | Make an invite link on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/invite-links/{link_id}/revoke`](#post-apiadminprojectsproject_idinvite-linkslink_idrevoke) | Stop an invite link on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/keys/{key_id}/revoke`](#post-apiadminprojectsproject_idkeyskey_idrevoke) | Revoke an API key on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/remote-users`](#post-apiadminprojectsproject_idremote-users) | Make an account on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/remote-users/{remote_user_id}/delete`](#post-apiadminprojectsproject_idremote-usersremote_user_iddelete) | Delete an account on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/remote-users/{remote_user_id}/disabled`](#post-apiadminprojectsproject_idremote-usersremote_user_iddisabled) | Disable or enable an account on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/remote-users/{remote_user_id}/expiry`](#post-apiadminprojectsproject_idremote-usersremote_user_idexpiry) | Set when an account on a project expires | browser session or API key, admins |
| GET | [`/api/admin/projects/{project_id}/remote-users/{remote_user_id}/keys`](#get-apiadminprojectsproject_idremote-usersremote_user_idkeys) | An account's API keys on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/remote-users/{remote_user_id}/keys`](#post-apiadminprojectsproject_idremote-usersremote_user_idkeys) | Make an API key on a project | browser session only, admins |
| POST | [`/api/admin/projects/{project_id}/remote-users/{remote_user_id}/login-link`](#post-apiadminprojectsproject_idremote-usersremote_user_idlogin-link) | A login link for an account on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/remote-users/{remote_user_id}/revoke-access`](#post-apiadminprojectsproject_idremote-usersremote_user_idrevoke-access) | Sign an account out everywhere on a project | browser session or API key, admins |
| POST | [`/api/admin/projects/{project_id}/remote-users/{remote_user_id}/role`](#post-apiadminprojectsproject_idremote-usersremote_user_idrole) | Change an account's role on a project | browser session or API key, admins |
| POST | [`/api/admin/users/{user_id}/project-accounts`](#post-apiadminusersuser_idproject-accounts) | Disable someone's accounts on every project | browser session or API key, admins |

### GET /api/admin/projects/{project_id}/invite-links

A project's invite links

The invite links on the project itself. Needs an app with invite links switched on.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_remote_invite_links`](mcp.md#admin_list_remote_invite_links) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `invite_links` | array of object | yes |  |  |
| `invite_links[].expires_at` | string (date-time) | yes |  |  |
| `invite_links[].id` | string | yes |  |  |
| `invite_links[].label` | string | yes |  |  |
| `invite_links[].live` | boolean | yes |  |  |
| `invite_links[].max_uses` | integer | yes |  |  |
| `invite_links[].remaining` | integer | yes |  |  |
| `invite_links[].revoked` | boolean | yes |  |  |
| `invite_links[].role` | string | yes |  |  |
| `invite_links[].used` | integer | yes |  |  |

### POST /api/admin/projects/{project_id}/invite-links

Make an invite link on a project

One link a group can join the project itself through: up to max_uses accounts, for days days. role viewer makes them view only, on an app that has it.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_create_remote_invite_link`](mcp.md#admin_create_remote_invite_link) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `days` | integer | yes | 1 to 365 |  |
| `label` | string | yes | 1 to 80 characters |  |
| `max_uses` | integer | yes | 1 to 1000 |  |
| `role` | string | no | one of `user`, `viewer` | What people who join through it become. Default user. |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `join_url` | string (uri) | yes |  | The link to hand out, shown in this response only. A credential: treat it like a password. |

### POST /api/admin/projects/{project_id}/invite-links/{link_id}/revoke

Stop an invite link on a project

The link stops working. Accounts already made through it stay.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_remote_invite_link`](mcp.md#admin_revoke_remote_invite_link) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `link_id` | path | string | yes | 1 to 100 characters | The link's id, from admin_list_remote_invite_links. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/keys/{key_id}/revoke

Revoke an API key on a project

The key stops working there now, and so does every key it made.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_remote_key`](mcp.md#admin_revoke_remote_key) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `key_id` | path | string | yes | 1 to 100 characters | The key's id, from admin_list_remote_keys. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/remote-users

Make an account on a project

Make an account on the project itself, with a role and an optional expiry, and get its login link. With user_id, that person here is put on the project with it.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_create_remote_user`](mcp.md#admin_create_remote_user) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `display_name` | string | yes | 1 to 80 characters |  |
| `expires_in_days` | integer | no | 0 to 3650 | The account stops working this many days from now. 0 or left out: never. Needs an app with accounts that expire. |
| `role` | string | yes | one of `user`, `viewer`, `admin` | viewer needs an app with view-only accounts. |
| `user_id` | integer | no | at least 0 | Someone here to put on the project with it, so their Open signs into it. Left out: a guest with no account here. |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `login_url` | string | yes |  | Signs the new account in. It works once, for 15 minutes. A credential: treat it like a password. |
| `remote_user_id` | integer | yes |  |  |

### POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/delete

Delete an account on a project

Delete the account on the project itself, with everything it held there. Anyone here it belonged to comes off the project. It cannot be undone.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_delete_remote_user`](mcp.md#admin_delete_remote_user) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | path | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/disabled

Disable or enable an account on a project

disabled true stops the account signing in there, and ends its sessions.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_remote_disabled`](mcp.md#admin_set_remote_disabled) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | path | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `disabled` | boolean | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/expiry

Set when an account on a project expires

The account stops working expires_in_days from now; 0 means never. Needs an app with accounts that expire.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_remote_expiry`](mcp.md#admin_set_remote_expiry) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | path | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer | yes | 0 to 3650 | Stops working this many days from now. 0: never. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_at` | string (date-time) or null | yes |  |  |
| `ok` | boolean | yes |  |  |

### GET /api/admin/projects/{project_id}/remote-users/{remote_user_id}/keys

An account's API keys on a project

The account's API keys there, never their secrets. Needs an app where the portal's key may make keys.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_list_remote_keys`](mcp.md#admin_list_remote_keys) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | path | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `keys` | array of object | yes |  |  |
| `keys[].created_at` | string (date-time) | yes |  |  |
| `keys[].expired` | boolean | yes |  |  |
| `keys[].expires_at` | string (date-time) or null | no |  |  |
| `keys[].id` | string | yes |  |  |
| `keys[].last_used_at` | string (date-time) or null | no |  |  |
| `keys[].live` | boolean | yes |  |  |
| `keys[].minted_by_key` | string | no |  |  |
| `keys[].name` | string | yes |  |  |
| `keys[].revoked` | boolean | yes |  |  |
| `keys[].role` | string | yes |  |  |

### POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/keys

Make an API key on a project

An API key for the account on the project, shown once. It always expires, its role is no higher than the account's, and revoking the portal's admin key there revokes it too.

|  |  |
|---|---|
| Who may call | browser session only, admins |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | path | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `expires_in_days` | integer | yes | 1 to 365 | A key made from the portal always expires. |
| `name` | string | yes | 1 to 80 characters |  |
| `role` | string | yes | one of `user`, `viewer`, `admin` | No higher than the account's own role. |

**Reply 201**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `credential` | string | yes |  |  |
| `expired` | boolean | yes |  |  |
| `expires_at` | string (date-time) or null | no |  |  |
| `id` | string | yes |  |  |
| `last_used_at` | string (date-time) or null | no |  |  |
| `live` | boolean | yes |  |  |
| `minted_by_key` | string | no |  |  |
| `name` | string | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `role` | string | yes |  |  |

### POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/login-link

A login link for an account on a project

A fresh login link for one account on the project. It works once, for 15 minutes.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_remote_login_link`](mcp.md#admin_remote_login_link) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | path | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `login_url` | string | yes |  | Open it to sign in to the project. It works once, for 15 minutes. A credential: treat it like a password. |
| `project_id` | integer | yes |  |  |

### POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/revoke-access

Sign an account out everywhere on a project

End every session, login link and API key the account has there. The account stays enabled; a new login link signs it back in.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_revoke_remote_access`](mcp.md#admin_revoke_remote_access) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | path | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/role

Change an account's role on a project

admin, user, or viewer (view only, on an app that has it). The project's owner cannot be changed.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_set_remote_role`](mcp.md#admin_set_remote_role) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | path | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `role` | string | yes | one of `user`, `viewer`, `admin` |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/admin/users/{user_id}/project-accounts

Disable someone's accounts on every project

For every project they are on, disable their account there (on_project disable) or end its sessions, links and keys (revoke). Use it before disabling or deleting them here. Answers how each project took it.

|  |  |
|---|---|
| Who may call | browser session or API key, admins |
| MCP tool | [`admin_act_on_project_accounts`](mcp.md#admin_act_on_project_accounts) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `user_id` | path | integer | yes | at least 1 | The account's id here. |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `on_project` | string | yes | one of `disable`, `revoke` | Disable their account on every project, or end its sessions, links and keys there. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `results` | array of object | yes |  | One per project they are on that the portal made or linked an account for. |
| `results[].error` | string | no |  |  |
| `results[].name` | string | yes |  |  |
| `results[].ok` | boolean | yes |  |  |
| `results[].project_id` | integer | yes |  |  |

## Projects

| Method | Path | Summary | Who may call |
|---|---|---|---|
| GET | [`/api/projects`](#get-apiprojects) | Your projects | browser session or API key, anyone signed in |
| POST | [`/api/projects/{project_id}/favorite`](#post-apiprojectsproject_idfavorite) | Star or unstar a project | browser session or API key, anyone signed in |
| POST | [`/api/projects/{project_id}/join`](#post-apiprojectsproject_idjoin) | Join a project | browser session or API key, anyone signed in |
| POST | [`/api/projects/{project_id}/leave`](#post-apiprojectsproject_idleave) | Leave a project | browser session or API key, anyone signed in |
| POST | [`/api/projects/{project_id}/open`](#post-apiprojectsproject_idopen) | Sign in to a project | browser session or API key, anyone signed in |
| POST | [`/api/projects/{project_id}/own-key`](#post-apiprojectsproject_idown-key) | Use your own key on a project | browser session only, anyone signed in |
| POST | [`/api/projects/{project_id}/own-key/remove`](#post-apiprojectsproject_idown-keyremove) | Stop using your own key on a project | browser session or API key, anyone signed in |
| POST | [`/api/projects/{project_id}/restore`](#post-apiprojectsproject_idrestore) | Undo leaving a project | browser session or API key, anyone signed in |
| POST | [`/api/projects/{project_id}/tags`](#post-apiprojectsproject_idtags) | Tag a project | browser session or API key, anyone signed in |

### GET /api/projects

Your projects

Every project you may use, a page at a time: the ones you are on first (favorites first), then the ones you could join. member says which.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`list_projects`](mcp.md#list_projects) |
| Errors | 400, 401, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `limit` | query | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | query | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | query | string | no |  | Keep projects whose name, description or one of your tags contains this, in any case. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `next` | integer or null | yes |  | Offset of the next page, or null on the last one. |
| `projects` | array of [`ProjectCard`](#projectcard) | yes |  |  |
| `total` | integer | yes |  |  |

### POST /api/projects/{project_id}/favorite

Star or unstar a project

favorite true keeps the project at the top of your list.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`favorite_project`](mcp.md#favorite_project) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `favorite` | boolean | yes |  |  |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project` | [`ProjectCard`](#projectcard) | yes |  |  |

### POST /api/projects/{project_id}/join

Join a project

Make yourself an account on a project you may use, named as you are here, and get a login link for it. 409 when you are already on it or it is not set up yet; 424 when the project refused or did not answer, with what it said.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`join_project`](mcp.md#join_project) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `login_url` | string | yes |  | Open it to sign in to the project. It works once, for 15 minutes. A credential: treat it like a password. |
| `project_id` | integer | yes |  |  |

### POST /api/projects/{project_id}/leave

Leave a project

Take a project off your list. Your account on the project stays as it is; joining again signs you back into it. Undo with restore_project_membership within 10 minutes keeps your stars, tags and key.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`leave_project`](mcp.md#leave_project) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/projects/{project_id}/open

Sign in to a project

A fresh login link for your account on a project you are on: with your own key when you attached one, otherwise made by the portal's admin key there. 409 when you are not on it yet (join it), 424 when the project refused, with what it said.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`open_project`](mcp.md#open_project) |
| Errors | 400, 401, 403, 404, 409, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `login_url` | string | yes |  | Open it to sign in to the project. It works once, for 15 minutes. A credential: treat it like a password. |
| `project_id` | integer | yes |  |  |

### POST /api/projects/{project_id}/own-key

Use your own key on a project

Keep an API key of yours on that project, and sign in with it from now on. On the project, switch on Sign-in links for that key (API keys card), or opening it will be refused.

|  |  |
|---|---|
| Who may call | browser session only, anyone signed in |
| Errors | 400, 401, 403, 404, 424, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `key` | string | yes | 8 to 400 characters | An API key of your own on that project, allowed to make sign-in links there. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |
| `remote_name` | string | yes |  | The account the key acts as there. |
| `remote_role` | string | yes |  |  |

### POST /api/projects/{project_id}/own-key/remove

Stop using your own key on a project

Forget your own key on that project. If the key was all that put you on it, you are no longer on it here.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`remove_own_project_key`](mcp.md#remove_own_project_key) |
| Errors | 400, 401, 403, 404, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

### POST /api/projects/{project_id}/restore

Undo leaving a project

Put back a project you left in the last 10 minutes, as it was.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`restore_project_membership`](mcp.md#restore_project_membership) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project` | [`ProjectCard`](#projectcard) | yes |  |  |

### POST /api/projects/{project_id}/tags

Tag a project

Your own labels for a project you are on, replacing the ones it had.

|  |  |
|---|---|
| Who may call | browser session or API key, anyone signed in |
| MCP tool | [`tag_project`](mcp.md#tag_project) |
| Errors | 400, 401, 403, 404, 409, 429, 503 |

**Parameters**

| Name | In | Type | Required | Limits | Description |
|---|---|---|---|---|---|
| `project_id` | path | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

**Request body**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `csrf_token` | string | no |  | From GET /api/me. Browser sessions only. |
| `tags` | array of string | yes |  | Up to 10, each 1-30 characters. They are lowercased. An empty list clears them. |

**Reply 200**: Success

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project` | [`ProjectCard`](#projectcard) | yes |  |  |

## Schemas

The shapes the routes above refer to by name.

### APIKey

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `created_by` | integer | no |  | The admin who made this key for its owner. Absent if the owner made it. |
| `created_by_name` | string | no |  |  |
| `display_name` | string | no |  | The key's owner. On the admin list only. |
| `expired` | boolean | yes |  |  |
| `expires_at` | string (date-time) | no |  |  |
| `id` | string | yes |  |  |
| `last_used_at` | string (date-time) | no |  |  |
| `live` | boolean | yes |  |  |
| `may_sign_in` | boolean | yes |  | Whether this key may make a login link for its own account. |
| `name` | string | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `revoked_at` | string (date-time) | no |  |  |
| `role` | string | yes | one of `admin`, `user` |  |
| `user_id` | integer | yes |  |  |

### BulkReply

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `done` | integer | yes |  | How many items worked. |
| `failed` | integer | yes |  | How many did not. |
| `results` | array of [`BulkResult`](#bulkresult) | yes |  | One per item, in the order they were sent. |

### BulkResult

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `error` | string | no |  | When it failed: the route's words for why. |
| `index` | integer | yes |  | The item's place in items, from 0. |
| `ok` | boolean | yes |  |  |
| `reply` | object | no |  | When it worked: the route's own reply, if it has one. |
| `request_id` | string | no |  | On a 5xx: the reference that finds this failure in the log. |
| `status` | integer | no |  | When it failed: the status the route would have answered. |

### CatalogProject

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `base_url` | string | yes |  |  |
| `board_share_url` | string | yes |  | A view-only TodoBoard share link, which everyone who may see the project gets. |
| `board_url` | string | yes |  | Its TodoBoard board, for admins. |
| `capabilities` | array of string | yes |  | What Manage project may do there: viewer_role, account_expiry, key_minting, invite_links, key_sign_in. |
| `check_note` | string | yes |  |  |
| `checked_at` | string (date-time) or null | yes |  |  |
| `color` | string | yes |  |  |
| `contract` | integer | yes |  | Which portal contract it speaks: 2 lets Manage project do everything its capabilities list; 0 is older. |
| `created_at` | string (date-time) | yes |  |  |
| `description` | string | yes |  |  |
| `excluded_user_ids` | array of integer | yes |  | People it is kept from. |
| `github_url` | string | yes |  | The project's code, for admins. |
| `health` | string | yes |  | up, down, or empty before the first check. |
| `icon` | string | yes |  |  |
| `id` | integer | yes |  |  |
| `key_set` | boolean | yes |  | Whether the portal holds an admin key for it. |
| `key_status` | string | yes |  | ok, invalid, not_admin or missing, from the last check. |
| `listed` | boolean | yes |  |  |
| `members` | integer | yes |  |  |
| `name` | string | yes |  |  |
| `position` | integer | yes |  |  |
| `version` | string | yes |  | The app's version, from its last check. |

### Error

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `error` | string | yes |  |  |
| `request_id` | string | no |  | On a 5xx: the reference that finds this failure in the log. |

### IPNames

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `names` | array of object | yes |  |  |
| `names[].created_at` | string (date-time) | yes |  |  |
| `names[].name` | string | yes |  |  |
| `names[].target` | string | yes |  |  |

### InviteLink

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_by` | integer | yes |  |  |
| `created_by_name` | string | no |  |  |
| `expires_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `label` | string | yes |  |  |
| `live` | boolean | yes |  |  |
| `max_uses` | integer | yes |  |  |
| `remaining` | integer | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `used` | integer | yes |  |  |

### Job

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `interrupted` | boolean | yes |  |  |
| `interval_seconds` | integer | yes |  |  |
| `last_error` | string | no |  |  |
| `last_finished` | string (date-time) | no |  |  |
| `last_note` | string | no |  |  |
| `last_ok` | boolean or null | no |  |  |
| `last_started` | string (date-time) | no |  |  |
| `name` | string | yes |  |  |
| `next_due` | string (date-time) | yes |  |  |
| `running` | boolean | yes |  |  |
| `runs` | integer | yes |  |  |

### Joined

One of these:

1. An object:

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes |  |  |
| `ok` | boolean | yes |  |  |

2. An object:

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `challenge` | string | yes |  |  |
| `two_factor` | boolean | yes |  |  |

### LogDay

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `date` | string | yes |  |  |
| `entries` | array of object | yes |  |  |
| `next` | integer or null | yes |  |  |

### LogDays

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `days` | array of object | yes |  |  |
| `dropped` | integer | yes |  |  |
| `settings` | object | yes |  |  |
| `total_bytes` | integer | yes |  |  |

### LoginLink

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `expires_at` | string (date-time) or null | yes |  | null when it never expires. |
| `last_used_at` | string (date-time) or null | yes |  |  |
| `max_uses` | integer | yes |  | Sign-ins it allows. 0 when unlimited. |
| `unlimited` | boolean | yes |  |  |
| `uses` | integer | yes |  | Sign-ins so far. |

### Maintenance

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `error` | string | yes |  |  |
| `maintenance` | string | yes | one of `read_only`, `full` |  |

### Me

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `account_role` | string | yes |  |  |
| `announcement` | string | yes |  | server.announcement. Empty when there is none. |
| `api_doc_url` | string | yes |  | Where to read the API reference as a page for people: config/api.md in app.repo, or /docs/api when there is no repo. |
| `api_key_id` | string | no |  | API keys only. |
| `api_key_name` | string | no |  | API keys only. |
| `api_keys_enabled` | boolean | yes |  |  |
| `auth` | string | yes |  |  |
| `authenticated` | boolean | yes |  |  |
| `csrf_token` | string | no |  | Browser sessions only. |
| `description` | string | yes |  | app.description. Empty when there is none. |
| `display_name` | string | yes |  |  |
| `has_password` | boolean | yes |  |  |
| `has_pin` | boolean | yes |  |  |
| `id` | integer | yes |  |  |
| `invite_links_enabled` | boolean | yes |  | Whether invite links are on. |
| `is_owner` | boolean | yes |  |  |
| `login_link` | [`LoginLink`](#loginlink) or null | yes |  | Your own live login link, or null. |
| `logs_enabled` | boolean | yes |  |  |
| `logs_ui_clicks` | boolean | yes |  |  |
| `maintenance` | string | yes | one of `none`, `read_only`, `full` |  |
| `mcp_doc_url` | string | yes |  | Where to read the MCP tools as a page for people: config/mcp.md in app.repo, or /docs/mcp when there is no repo. |
| `mcp_enabled` | boolean | yes |  | server.mcp: whether POST /mcp answers. |
| `mcp_url` | string | yes |  | Where to read the MCP reference: config/mcp.json in app.repo, or /api/mcp.json when there is no repo. Only with server.mcp on. |
| `openapi_url` | string | yes |  | Where to read this API's reference: config/api.json in app.repo, or /api/api.json when there is no repo. |
| `password_blocked` | boolean | yes |  |  |
| `passwords_enabled` | boolean | yes |  |  |
| `repo` | string | yes |  | app.repo without a trailing / or .git: where the source lives. Empty when it is not set. |
| `role` | string | yes |  | The effective role: never higher than an API key's own. |
| `shortcuts` | map of string | yes |  | The keyboard shortcuts you changed, by action. Mod is Cmd on a Mac and Ctrl elsewhere. |
| `show_about` | boolean | yes |  | app.show_about: whether the menu has About. |
| `theme_id` | string | yes |  | The theme you use. Empty: the first. |
| `themes` | array of [`Theme`](#theme) or null | yes |  | Your themes (see the Theme schema): your own list, else the app's defaults, else null for the built-in ones. |
| `themes_builtins` | integer | yes |  | Which built-in themes that list has seen (see builtins on POST /api/account/themes). |
| `themes_custom` | boolean | yes |  | Whether you have your own list rather than following the defaults. |
| `two_factor` | boolean | yes |  |  |
| `two_factor_enabled` | boolean | yes |  |  |
| `username` | string | yes |  |  |
| `version` | string | yes |  | The running build. dev when none was stamped. |

### Member

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes |  |  |
| `id` | integer | yes |  |  |

### NewAPIKey

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `created_by` | integer | no |  | The admin who made this key for its owner. Absent if the owner made it. |
| `created_by_name` | string | no |  |  |
| `credential` | string | yes |  | Shown in this response only, and never recoverable. |
| `display_name` | string | no |  | The key's owner. On the admin list only. |
| `expired` | boolean | yes |  |  |
| `expires_at` | string (date-time) | no |  |  |
| `id` | string | yes |  |  |
| `last_used_at` | string (date-time) | no |  |  |
| `live` | boolean | yes |  |  |
| `may_sign_in` | boolean | yes |  | Whether this key may make a login link for its own account. |
| `name` | string | yes |  |  |
| `revoked` | boolean | yes |  |  |
| `revoked_at` | string (date-time) | no |  |  |
| `role` | string | yes | one of `admin`, `user` |  |
| `user_id` | integer | yes |  |  |

### ProjectAccess

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `allowed` | boolean | yes |  | Whether they may see and join it. Admins always may. |
| `disabled` | boolean | yes |  |  |
| `display_name` | string | yes |  |  |
| `member` | boolean | yes |  |  |
| `own_key` | boolean | yes |  |  |
| `remote_name` | string | yes |  |  |
| `remote_user_id` | integer | yes |  |  |
| `role` | string | yes |  |  |
| `user_id` | integer | yes |  |  |
| `via` | string | yes |  |  |

### ProjectCard

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `base_url` | string | yes |  |  |
| `board_share_url` | string | yes |  | A view-only link to the project's TodoBoard board, or empty. |
| `checked_at` | string (date-time) or null | yes |  |  |
| `color` | string | yes |  |  |
| `created_at` | string (date-time) | yes |  |  |
| `description` | string | yes |  |  |
| `favorite` | boolean | yes |  |  |
| `github_url` | string | yes |  | Where the project's code lives, or empty. |
| `health` | string | yes |  | up, down, or empty before the first check. |
| `icon` | string | yes |  |  |
| `id` | integer | yes |  |  |
| `last_opened_at` | string (date-time) or null | yes |  |  |
| `listed` | boolean | yes |  | false only for a project hidden from everyone but admins. |
| `member` | boolean | yes |  | You are on it: open_project signs you in. |
| `name` | string | yes |  |  |
| `new` | boolean | yes |  | Added in the last 14 days. |
| `own_key` | boolean | yes |  | Opening uses your own key on that project. |
| `position` | integer | yes |  |  |
| `ready` | boolean | yes |  | The portal holds a working admin key, so joining works. |
| `remote_name` | string | yes |  | Your account's name there. |
| `tags` | array of string | yes |  |  |
| `via` | string | yes |  | How: joined, linked (an admin matched your account), own_key. Empty when not a member. |

### PruneResult

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `bytes_after` | integer | yes |  |  |
| `bytes_before` | integer | yes |  |  |
| `compressed` | integer | yes |  |  |
| `deleted` | integer | yes |  |  |

### RemoteUser

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `disabled` | boolean | yes |  |  |
| `display_name` | string | yes |  |  |
| `expired` | boolean | yes |  |  |
| `expires_at` | string (date-time) or null | yes |  | When it stops working there, or null for never (or an app without expiring accounts). |
| `id` | integer | yes |  |  |
| `owner` | boolean | yes |  |  |
| `portal_name` | string | no |  |  |
| `portal_user_id` | integer | no |  | Who it belongs to here, when someone is on the project with it. |
| `role` | string | yes |  |  |

### Revoked

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `api_keys` | integer | yes |  |  |
| `login_links` | integer | yes |  |  |
| `passwords` | integer | yes |  |  |
| `sessions` | integer | yes |  |  |

### Session

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `current` | boolean | yes |  |  |
| `expires_at` | string (date-time) | yes |  |  |
| `id` | string | yes |  |  |
| `ip` | string | no |  |  |
| `ip_name` | string | no |  | What the owner calls this address on the Logs screen. The owner's own list only. |
| `label` | string | no |  | The name you gave this device. |
| `last_seen_at` | string (date-time) | no |  |  |
| `user_agent` | string | no |  |  |

### SettingsResult

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `applied` | array of string or null | yes |  |  |
| `changes` | array of object or null | yes |  |  |
| `restart` | array of string or null | yes |  |  |
| `scripts` | array of string or null | yes |  |  |
| `unchanged` | array of string or null | yes |  |  |
| `version` | string | yes |  |  |

### SettingsSnapshot

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `error` | object or null | yes |  |  |
| `error.at` | string (date-time) | yes |  |  |
| `error.message` | string | yes |  |  |
| `error.origin` | string | yes |  |  |
| `file` | string | yes |  |  |
| `in_container` | boolean | yes |  |  |
| `secret_file` | string | yes |  |  |
| `sections` | array of object | yes |  |  |
| `sections[].help` | string | yes |  |  |
| `sections[].keys` | array of object | yes |  |  |
| `sections[].keys[].apply` | string | yes |  |  |
| `sections[].keys[].default` | any | yes |  |  |
| `sections[].keys[].file_only` | boolean | yes |  | A path: change it in config.yaml, never here. |
| `sections[].keys[].help` | string | yes |  |  |
| `sections[].keys[].in_file` | boolean | yes |  |  |
| `sections[].keys[].key` | string | yes |  |  |
| `sections[].keys[].locked` | boolean | yes |  |  |
| `sections[].keys[].name` | string | yes |  |  |
| `sections[].keys[].pending_restart` | boolean | yes |  |  |
| `sections[].keys[].running` | any | yes |  |  |
| `sections[].keys[].type` | string | yes |  |  |
| `sections[].keys[].value` | any | yes |  |  |
| `sections[].name` | string | yes |  |  |
| `stale` | boolean | yes |  |  |
| `version` | string | yes |  | Send it back when saving: a changed file is refused with a 409. |
| `writable` | boolean | yes |  |  |

### SignIn

One of these:

1. An object:

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `ok` | boolean | yes |  |  |

2. An object:

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `challenge` | string | yes |  |  |
| `two_factor` | boolean | yes |  |  |

### Theme

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `accent` | string | yes |  | The main color: what you can press. |
| `background` | string | no |  | The page. Left out: the mode's own. |
| `button` | object or null | no |  | How the theme's buttons look. A choice left out is the default (rounded, normal, raised, normal). |
| `button.case` | string | no | one of `normal`, `upper` |  |
| `button.depth` | string | no | one of `flat`, `raised`, `glossy` |  |
| `button.shape` | string | no | one of `square`, `soft`, `rounded`, `pill` |  |
| `button.weight` | string | no | one of `normal`, `medium`, `bold` |  |
| `dark` | object or null | no |  | A light theme's colors for dark mode, kept while it is light. Switching the theme to dark swaps them in. A light theme's light (or a dark theme's dark) is dropped. |
| `dark.accent` | string | yes |  | The main color: what you can press. |
| `dark.background` | string | no |  | The page. Left out: the mode's own. |
| `dark.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `dark.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `dark.surface` | string | no |  | Cards. Left out: the mode's own. |
| `id` | string | yes |  | Lower-case letters, digits, - and _, up to 32. |
| `light` | object or null | no |  | A dark theme's colors for light mode, kept while it is dark. Switching the theme to light swaps them in. |
| `light.accent` | string | yes |  | The main color: what you can press. |
| `light.background` | string | no |  | The page. Left out: the mode's own. |
| `light.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `light.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `light.surface` | string | no |  | Cards. Left out: the mode's own. |
| `mode` | string | yes | one of `light`, `dark` |  |
| `name` | string | yes |  | 1 to 60 characters. |
| `overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `surface` | string | no |  | Cards. Left out: the mode's own. |

### User

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `created_at` | string (date-time) | yes |  |  |
| `disabled` | boolean | yes |  |  |
| `display_name` | string | yes |  |  |
| `has_password` | boolean | yes |  |  |
| `id` | integer | yes |  |  |
| `login_link` | [`LoginLink`](#loginlink) or null | yes |  | The account's live login link, or null. |
| `owner` | boolean | yes |  |  |
| `password_blocked` | boolean | yes |  |  |
| `role` | string | yes | one of `admin`, `user` |  |
| `two_factor` | boolean | yes |  |  |
| `username` | string | yes |  | Empty until a password is set. |
