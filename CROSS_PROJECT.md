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

### 2026-11-23, portal only, nothing to do in the apps: apps are off until allowed
Each catalog app is now off for every portal user until an admin allows it for
them (Users → Apps, or the Access screen). People only see, and can only Join,
the apps they are allowed. Admins see everything. The first start after deploy
runs migration 1: each person keeps the apps they already have an account on,
and nothing else. The apps themselves need no change.

### The portal contract: the skills are the source of truth (v3, 2026-11-22)
The binding wire format now lives in the webserver skills, the same text in
both: the "Portal contract" section of `~/.claude/skills/go-webserver/SKILL.md`
and `~/.claude/skills/python-webserver/SKILL.md`. It covers the routes,
`/api/me`'s fields, the capabilities, the refusal codes, service accounts and
rotation. `server/remote/` is written against that section. Change the
contract there, in both skills, and not here.

v3 in brief:
- `"contract": 3`. ProjectPortal works through a service account the app's
  owner connects (Connect on the owner's API keys screen, or
  `manage portal-connect`).
- `/api/me` adds `key` (grants and expiry), `owner`, `kind`, `skill` and
  `skill_commit`.
- Every 4xx carries a `code`.
- `POST /api/keys/self/rotate`.
- Admin actions and webhooks name the actor.
- Check a live app with `scripts/contract_check.sh <url>`, the key on stdin.

The v2 parts (viewer role, account expiry, invite-link roles, key minting) are
unchanged, and the skills' section spells them out too.

### 2026-11-22, upgraded, not yet committed or deployed: `admin_sessions`
The portal's People screen (phase 3b) shows each account's devices through
`GET /api/admin/users/{id}/sessions`, capability `admin_sessions` (skills commit
`608cadf`, upgrade `admin-sessions`). `webserver-upgrade` has applied it, tests
passing, in ZoteroWebDAV, QRCodeOrganizer, MediaTracker, PaperReferenceManagerAddons,
TodoBoard, MRIExperimentalDesignPlanner, Rook and ProjectPortal. Each still needs
a commit with its `gitPushUpdate.sh`, then Check now in the portal.
Not upgraded, being further behind: AIS Counter, E2EEDB, iPB, VocabTrainer.
Until an app has it, People shows "Not listed" for that app's devices and
suggests matches by name only.

Notes from the rollout:
- TodoBoard's Devices entry lives in its own row menu, which now follows the
  same ladder (`seesSessions`) as the template's button.
- Rook builds `/api/me` as a map and has no `sessions_test.go`. It got the
  `admin_sessions` line in `capabilities()` and a `devices` test helper in
  `server/routes/sessions_helpers_test.go`. Its admin screen is the owner's
  alone, so there is no button change.
- Older test failures, already on the committed code and not caused by the
  upgrade: PaperReference `test_list_ops.test_cite` (the list of citation styles
  differs) and the MRI planner's 4 `test_actions.ToolArgumentsTest` errors
  (`'Call' object has no attribute 'bind'`).

### Known gap: apps that cannot make a sign-in link for a key
Attaching your own key only checks `GET /api/me`, so the card says "Signs you in
as …" even on an app whose `POST /api/account/login-link` turns keys away. Open
then fails. Since 2026-11-20 the portal words that refusal ("switch on Sign-in
links for that key… or it needs updating"). Each app that still answers 403
`this endpoint requires a signed-in session, not an API key` there needs the
fix below, one app at a time as they come up.

---

## Rook (`~/WORKSPACE/PROJECTS/Games/Rook`)

### 2026-11-22, fixed here: any admin may run accounts, so the portal's own admin works
Manage project answered "forbidden (403)": every `/api/admin/users*` route had
`use=[handlers.RequireOwner]`, so only the owner could manage accounts, and the
portal works through a "ProjectPortal" admin account of its own. Those routes
now use `RequireAccountAdmin` (`server/routes/admin_games.go`), which follows
the template's ladder (`models.MaySuperActOn`): any admin runs players'
accounts, nobody touches the owner's account, and only the owner touches
another admin's. Still the owner's alone: game settings, everyone's API keys,
key minting for accounts (`/api/admin/users/{id}/keys`), logs. Rook's own
Admin screen is still shown to the owner only. Tests: `server/routes/account_admin_test.go`.
Not committed or deployed yet.

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
