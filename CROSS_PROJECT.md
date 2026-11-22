# Cross-project notes

The portal only works when the apps in its catalog speak the same API. This file
records the changes other projects need, as we find them.

Each entry is either:
- **fixed here**: small enough that it was changed in that repo directly, with the date, or
- **plan to paste**: a self-contained brief to paste into a new Claude session opened in that project.

Newest entries go at the top of each section.

---

## All catalog apps (go-webserver / python-webserver)

### The contract the portal relies on (reference)
Every project in the catalog must serve these (see `server/remote/remote.go`):

| Call | Used for | Key |
|---|---|---|
| `GET  /api/me` | who a key is, and at what role | admin or own |
| `GET  /api/health` | the up/down check, every 5 minutes | none |
| `POST /api/admin/users` | Join: make an account, answer a login link | admin |
| `POST /api/admin/users/{id}/reissue-login` | Open: a fresh login link | admin |
| `GET  /api/admin/users` | matching portal people to existing accounts | admin |
| `POST /api/account/login-link` | Open with someone's own key | own key, with Sign-in links on |

If an app falls out of sync (Check now shows a refused key, or Join/Open answers
424, which the portal used to send as 502), check it against this table first. The usual fix is the `webserver-upgrade`
skill run in that app.

### Portal contract v2 (2026-11-20): exact wire format
Added to both webserver skills so the portal's **Manage project** screen can run
an app from outside. The portal turns each part on only when `/api/me` lists its
capability, so an app that is behind keeps working with v1. Field names and error
codes below are binding: the Go and Python templates and `server/remote/remote.go`
must match them exactly. Every route answers `{"error": "..."}` on refusal, as today.
An "admin key" below means an admin-role API key, or an admin session.

**Discovery**: `GET /api/me` adds `"contract": 2` and `"capabilities": [...]`, drawn from:
`viewer_role`, `account_expiry`, `key_minting`, `invite_links` (only while invite links are
switched on), `key_sign_in` (`POST /api/account/login-link` takes a key with `may_sign_in`).

**Viewer role**: `role` everywhere also takes `"viewer"`, ranked viewer < user < admin.
A viewer's session or key may send only GET/HEAD, plus `POST /api/logout`, its own
`/api/account/sessions*` revokes and `POST /api/account/login-link/cancel`. Anything else
answers **403 `{"error":"view_only"}`**. `/api/me` reports `"role": "viewer"`.

**Account expiry** (capability `account_expiry`):
| Call | Body | Answer |
|---|---|---|
| `POST /api/admin/users` (existing) | adds optional `"account_expires_at"`: RFC 3339 time, or omitted for never. (`expires_in_minutes` stays the *login link's* expiry.) | as today: `user_id`, `login_path` |
| `POST /api/admin/users/{user_id}/expiry` | `{"expires_at": "<RFC 3339>" or null}` (null = never) | `{"ok": true, "expires_at": ... or null}` |
| `GET /api/admin/users` (existing) | | each user adds `"expires_at"` (RFC 3339 or null) and `"expired"` (bool) |

When it passes, the account counts as disabled. Sign-in, its sessions and its keys stop
working, and housekeeping disables it and revokes its access. An admin may expire anyone
but the owner.

**Role change**: `POST /api/admin/users/{user_id}/role` `{"role": "admin"|"user"|"viewer"}`
(existing, admin key OK), now also takes `viewer`.

**Invite links** (capability `invite_links`): existing routes. Create adds optional
`"role": "user"|"viewer"` (default `user`). List rows add `"role"`. The portal uses
`GET /api/admin/invite-links`, `POST /api/admin/invite-links`
(`{label, max_uses, days, role}` → `{..., "join_path"}`), and
`POST /api/admin/invite-links/{link_id}/revoke`.

**Key minting** (capability `key_minting`). The calling key needs `may_mint_keys`, which
only the app's owner can switch on, from a browser session, on an admin-role key
(`POST /api/keys/{key_id}/mint-keys` `{"may_mint_keys": bool}`, session only). Otherwise
it gets **403 `{"error":"key_minting_off"}`**.
| Call | Body | Answer |
|---|---|---|
| `POST /api/admin/users/{user_id}/keys` | `{"name": 1-80, "role": "admin"|"user"|"viewer", "expires_in_days": 1-365}` (required, never 0) | the app's `NewAPIKey` shape: key row + `"credential"`, shown once |
| `GET /api/admin/users/{user_id}/keys` | | `{"keys": [key rows]}`; rows add `"minted_by_key"` (id or empty) |
| `POST /api/admin/keys/{key_id}/revoke` | | `{"ok": true}` |

Rules for minted keys:
- The role is capped at the account's own role.
- A minted key can never get `may_mint_keys` or `may_sign_in`.
- Revoking or deleting a key revokes every key it minted, down the whole chain, and voids
  the links it made (as `VoidLinksMintedByKeyTx` already does for links).
- The owner's keys cannot be minted or revoked this way.

Disable, revoke access and delete are the existing
`POST /api/admin/users/{user_id}/disabled` `{"disabled": bool}`, `/revoke-access` and
`/delete` routes, which already take an admin key.

### Known gap: apps that cannot make a sign-in link for a key
Attaching your own key only checks `GET /api/me`, so the card says "Signs you in
as …" even on an app whose `POST /api/account/login-link` turns keys away. Open
then fails. Since 2026-11-20 the portal words that refusal ("switch on Sign-in
links for that key… or it needs updating"). Each app that still answers 403
`this endpoint requires a signed-in session, not an API key` there needs the
fix below, one app at a time as they come up.

---

## Rook (`~/WORKSPACE/PROJECTS/Games/Rook`)

### 2026-11-20, plan to paste: let a key make its own account a sign-in link
Found when a portal user attached their own Rook key and Open failed. Rook was
scaffolded from go-webserver at `145e4d48` and its login-link route is still
`access=SESSION_ONLY` (`server/routes/account.go`, `CreateMyLoginLink`). The
comment in `server/security/security.go` already describes `APIKey.MaySignIn`,
but the field and the route change were never brought in.

> Run the `webserver-upgrade` skill in this app and take the upgrade that lets an
> API key make its own account a login link once that account switched **Sign-in
> links** on for the key (the `MaySignIn` flag on API keys, set from a browser
> session on the API keys card). Check TodoBoard
> (`~/WORKSPACE/PROJECTS/TodoBoard/server/routes/account.go`, `createMyLoginLink`)
> for how it should end up. `POST /api/account/login-link` must take an API key
> with that flag on. It must accept `{ "max_uses": 1, "expires_in_minutes": 15 }`,
> cap a key's link at once and 15 minutes, and answer `{ "login_path": "/login/…" }`.
> Every link must record the key that made it, and revoking the key must void it
> (`VoidLinksMintedByKeyTx`). A key without the flag still gets 403. Keep it
> browser-only for minting keys themselves. Add tests (key with flag gets 200 and a
> login_path; key without gets 403; revoking the key voids the link), regenerate the
> API docs, and check the portal: attach a Rook key with Sign-in links on, then
> **Open** and **Make a new login link** both work.
>
> Also check the admin-key routes the portal uses (`/api/admin/users`,
> `/api/admin/users/{id}/reissue-login`). They answer "browser session only" in
> `config/api.md` too, which would also break Join for Rook.

---

## TodoBoard (`~/WORKSPACE/PROJECTS/TodoBoard`)

### 2026-11-19: View-only board links for portal users. Nothing required.
TodoBoard already has view-only share links (`server/models/share_link.go`,
mode `view`, served at `/s/<id>.<secret>`). The portal stores one per project
(Catalog → "TodoBoard view-only link"), and everyone who may see the project gets
a Board button. To make one: open the board in TodoBoard → Share → mode **View**
→ copy the link → paste it into the project in the portal.

### Optional, plan to paste: let an admin API key mint view-only board links
Not needed today. It would let the portal make and rotate the links itself
instead of the admin pasting them. It loosens a deliberate rule (share links are
browser-session only), so decide before running it.

> In TodoBoard, `POST /api/boards/{board_id}/shares` (`createShareLink` in
> `server/routes/sharing.go`) is `access=SESSION_ONLY`, so only a browser session
> can mint a share link. Allow an **admin-role API key** to call it too, but
> **only for `mode: "view"` board links** (no `card_id` and no `list_id`).
> Comment and edit links, and any non-admin key, stay session-only and get 403 with a
> clear message. Keep the owner-of-the-board check (`access.Manage`). Record an
> audit entry naming the key. Add tests for: admin key + view → 201 and
> `share_path`; admin key + edit → 403; user key + view → 403; session unchanged.
> Regenerate the API docs (`scripts/api_docs.sh`) and update the route's
> description. Expose it over MCP as `create_view_share` if the MCP tool list is
> generated from `mcp=` on routes.

---

## ProjectPortal (this repo)

### Follow-up idea: check that each project's view-only link still works
The project-health job (`server/jobs/portal.go`, `server/portal/check.go`) could
call TodoBoard's public `GET /api/shared` with `X-Share-Token: <id>.<secret>` (the
part after `/s/`) and flag a revoked or expired link on the Catalog row, so
people never get a dead Board button.
