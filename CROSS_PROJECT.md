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
502), check it against this table first. The usual fix is the `webserver-upgrade`
skill run in that app.

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
