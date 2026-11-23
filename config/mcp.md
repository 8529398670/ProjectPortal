# ProjectPortal MCP tools

This file is generated: do not edit it. Change the tools next to their handlers and run `scripts/api_docs.sh`. Every route is in [api.md](api.md).

Every tool POST /mcp offers. Written by scripts/api_docs.sh from the registered tools; do not edit by hand. A key sees the tools its role allows (adminOnly ones need an admin key), and a tool whose feature is switched off answers like one that does not exist. route is the HTTP route the tool stands for, with the same rules; http_only are the routes a key may call that have no tool, and why.

|  |  |
|---|---|
| Endpoint | `POST /mcp` |
| Transport | Streamable HTTP, stateless: one JSON-RPC message per POST, a JSON reply |
| Auth | Authorization: Bearer <API key> |

| Tool | Route | Who | What |
|---|---|---|---|
| [`admin_act_on_project_accounts`](#admin_act_on_project_accounts) | [`POST /api/admin/users/{user_id}/project-accounts`](api.md#post-apiadminusersuser_idproject-accounts) | admin keys | Disable someone's accounts on every project (admin) |
| [`admin_cancel_login_link`](#admin_cancel_login_link) | [`POST /api/admin/users/{user_id}/cancel-login-link`](api.md#post-apiadminusersuser_idcancel-login-link) | admin keys | Stop an account's login link working (admin) |
| [`admin_cancel_login_link_bulk`](#admin_cancel_login_link_bulk) | [`POST /api/bulk/admin/users/cancel-login-link`](api.md#post-apibulkadminuserscancel-login-link) | admin keys | Stop an account's login link working, for many at once (admin) |
| [`admin_check_project`](#admin_check_project) | [`POST /api/admin/projects/{project_id}/check`](api.md#post-apiadminprojectsproject_idcheck) | admin keys | Check a project (admin) |
| [`admin_create_invite_link`](#admin_create_invite_link) | [`POST /api/admin/invite-links`](api.md#post-apiadmininvite-links) | admin keys | Create an invite link (admin) |
| [`admin_create_person_portal_user`](#admin_create_person_portal_user) | [`POST /api/admin/people/{person_id}/portal-user`](api.md#post-apiadminpeopleperson_idportal-user) | admin keys | Make a person a portal account (admin) |
| [`admin_create_remote_invite_link`](#admin_create_remote_invite_link) | [`POST /api/admin/projects/{project_id}/invite-links`](api.md#post-apiadminprojectsproject_idinvite-links) | admin keys | Make an invite link on a project (admin) |
| [`admin_create_remote_user`](#admin_create_remote_user) | [`POST /api/admin/projects/{project_id}/remote-users`](api.md#post-apiadminprojectsproject_idremote-users) | admin keys | Make an account on a project (admin) |
| [`admin_create_user`](#admin_create_user) | [`POST /api/admin/users`](api.md#post-apiadminusers) | admin keys | Create an account (admin) |
| [`admin_delete_invite_link`](#admin_delete_invite_link) | [`POST /api/admin/invite-links/{link_id}/delete`](api.md#post-apiadmininvite-linkslink_iddelete) | admin keys | Delete an invite link (admin) |
| [`admin_delete_invite_link_bulk`](#admin_delete_invite_link_bulk) | [`POST /api/bulk/admin/invite-links/delete`](api.md#post-apibulkadmininvite-linksdelete) | admin keys | Delete an invite link, for many at once (admin) |
| [`admin_delete_project`](#admin_delete_project) | [`POST /api/admin/projects/{project_id}/delete`](api.md#post-apiadminprojectsproject_iddelete) | admin keys | Delete a project (admin) |
| [`admin_delete_remote_user`](#admin_delete_remote_user) | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/delete`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_iddelete) | admin keys | Delete an account on a project (admin) |
| [`admin_delete_user`](#admin_delete_user) | [`POST /api/admin/users/{user_id}/delete`](api.md#post-apiadminusersuser_iddelete) | admin keys | Delete an account (admin) |
| [`admin_delete_user_bulk`](#admin_delete_user_bulk) | [`POST /api/bulk/admin/users/delete`](api.md#post-apibulkadminusersdelete) | admin keys | Delete an account, for many at once (admin) |
| [`admin_dismiss_people`](#admin_dismiss_people) | [`POST /api/admin/people/dismiss`](api.md#post-apiadminpeopledismiss) | admin keys | Dismiss a suggested match (admin) |
| [`admin_edit_person`](#admin_edit_person) | [`POST /api/admin/people/{person_id}`](api.md#post-apiadminpeopleperson_id) | admin keys | Rename a person (admin) |
| [`admin_get_theme_defaults`](#admin_get_theme_defaults) | [`GET /api/admin/theme-defaults`](api.md#get-apiadmintheme-defaults) | admin keys | Default themes (admin) |
| [`admin_grant_project`](#admin_grant_project) | [`POST /api/admin/projects/{project_id}/grant/{user_id}`](api.md#post-apiadminprojectsproject_idgrantuser_id) | admin keys | Put someone on a project (admin) |
| [`admin_link_people`](#admin_link_people) | [`POST /api/admin/people/link`](api.md#post-apiadminpeoplelink) | admin keys | Link accounts to one person (admin) |
| [`admin_link_remote_user`](#admin_link_remote_user) | [`POST /api/admin/projects/{project_id}/link/{user_id}`](api.md#post-apiadminprojectsproject_idlinkuser_id) | admin keys | Match someone to their account on a project (admin) |
| [`admin_list_catalog`](#admin_list_catalog) | [`GET /api/admin/projects`](api.md#get-apiadminprojects) | admin keys | The catalog (admin) |
| [`admin_list_invite_links`](#admin_list_invite_links) | [`GET /api/admin/invite-links`](api.md#get-apiadmininvite-links) | admin keys | List invite links (admin) |
| [`admin_list_people`](#admin_list_people) | [`GET /api/admin/people`](api.md#get-apiadminpeople) | admin keys | People across the catalog (admin) |
| [`admin_list_remote_invite_links`](#admin_list_remote_invite_links) | [`GET /api/admin/projects/{project_id}/invite-links`](api.md#get-apiadminprojectsproject_idinvite-links) | admin keys | A project's invite links (admin) |
| [`admin_list_remote_keys`](#admin_list_remote_keys) | [`GET /api/admin/projects/{project_id}/remote-users/{remote_user_id}/keys`](api.md#get-apiadminprojectsproject_idremote-usersremote_user_idkeys) | admin keys | An account's API keys on a project (admin) |
| [`admin_list_remote_users`](#admin_list_remote_users) | [`GET /api/admin/projects/{project_id}/remote-users`](api.md#get-apiadminprojectsproject_idremote-users) | admin keys | A project's own accounts (admin) |
| [`admin_list_user_projects`](#admin_list_user_projects) | [`GET /api/admin/users/{user_id}/projects`](api.md#get-apiadminusersuser_idprojects) | admin keys | Which projects someone may join (admin) |
| [`admin_list_user_sessions`](#admin_list_user_sessions) | [`GET /api/admin/users/{user_id}/sessions`](api.md#get-apiadminusersuser_idsessions) | admin keys | An account's signed-in devices (admin) |
| [`admin_list_users`](#admin_list_users) | [`GET /api/admin/users`](api.md#get-apiadminusers) | admin keys | List users (admin) |
| [`admin_merge_people`](#admin_merge_people) | [`POST /api/admin/people/merge`](api.md#post-apiadminpeoplemerge) | admin keys | Merge two people (admin) |
| [`admin_order_catalog`](#admin_order_catalog) | [`POST /api/admin/projects/order`](api.md#post-apiadminprojectsorder) | admin keys | Order the catalog (admin) |
| [`admin_project_access`](#admin_project_access) | [`GET /api/admin/projects/{project_id}/access`](api.md#get-apiadminprojectsproject_idaccess) | admin keys | Who may use a project (admin) |
| [`admin_reissue_login`](#admin_reissue_login) | [`POST /api/admin/users/{user_id}/reissue-login`](api.md#post-apiadminusersuser_idreissue-login) | admin keys | A new login link for an account (admin) |
| [`admin_remove_project_member`](#admin_remove_project_member) | [`POST /api/admin/projects/{project_id}/members/{user_id}/remove`](api.md#post-apiadminprojectsproject_idmembersuser_idremove) | admin keys | Take someone off a project here (admin) |
| [`admin_rename_invite_link`](#admin_rename_invite_link) | [`POST /api/admin/invite-links/{link_id}/rename`](api.md#post-apiadmininvite-linkslink_idrename) | admin keys | Rename an invite link (admin) |
| [`admin_rename_user`](#admin_rename_user) | [`POST /api/admin/users/{user_id}/rename`](api.md#post-apiadminusersuser_idrename) | admin keys | Rename someone (admin) |
| [`admin_reset_two_factor`](#admin_reset_two_factor) | [`POST /api/admin/users/{user_id}/reset-two-factor`](api.md#post-apiadminusersuser_idreset-two-factor) | admin keys | Turn off an account's two-factor (admin) |
| [`admin_reset_two_factor_bulk`](#admin_reset_two_factor_bulk) | [`POST /api/bulk/admin/users/reset-two-factor`](api.md#post-apibulkadminusersreset-two-factor) | admin keys | Turn off an account's two-factor, for many at once (admin) |
| [`admin_restore_invite_link`](#admin_restore_invite_link) | [`POST /api/admin/invite-links/{link_id}/restore`](api.md#post-apiadmininvite-linkslink_idrestore) | admin keys | Undo deleting an invite link (admin) |
| [`admin_restore_invite_link_bulk`](#admin_restore_invite_link_bulk) | [`POST /api/bulk/admin/invite-links/restore`](api.md#post-apibulkadmininvite-linksrestore) | admin keys | Undo deleting an invite link, for many at once (admin) |
| [`admin_restore_project`](#admin_restore_project) | [`POST /api/admin/projects/{project_id}/restore`](api.md#post-apiadminprojectsproject_idrestore) | admin keys | Undo deleting a project (admin) |
| [`admin_revoke_access`](#admin_revoke_access) | [`POST /api/admin/users/{user_id}/revoke-access`](api.md#post-apiadminusersuser_idrevoke-access) | admin keys | Revoke a user's access (admin) |
| [`admin_revoke_access_bulk`](#admin_revoke_access_bulk) | [`POST /api/bulk/admin/users/revoke-access`](api.md#post-apibulkadminusersrevoke-access) | admin keys | Revoke a user's access, for many at once (admin) |
| [`admin_revoke_invite_link`](#admin_revoke_invite_link) | [`POST /api/admin/invite-links/{link_id}/revoke`](api.md#post-apiadmininvite-linkslink_idrevoke) | admin keys | Switch off an invite link (admin) |
| [`admin_revoke_invite_link_bulk`](#admin_revoke_invite_link_bulk) | [`POST /api/bulk/admin/invite-links/revoke`](api.md#post-apibulkadmininvite-linksrevoke) | admin keys | Switch off an invite link, for many at once (admin) |
| [`admin_revoke_remote_access`](#admin_revoke_remote_access) | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/revoke-access`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_idrevoke-access) | admin keys | Sign an account out everywhere on a project (admin) |
| [`admin_revoke_remote_invite_link`](#admin_revoke_remote_invite_link) | [`POST /api/admin/projects/{project_id}/invite-links/{link_id}/revoke`](api.md#post-apiadminprojectsproject_idinvite-linkslink_idrevoke) | admin keys | Stop an invite link on a project (admin) |
| [`admin_revoke_remote_key`](#admin_revoke_remote_key) | [`POST /api/admin/projects/{project_id}/keys/{key_id}/revoke`](api.md#post-apiadminprojectsproject_idkeyskey_idrevoke) | admin keys | Revoke an API key on a project (admin) |
| [`admin_revoke_user_session`](#admin_revoke_user_session) | [`POST /api/admin/users/{user_id}/sessions/{session_id}/revoke`](api.md#post-apiadminusersuser_idsessionssession_idrevoke) | admin keys | Sign out one of an account's devices (admin) |
| [`admin_revoke_user_session_bulk`](#admin_revoke_user_session_bulk) | [`POST /api/bulk/admin/users/sessions/revoke`](api.md#post-apibulkadminuserssessionsrevoke) | admin keys | Sign out one of an account's devices, for many at once (admin) |
| [`admin_rotate_project_key`](#admin_rotate_project_key) | [`POST /api/admin/projects/{project_id}/rotate-key`](api.md#post-apiadminprojectsproject_idrotate-key) | admin keys | Rotate a project's key (admin) |
| [`admin_set_invite_link_limits`](#admin_set_invite_link_limits) | [`POST /api/admin/invite-links/{link_id}/limits`](api.md#post-apiadmininvite-linkslink_idlimits) | admin keys | Change an invite link's places and expiry (admin) |
| [`admin_set_project_access`](#admin_set_project_access) | [`POST /api/admin/projects/{project_id}/access/{user_id}`](api.md#post-apiadminprojectsproject_idaccessuser_id) | admin keys | Let someone use a project, or not (admin) |
| [`admin_set_project_access_bulk`](#admin_set_project_access_bulk) | [`POST /api/bulk/admin/projects/access`](api.md#post-apibulkadminprojectsaccess) | admin keys | Let someone use a project, or not, for many at once (admin) |
| [`admin_set_remote_disabled`](#admin_set_remote_disabled) | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/disabled`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_iddisabled) | admin keys | Disable or enable an account on a project (admin) |
| [`admin_set_remote_expiry`](#admin_set_remote_expiry) | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/expiry`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_idexpiry) | admin keys | Set when an account on a project expires (admin) |
| [`admin_set_remote_role`](#admin_set_remote_role) | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/role`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_idrole) | admin keys | Change an account's role on a project (admin) |
| [`admin_set_theme_defaults`](#admin_set_theme_defaults) | [`POST /api/admin/theme-defaults`](api.md#post-apiadmintheme-defaults) | admin keys | Replace the default themes (admin) |
| [`admin_set_user_disabled`](#admin_set_user_disabled) | [`POST /api/admin/users/{user_id}/disabled`](api.md#post-apiadminusersuser_iddisabled) | admin keys | Disable or enable a user (admin) |
| [`admin_set_user_disabled_bulk`](#admin_set_user_disabled_bulk) | [`POST /api/bulk/admin/users/disabled`](api.md#post-apibulkadminusersdisabled) | admin keys | Disable or enable a user, for many at once (admin) |
| [`admin_set_user_expiry`](#admin_set_user_expiry) | [`POST /api/admin/users/{user_id}/expiry`](api.md#post-apiadminusersuser_idexpiry) | admin keys | Set when an account expires (admin) |
| [`admin_set_user_role`](#admin_set_user_role) | [`POST /api/admin/users/{user_id}/role`](api.md#post-apiadminusersuser_idrole) | admin keys | Change an account's role (admin) |
| [`admin_set_user_role_bulk`](#admin_set_user_role_bulk) | [`POST /api/bulk/admin/users/role`](api.md#post-apibulkadminusersrole) | admin keys | Change an account's role, for many at once (admin) |
| [`admin_switch_on_invite_link`](#admin_switch_on_invite_link) | [`POST /api/admin/invite-links/{link_id}/switch-on`](api.md#post-apiadmininvite-linkslink_idswitch-on) | admin keys | Switch an invite link back on (admin) |
| [`admin_switch_on_invite_link_bulk`](#admin_switch_on_invite_link_bulk) | [`POST /api/bulk/admin/invite-links/switch-on`](api.md#post-apibulkadmininvite-linksswitch-on) | admin keys | Switch an invite link back on, for many at once (admin) |
| [`admin_sync_people`](#admin_sync_people) | [`POST /api/admin/people/sync`](api.md#post-apiadminpeoplesync) | admin keys | Refresh people now (admin) |
| [`admin_unlink_person`](#admin_unlink_person) | [`POST /api/admin/people/unlink`](api.md#post-apiadminpeopleunlink) | admin keys | Take an account off its person (admin) |
| [`cancel_my_login_link`](#cancel_my_login_link) | [`POST /api/account/login-link/cancel`](api.md#post-apiaccountlogin-linkcancel) | any key | Stop your own login link working |
| [`create_my_login_link`](#create_my_login_link) | [`POST /api/account/login-link`](api.md#post-apiaccountlogin-link) | any key | A login link for yourself |
| [`favorite_project`](#favorite_project) | [`POST /api/projects/{project_id}/favorite`](api.md#post-apiprojectsproject_idfavorite) | any key | Star or unstar a project |
| [`join_project`](#join_project) | [`POST /api/projects/{project_id}/join`](api.md#post-apiprojectsproject_idjoin) | any key | Join a project |
| [`label_my_session`](#label_my_session) | [`POST /api/account/sessions/{session_id}/label`](api.md#post-apiaccountsessionssession_idlabel) | any key | Name one of your devices |
| [`leave_project`](#leave_project) | [`POST /api/projects/{project_id}/leave`](api.md#post-apiprojectsproject_idleave) | any key | Leave a project |
| [`list_my_sessions`](#list_my_sessions) | [`GET /api/account/sessions`](api.md#get-apiaccountsessions) | any key | Your signed-in devices |
| [`list_projects`](#list_projects) | [`GET /api/projects`](api.md#get-apiprojects) | any key | Your projects |
| [`list_team`](#list_team) | [`GET /api/team`](api.md#get-apiteam) | any key | List team |
| [`open_project`](#open_project) | [`POST /api/projects/{project_id}/open`](api.md#post-apiprojectsproject_idopen) | any key | Sign in to a project |
| [`remove_own_project_key`](#remove_own_project_key) | [`POST /api/projects/{project_id}/own-key/remove`](api.md#post-apiprojectsproject_idown-keyremove) | any key | Stop using your own key on a project |
| [`rename_account`](#rename_account) | [`POST /api/account/rename`](api.md#post-apiaccountrename) | any key | Rename your account |
| [`restore_project_membership`](#restore_project_membership) | [`POST /api/projects/{project_id}/restore`](api.md#post-apiprojectsproject_idrestore) | any key | Undo leaving a project |
| [`revoke_my_other_sessions`](#revoke_my_other_sessions) | [`POST /api/account/sessions/revoke-others`](api.md#post-apiaccountsessionsrevoke-others) | any key | Sign out all your devices |
| [`revoke_my_session`](#revoke_my_session) | [`POST /api/account/sessions/{session_id}/revoke`](api.md#post-apiaccountsessionssession_idrevoke) | any key | Sign out one of your devices |
| [`revoke_my_session_bulk`](#revoke_my_session_bulk) | [`POST /api/bulk/account/sessions/revoke`](api.md#post-apibulkaccountsessionsrevoke) | any key | Sign out one of your devices, for many at once |
| [`set_shortcuts`](#set_shortcuts) | [`POST /api/account/shortcuts`](api.md#post-apiaccountshortcuts) | any key | Change your keyboard shortcuts |
| [`set_theme`](#set_theme) | [`POST /api/account/theme`](api.md#post-apiaccounttheme) | any key | Pick your theme |
| [`set_themes`](#set_themes) | [`POST /api/account/themes`](api.md#post-apiaccountthemes) | any key | Replace your own themes |
| [`tag_project`](#tag_project) | [`POST /api/projects/{project_id}/tags`](api.md#post-apiprojectsproject_idtags) | any key | Tag a project |
| [`whoami`](#whoami) | [`GET /api/me`](api.md#get-apime) | any key | Who am I |

## admin_act_on_project_accounts

For every project they are on, disable their account there (on_project disable) or end its sessions, links and keys (revoke). Use it before disabling or deleting them here. Answers how each project took it. Only the owner does it to an admin; nobody does it to the owner.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/project-accounts`](api.md#post-apiadminusersuser_idproject-accounts) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `on_project` | string | yes | one of `disable`, `revoke` | Disable their account on every project, or end its sessions, links and keys there. |
| `user_id` | integer | yes | at least 1 | The account's id here. |

## admin_cancel_login_link

Void an account's live login link without making a new one: a link that went somewhere it should not have. Devices it already signed in stay signed in; admin_revoke_access ends those too. Answers with removed: 0 or 1.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/cancel-login-link`](api.md#post-apiadminusersuser_idcancel-login-link) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_cancel_login_link_bulk

Runs admin_cancel_login_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/cancel-login-link`](api.md#post-apibulkadminuserscancel-login-link) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_check_project

Ask the project now whether it is up and whether the portal's admin key still works there. The project-health job does this every 5 minutes.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/check`](api.md#post-apiadminprojectsproject_idcheck) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## admin_create_invite_link

One link for a group: each person who opens it joins as a new account with role (user, the default, or viewer), up to max_uses people, for days days. Answers with join_url (shown only now) and the link's details. Revoking or deleting this API key switches the link off.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links`](api.md#post-apiadmininvite-links) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `days` | integer | yes | 1 to 365 |  |
| `label` | string | yes | 1 to 80 characters |  |
| `max_uses` | integer | yes | 1 to 1000 |  |
| `role` | string | no | one of `user`, `viewer`; default `user` | The role each person who joins gets: user, or viewer (reads only). |

## admin_create_person_portal_user

For someone from before the portal: make them an account here, put it under the person, and make each of their app accounts its own (as matching would), so Open signs them into the accounts they have. Answers with user_id and its first login link. A person who already has an account here is 409: link that one instead. Answers with login_url (open it to sign in; it works once unless max_uses says otherwise).

|  |  |
|---|---|
| Route | [`POST /api/admin/people/{person_id}/portal-user`](api.md#post-apiadminpeopleperson_idportal-user) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `account_expires_at` | string (date-time) or null | no |  | When the account stops working (RFC 3339). Left out: never. |
| `display_name` | string | no | at most 80 characters | The account's name here. Left out: the person's name. |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |
| `person_id` | integer | yes | at least 1 |  |
| `role` | string | yes | one of `admin`, `user`, `viewer` |  |

## admin_create_remote_invite_link

One link a group can join the project itself through: up to max_uses accounts, for days days. role viewer makes them view only, on an app that has it. Answers join_url: hand it out; it is shown only once.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/invite-links`](api.md#post-apiadminprojectsproject_idinvite-links) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `days` | integer | yes | 1 to 365 |  |
| `label` | string | yes | 1 to 80 characters |  |
| `max_uses` | integer | yes | 1 to 1000 |  |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `role` | string | no | one of `user`, `viewer` | What people who join through it become. Default user. |

## admin_create_remote_user

Make an account on the project itself, with a role and an optional expiry, and get its login link. With user_id, that person here is put on the project with it. An admin account takes a browser session: an API key gets 403 session_only. Answers login_url: hand it to the person it is for (it works once, for 15 minutes).

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/remote-users`](api.md#post-apiadminprojectsproject_idremote-users) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes | 1 to 80 characters |  |
| `expires_in_days` | integer | no | 0 to 3650 | The account stops working this many days from now. 0 or left out: never. Needs an app with accounts that expire. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `role` | string | yes | one of `user`, `viewer`, `admin` | viewer needs an app with view-only accounts. |
| `user_id` | integer | no | at least 0 | Someone here to put on the project with it, so their Open signs into it. Left out: a guest with no account here. |

## admin_create_user

Make a new account and its first login link, to hand to the person. role is admin, user or viewer (reads only). account_expires_at, when given, is when the account stops working. Answers with user_id and the link. Answers with login_url (open it to sign in; it works once unless max_uses says otherwise), expires_at and max_uses. Any older link the account could still use stops working; its signed-in devices stay signed in. Revoking or deleting this API key voids the link while it is unused, and signs out every device it signed in.

|  |  |
|---|---|
| Route | [`POST /api/admin/users`](api.md#post-apiadminusers) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `account_expires_at` | string (date-time) or null | no |  | When the account stops working (RFC 3339). Left out: never. expires_in_minutes is the login link's expiry, not this. |
| `display_name` | string | yes | 1 to 80 characters |  |
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `kind` | string | no | one of `person`, `service`; default `person` | person, the default. service is refused (403 service_account): a service account is made by Connect on the owner's API keys screen. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |
| `role` | string | yes | one of `admin`, `user`, `viewer` |  |

## admin_delete_invite_link

Take an invite link off the list and stop it working. The accounts it made stay. admin_restore_invite_link undoes it for a few minutes.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/delete`](api.md#post-apiadmininvite-linkslink_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_delete_invite_link_bulk

Runs admin_delete_invite_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/invite-links/delete`](api.md#post-apibulkadmininvite-linksdelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_delete_project

Take a project out of the portal, with its key and everyone's membership of it. Accounts on the project itself stay. admin_restore_project undoes it within 10 minutes.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/delete`](api.md#post-apiadminprojectsproject_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## admin_delete_remote_user

Delete the account on the project itself, with everything it held there. Anyone here it belonged to comes off the project. It cannot be undone.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/delete`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

## admin_delete_user

Remove an account for good, with its sessions, login link and API keys. Cannot be undone: admin_set_user_disabled is the reversible version. Only the owner deletes an admin; nobody deletes the owner or themselves.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/delete`](api.md#post-apiadminusersuser_iddelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_delete_user_bulk

Runs admin_delete_user once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/delete`](api.md#post-apibulkadminusersdelete) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_dismiss_people

The accounts in a are not the person of those in b: that match is never suggested again.

|  |  |
|---|---|
| Route | [`POST /api/admin/people/dismiss`](api.md#post-apiadminpeopledismiss) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `a` | array of object | yes |  |  |
| `a[].project_id` | integer | yes | at least 0 | The project, or 0 for this portal. |
| `a[].remote_user_id` | integer | yes | at least 1 | The account's id there. |
| `b` | array of object | yes |  |  |
| `b[].project_id` | integer | yes | at least 0 | The project, or 0 for this portal. |
| `b[].remote_user_id` | integer | yes | at least 1 | The account's id there. |

## admin_edit_person

The name and notes the portal keeps for a person. Their accounts keep their own names.

|  |  |
|---|---|
| Route | [`POST /api/admin/people/{person_id}`](api.md#post-apiadminpeopleperson_id) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes | 1 to 80 characters |  |
| `notes` | string | yes | at most 2000 characters |  |
| `person_id` | integer | yes | at least 1 |  |

## admin_get_theme_defaults

The theme list everyone without their own list sees. themes is null while the built-in ones apply.

|  |  |
|---|---|
| Route | [`GET /api/admin/theme-defaults`](api.md#get-apiadmintheme-defaults) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## admin_grant_project

Make someone an account on the project now, as if they had joined, optionally view only or expiring, and allow them the project. They sign in from their own list; no link is handed to you.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/grant/{user_id}`](api.md#post-apiadminprojectsproject_idgrantuser_id) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_in_days` | integer | no | 0 to 3650 | Their account there stops working this many days from now. 0 or left out: never. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `role` | string | no | one of `user`, `viewer` | Their role on the project. Default user; viewer (view only) needs an app that has it. |
| `user_id` | integer | yes | at least 1 | The account's id here. |

## admin_link_people

Put accounts under one person. People they already belong to are merged into it. Service accounts are refused: they belong to nobody. When the person then has one account here, their app accounts become its own (memberships_added), as matching would.

|  |  |
|---|---|
| Route | [`POST /api/admin/people/link`](api.md#post-apiadminpeoplelink) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `accounts` | array of object | yes |  | The accounts that are one person. |
| `accounts[].project_id` | integer | yes | at least 0 | The project, or 0 for this portal. |
| `accounts[].remote_user_id` | integer | yes | at least 1 | The account's id there. |
| `person_id` | integer | no |  | Put them under this person; 0 to use the person one of them has, or a new one. |

## admin_link_remote_user

Say which account on the project is theirs. Opening the project then signs them into it. The project's owner, and its admins unless the portal's key is the owner's, cannot be signed in this way: they attach their own key instead. An account already someone else's here is 409. Only the owner links another admin; nobody links the owner. It also allows them the project.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/link/{user_id}`](api.md#post-apiadminprojectsproject_idlinkuser_id) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | integer | yes | at least 1 | Their account's id on the project, from admin_list_remote_users. |
| `user_id` | integer | yes | at least 1 | The account's id here. |

## admin_list_catalog

Every project in the portal, a page at a time, with whether its admin key works, whether it is up, how many people are on it, and its GitHub and TodoBoard links.

|  |  |
|---|---|
| Route | [`GET /api/admin/projects`](api.md#get-apiadminprojects) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep projects whose name or address contains this, in any case. |
| `sort` | string | no | one of `name`, `base_url`, `health`, `members`, `position` | Field to sort by. Default the catalog's order. |

## admin_list_invite_links

Group invite links, a page at a time: label, places used and left, expiry, and whether each still works. The links themselves are never shown again after creation. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/admin/invite-links`](api.md#get-apiadmininvite-links) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep links whose label or maker contains this, in any case. |
| `sort` | string | no | one of `label`, `used`, `expires_at` | Field to sort the whole list by. Rows without it go last. |

## admin_list_people

Every person and their accounts in each app (project 0 is this portal), the accounts nobody has placed, service accounts, and suggested matches with why. From the last people-sync.

|  |  |
|---|---|
| Route | [`GET /api/admin/people`](api.md#get-apiadminpeople) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## admin_list_remote_invite_links

The invite links on the project itself. Needs an app with invite links switched on.

|  |  |
|---|---|
| Route | [`GET /api/admin/projects/{project_id}/invite-links`](api.md#get-apiadminprojectsproject_idinvite-links) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## admin_list_remote_keys

The account's API keys there, never their secrets. Needs an app where the portal's key may make keys.

|  |  |
|---|---|
| Route | [`GET /api/admin/projects/{project_id}/remote-users/{remote_user_id}/keys`](api.md#get-apiadminprojectsproject_idremote-usersremote_user_idkeys) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

## admin_list_remote_users

The accounts on the project itself, as its admin key sees them, 200 at a time.

|  |  |
|---|---|
| Route | [`GET /api/admin/projects/{project_id}/remote-users`](api.md#get-apiadminprojectsproject_idremote-users) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `offset` | integer | no | at least 0 | From the previous page's next. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `q` | string | no |  | Keep accounts whose name contains this. |

## admin_list_user_projects

Every project, a page at a time, with whether this person may see and join it and whether they are on it. Projects are off for everyone by default; switch one on with admin_set_project_access.

|  |  |
|---|---|
| Route | [`GET /api/admin/users/{user_id}/projects`](api.md#get-apiadminusersuser_idprojects) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep projects whose name contains this, in any case. |
| `sort` | string | no | one of `name`, `allowed`, `member`, `position` | Field to sort by. Default the catalog's order. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_list_user_sessions

Where an account is signed in: every session, with when it started and was last seen. Addresses are shown to the owner and service accounts only. Nobody sees the owner's, and only the owner sees another admin's.

|  |  |
|---|---|
| Route | [`GET /api/admin/users/{user_id}/sessions`](api.md#get-apiadminusersuser_idsessions) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_list_users

Accounts, a page at a time, including disabled ones, with each one's role and which is the owner. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/admin/users`](api.md#get-apiadminusers) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep accounts whose name or username contains this, in any case. |
| `sort` | string | no | one of `display_name`, `role`, `disabled`, `created_at` | Field to sort the whole list by. Rows without it go last. |

## admin_merge_people

Every account of from goes to into, with its notes, and from is deleted.

|  |  |
|---|---|
| Route | [`POST /api/admin/people/merge`](api.md#post-apiadminpeoplemerge) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `from` | integer | yes | at least 1 | The person folded into them, then gone. |
| `into` | integer | yes | at least 1 | The person who stays. |

## admin_order_catalog

Put projects in this order everywhere they are listed. Ids left out follow.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/order`](api.md#post-apiadminprojectsorder) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_ids` | array of integer | yes |  | Project ids in the order people should see them. |

## admin_project_access

Every account here, a page at a time, with whether it may use the project and whether it is on it.

|  |  |
|---|---|
| Route | [`GET /api/admin/projects/{project_id}/access`](api.md#get-apiadminprojectsproject_idaccess) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `dir` | string | no | one of `asc`, `desc` | Default asc. |
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `q` | string | no |  | Keep people whose name contains this, in any case. |
| `sort` | string | no | one of `display_name`, `allowed`, `member` | Field to sort by. |

## admin_reissue_login

For someone who lost their device or their link: a fresh login link for their account. Answers with login_url (open it to sign in; it works once unless max_uses says otherwise), expires_at and max_uses. Any older link the account could still use stops working; its signed-in devices stay signed in. Revoking or deleting this API key voids the link while it is unused, and signs out every device it signed in.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/reissue-login`](api.md#post-apiadminusersuser_idreissue-login) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_remove_project_member

Forget someone's account on the project, and their own key for it. on_project says what happens to the account there: keep (default), disable, or revoke its sessions, links and keys. Only the owner takes another admin off; nobody takes the owner off.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/members/{user_id}/remove`](api.md#post-apiadminprojectsproject_idmembersuser_idremove) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `on_project` | string | no | one of `keep`, `disable`, `revoke` | What happens to their account on the project itself: keep it (the default), disable it, or end its sessions, links and keys. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `user_id` | integer | yes | at least 1 | The account's id here. |

## admin_rename_invite_link

Change the label people see on an invite link's join page (1-80 characters). expect, when given, is the label you think it has now.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/rename`](api.md#post-apiadmininvite-linkslink_idrename) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expect` | string or null | no |  | The label you think it has. Answers 409 if it has another. |
| `label` | string | yes | 1 to 80 characters | 1-80 characters. |
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_rename_user

Change another account's display name (1-80 characters). Only the owner renames an admin. expect, when given, is the name you think it has now; if it changed meanwhile the rename is refused with the current name.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/rename`](api.md#post-apiadminusersuser_idrename) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes | 1 to 80 characters | 1-80 characters. |
| `expect` | string or null | no |  | The name you think they have. Answers 409 if it is another. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_reset_two_factor

For someone who lost both their phone and their recovery codes: switches their two-factor sign-in off so a login link lets them in again. Only the owner resets an admin's; the owner's is reset from a shell.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/reset-two-factor`](api.md#post-apiadminusersuser_idreset-two-factor) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_reset_two_factor_bulk

Runs admin_reset_two_factor once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/reset-two-factor`](api.md#post-apibulkadminusersreset-two-factor) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_restore_invite_link

Bring back an invite link deleted in the last few minutes, working as before. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/restore`](api.md#post-apiadmininvite-linkslink_idrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_restore_invite_link_bulk

Runs admin_restore_invite_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/invite-links/restore`](api.md#post-apibulkadmininvite-linksrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_restore_project

Put back a project deleted in the last 10 minutes.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/restore`](api.md#post-apiadminprojectsproject_idrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## admin_revoke_access

For an account that may be compromised: ends every session and voids its live login link and every API key it holds, at once. The account stays enabled; admin_reissue_login lets its owner back in. The owner and your own account cannot be revoked.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/revoke-access`](api.md#post-apiadminusersuser_idrevoke-access) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_revoke_access_bulk

Runs admin_revoke_access once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/revoke-access`](api.md#post-apibulkadminusersrevoke-access) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_revoke_invite_link

Stop an invite link from letting anyone else join or sign in through it. Accounts it already made keep working. Cannot be switched back on.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/revoke`](api.md#post-apiadmininvite-linkslink_idrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_revoke_invite_link_bulk

Runs admin_revoke_invite_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/invite-links/revoke`](api.md#post-apibulkadmininvite-linksrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_revoke_remote_access

End every session, login link and API key the account has there. The account stays enabled; a new login link signs it back in.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/revoke-access`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_idrevoke-access) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

## admin_revoke_remote_invite_link

The link stops working. Accounts already made through it stay.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/invite-links/{link_id}/revoke`](api.md#post-apiadminprojectsproject_idinvite-linkslink_idrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes | 1 to 100 characters | The link's id, from admin_list_remote_invite_links. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## admin_revoke_remote_key

The key stops working there now, and so does every key it made.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/keys/{key_id}/revoke`](api.md#post-apiadminprojectsproject_idkeyskey_idrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `key_id` | string | yes | 1 to 100 characters | The key's id, from admin_list_remote_keys. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## admin_revoke_user_session

End one of an account's sessions, rather than all of them with admin_revoke_access.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/sessions/{session_id}/revoke`](api.md#post-apiadminusersuser_idsessionssession_idrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_revoke_user_session_bulk

Runs admin_revoke_user_session once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/sessions/revoke`](api.md#post-apibulkadminuserssessionsrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_rotate_project_key

Swap the portal's admin key on the project for a new secret, now. The key keeps its grants and its expiry there, and what it made keeps working. Needs an app with key rotation; the key-rotation job also does this every catalog.rotate_key_days.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/rotate-key`](api.md#post-apiadminprojectsproject_idrotate-key) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## admin_set_invite_link_limits

Set how many people the link may take in (never fewer than have joined) and, when days is given, a new expiry counted from now. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/limits`](api.md#post-apiadmininvite-linkslink_idlimits) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `days` | integer or null | no | 1 to 365 | Counted from now. Leave out to keep the expiry. |
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `max_uses` | integer | yes | 1 to 1000 | Not fewer than have joined. |

## admin_set_project_access

Projects are off for everyone by default. allowed true lets one person see and join a listed project; allowed false keeps it from them again: it leaves their list and the portal stops signing them in. Their account on the project itself stays unless on_project says otherwise. Admins always see everything.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/access/{user_id}`](api.md#post-apiadminprojectsproject_idaccessuser_id) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `allowed` | boolean | yes |  | true lets them see and join the project. false keeps it from them again: it leaves their list, and the portal will not sign them into it. |
| `on_project` | string | no | one of `keep`, `disable`, `revoke` | With allowed false, what happens to an account they have on the project: keep it (the default), disable it, or end its sessions, links and keys. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `user_id` | integer | yes | at least 1 | The account's id here, from admin_list_users. |

## admin_set_project_access_bulk

Runs admin_set_project_access once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/projects/access`](api.md#post-apibulkadminprojectsaccess) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].allowed` | boolean | no |  | true lets them see and join the project. false keeps it from them again: it leaves their list, and the portal will not sign them into it. |
| `items[].on_project` | string | no | one of `keep`, `disable`, `revoke` | With allowed false, what happens to an account they have on the project: keep it (the default), disable it, or end its sessions, links and keys. |
| `items[].project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `items[].user_id` | integer | yes | at least 1 | The account's id here, from admin_list_users. |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.allowed` | boolean | no |  | true lets them see and join the project. false keeps it from them again: it leaves their list, and the portal will not sign them into it. |
| `shared.on_project` | string | no | one of `keep`, `disable`, `revoke` | With allowed false, what happens to an account they have on the project: keep it (the default), disable it, or end its sessions, links and keys. |

## admin_set_remote_disabled

disabled true stops the account signing in there, and ends its sessions.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/disabled`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_iddisabled) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `disabled` | boolean | yes |  |  |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

## admin_set_remote_expiry

The account stops working expires_in_days from now; 0 means never. Needs an app with accounts that expire.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/expiry`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_idexpiry) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_in_days` | integer | yes | 0 to 3650 | Stops working this many days from now. 0: never. |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |

## admin_set_remote_role

admin, user, or viewer (view only, on an app that has it). The project's owner cannot be changed. Making someone admin takes a browser session: an API key gets 403 session_only.

|  |  |
|---|---|
| Route | [`POST /api/admin/projects/{project_id}/remote-users/{remote_user_id}/role`](api.md#post-apiadminprojectsproject_idremote-usersremote_user_idrole) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `remote_user_id` | integer | yes | at least 1 | The account's id on the project, from admin_list_remote_users. |
| `role` | string | yes | one of `user`, `viewer`, `admin` |  |

## admin_set_theme_defaults

Replace the theme list everyone without their own list sees, 1 to 40 themes in the Theme schema of the API reference. null puts the built-in themes back.

|  |  |
|---|---|
| Route | [`POST /api/admin/theme-defaults`](api.md#post-apiadmintheme-defaults) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | no |  | Which built-in themes this list has seen: the highest since in Theme.builtIn (static/js/theme.js) when it was saved. Built-in themes added after that join the list the browser shows. Left out: 0, the first ones. |
| `themes` | array of object or null | no |  | 1 to 40 themes. null goes back to the built-in ones. |
| `themes[].accent` | string | yes |  | The main color: what you can press. |
| `themes[].background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].button` | object or null | no |  | How the theme's buttons look. A choice left out is the default (rounded, normal, raised, normal). |
| `themes[].button.case` | string | no | one of `normal`, `upper` |  |
| `themes[].button.depth` | string | no | one of `flat`, `raised`, `glossy` |  |
| `themes[].button.shape` | string | no | one of `square`, `soft`, `rounded`, `pill` |  |
| `themes[].button.weight` | string | no | one of `normal`, `medium`, `bold` |  |
| `themes[].dark` | object or null | no |  | A light theme's colors for dark mode, kept while it is light. Switching the theme to dark swaps them in. A light theme's light (or a dark theme's dark) is dropped. |
| `themes[].dark.accent` | string | yes |  | The main color: what you can press. |
| `themes[].dark.background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].dark.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].dark.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].dark.surface` | string | no |  | Cards. Left out: the mode's own. |
| `themes[].id` | string | yes |  | Lower-case letters, digits, - and _, up to 32. |
| `themes[].light` | object or null | no |  | A dark theme's colors for light mode, kept while it is dark. Switching the theme to light swaps them in. |
| `themes[].light.accent` | string | yes |  | The main color: what you can press. |
| `themes[].light.background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].light.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].light.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].light.surface` | string | no |  | Cards. Left out: the mode's own. |
| `themes[].mode` | string | yes | one of `light`, `dark` |  |
| `themes[].name` | string | yes |  | 1 to 60 characters. |
| `themes[].overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].surface` | string | no |  | Cards. Left out: the mode's own. |

## admin_set_user_disabled

Switch an account off (disabled: true) or back on (false). Disabling signs it out everywhere and voids its live login link; its API keys stop working while it is off. Reversible. The owner and your own account cannot be disabled.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/disabled`](api.md#post-apiadminusersuser_iddisabled) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `disabled` | boolean | yes |  |  |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_set_user_disabled_bulk

Runs admin_set_user_disabled once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/disabled`](api.md#post-apibulkadminusersdisabled) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].disabled` | boolean | no |  |  |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.disabled` | boolean | no |  |  |

## admin_set_user_expiry

expires_at is when the account stops working (RFC 3339), or null for never. Once it passes the account counts as disabled: sign-in, its sessions and its keys stop, and housekeeping disables it and revokes its access. A later expiry does not re-enable an account housekeeping already disabled. It must be in the future. The owner's account never expires, nobody sets their own, and only the owner sets an admin's.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/expiry`](api.md#post-apiadminusersuser_idexpiry) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_at` | string (date-time) or null | yes |  | When the account stops working (RFC 3339), or null for never. |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_set_user_role

Make an account an admin, an ordinary user, or a viewer, who may read but change nothing. Only the owner takes an admin's rights away; the owner is always an admin, and nobody changes their own role. Becoming an admin ends the account's sessions and live login link.

|  |  |
|---|---|
| Route | [`POST /api/admin/users/{user_id}/role`](api.md#post-apiadminusersuser_idrole) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `role` | string | yes | one of `admin`, `user`, `viewer` |  |
| `user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |

## admin_set_user_role_bulk

Runs admin_set_user_role once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/users/role`](api.md#post-apibulkadminusersrole) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].role` | string | no | one of `admin`, `user`, `viewer` |  |
| `items[].user_id` | integer | yes | at least 1 | The account's id, from GET /api/admin/users (admin_list_users over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |
| `shared.role` | string | no | one of `admin`, `user`, `viewer` |  |

## admin_switch_on_invite_link

Undo admin_revoke_invite_link: anyone still holding the link can use it again. The link then counts as this key's: revoking the key switches it off.

|  |  |
|---|---|
| Route | [`POST /api/admin/invite-links/{link_id}/switch-on`](api.md#post-apiadmininvite-linkslink_idswitch-on) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |

## admin_switch_on_invite_link_bulk

Runs admin_switch_on_invite_link once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/admin/invite-links/switch-on`](api.md#post-apibulkadmininvite-linksswitch-on) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].link_id` | string | yes |  | The link's id, from GET /api/admin/invite-links (admin_list_invite_links over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## admin_sync_people

Start the people-sync job now: every app's accounts and devices, and this portal's. It runs in the background; list people again for the result. It also runs every hour.

|  |  |
|---|---|
| Route | [`POST /api/admin/people/sync`](api.md#post-apiadminpeoplesync) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## admin_unlink_person

The account no longer counts as that person's, and is not suggested for them again.

|  |  |
|---|---|
| Route | [`POST /api/admin/people/unlink`](api.md#post-apiadminpeopleunlink) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 0 | The project, or 0 for this portal. |
| `remote_user_id` | integer | yes | at least 1 | The account's id there. |

## cancel_my_login_link

Void your account's live login link without making a new one. Devices it already signed in stay signed in. Answers with removed: 0 or 1.

|  |  |
|---|---|
| Route | [`POST /api/account/login-link/cancel`](api.md#post-apiaccountlogin-linkcancel) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## create_my_login_link

A link that signs your own account in on a browser or another device. An API key may ask only if your account let it, from a browser session (POST /api/keys/{key_id}/sign-in); its link works once, for at most 15 minutes. Only works if this key was allowed to make sign-in links, on the API keys card. Answers with login_url (open it to sign in; it works once, for at most 15 minutes), expires_at and max_uses. Any older link the account could still use stops working; its signed-in devices stay signed in. Revoking or deleting this API key voids the link while it is unused, and signs out every device it signed in.

|  |  |
|---|---|
| Route | [`POST /api/account/login-link`](api.md#post-apiaccountlogin-link) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expires_in_minutes` | integer or null | no | 0 to 525600 | 0 never expires. Default auth.login_token_ttl_seconds for a link that works once, auth.multi_use_login_token_ttl_seconds otherwise. |
| `max_uses` | integer or null | no | 0 to 1000 | Sign-ins the link allows: 1-1000, or 0 for unlimited. Default 1. |

## favorite_project

favorite true keeps the project at the top of your list.

|  |  |
|---|---|
| Route | [`POST /api/projects/{project_id}/favorite`](api.md#post-apiprojectsproject_idfavorite) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `favorite` | boolean | yes |  |  |
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## join_project

Make yourself an account on a project you may use, named as you are here, and get a login link for it. An API key needs Sign-in links switched on for it (403 otherwise). 409 when you are already on it or it is not set up yet; 424 when the project refused or did not answer, with what it said. Only works if this key was allowed to make sign-in links, on the API keys card. Answers login_url: give it to the person to open (it works once, for 15 minutes).

|  |  |
|---|---|
| Route | [`POST /api/projects/{project_id}/join`](api.md#post-apiprojectsproject_idjoin) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## label_my_session

Give one of your devices a name (at most 80 characters; empty clears it). expect, when given, is the name you think it has now.

|  |  |
|---|---|
| Route | [`POST /api/account/sessions/{session_id}/label`](api.md#post-apiaccountsessionssession_idlabel) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `expect` | string or null | no |  | The name you think it has. Answers 409 if it has another. |
| `label` | string | yes | at most 80 characters | At most 80 characters. |
| `session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

## leave_project

Take a project off your list. Your account on the project stays as it is; joining again signs you back into it. Undo with restore_project_membership within 10 minutes keeps your stars, tags and key.

|  |  |
|---|---|
| Route | [`POST /api/projects/{project_id}/leave`](api.md#post-apiprojectsproject_idleave) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## list_my_sessions

Where your account is signed in, a page at a time: id, when it started and was last seen, the browser, and the name you gave it. Through an API key no device is current. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/account/sessions`](api.md#get-apiaccountsessions) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

## list_projects

Every project you may use, a page at a time: the ones you are on first (favorites first), then the ones you could join. member says which. Use open_project with a project you are a member of to get a login link for it, and join_project for one you are not.

|  |  |
|---|---|
| Route | [`GET /api/projects`](api.md#get-apiprojects) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |
| `q` | string | no |  | Keep projects whose name, description or one of your tags contains this, in any case. |

## list_team

People with an active account, a page at a time: id and display name. Use the id wherever another tool asks for a user. next is the offset of the following page, or null on the last one.

|  |  |
|---|---|
| Route | [`GET /api/team`](api.md#get-apiteam) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `limit` | integer | no | 1 to 1000; default `200` | Rows to return. Default 200. |
| `offset` | integer | no | at least 0 | Rows to skip: the next value of the previous page. |

## open_project

A fresh login link for your account on a project you are on: with your own key when you attached one, otherwise made by the portal's admin key there. An API key needs Sign-in links switched on for it (403 otherwise). 409 when you are not on it yet (join it), 424 when the project refused, with what it said. Only works if this key was allowed to make sign-in links, on the API keys card. Answers login_url: give it to the person to open (it works once, for 15 minutes).

|  |  |
|---|---|
| Route | [`POST /api/projects/{project_id}/open`](api.md#post-apiprojectsproject_idopen) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## remove_own_project_key

Forget your own key on that project. If the key was all that put you on it, you are no longer on it here.

|  |  |
|---|---|
| Route | [`POST /api/projects/{project_id}/own-key/remove`](api.md#post-apiprojectsproject_idown-keyremove) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## rename_account

Change your own display name (1-80 characters). expect, when given, is the name you think it has now; if someone changed it meanwhile the rename is refused with the current name, so nothing is overwritten unseen.

|  |  |
|---|---|
| Route | [`POST /api/account/rename`](api.md#post-apiaccountrename) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `display_name` | string | yes | 1 to 80 characters |  |
| `expect` | string or null | no |  |  |

## restore_project_membership

Put back a project you left in the last 10 minutes, as it was.

|  |  |
|---|---|
| Route | [`POST /api/projects/{project_id}/restore`](api.md#post-apiprojectsproject_idrestore) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |

## revoke_my_other_sessions

End every session your account has: through an API key there is no current device to keep, so every browser signs in again. API keys keep working. Answers with how many were removed.

|  |  |
|---|---|
| Route | [`POST /api/account/sessions/revoke-others`](api.md#post-apiaccountsessionsrevoke-others) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## revoke_my_session

End one of your signed-in sessions; that browser has to sign in again.

|  |  |
|---|---|
| Route | [`POST /api/account/sessions/{session_id}/revoke`](api.md#post-apiaccountsessionssession_idrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |

## revoke_my_session_bulk

Runs revoke_my_session once per item: items holds each call's arguments, shared the arguments they have in common (an item's own win). In order; one failing does not stop the rest, and the reply says how each went. At most 100 items.

|  |  |
|---|---|
| Route | [`POST /api/bulk/account/sessions/revoke`](api.md#post-apibulkaccountsessionsrevoke) |
| Read only | no |
| Destructive | yes |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `items` | array of object | yes |  | One entry per thing to change: its id, and any field that differs from shared. |
| `items[].session_id` | string | yes |  | The device's id, from GET /api/account/sessions or GET /api/admin/users/{user_id}/sessions (list_my_sessions or admin_list_user_sessions over MCP). |
| `shared` | object | no |  | Fields every item gets, unless the item sets its own. |

## set_shortcuts

Replace your changed keyboard shortcuts: action name to binding, such as "search": "Mod+K" (Mod is Cmd on a Mac, Ctrl elsewhere). Send the whole set; an empty binding switches a key off, and {} puts every key back to its default. whoami shows the current set.

|  |  |
|---|---|
| Route | [`POST /api/account/shortcuts`](api.md#post-apiaccountshortcuts) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `shortcuts` | map of string | yes |  | Action name to binding, such as "search": "Mod+K". Mod is Cmd on a Mac and Ctrl elsewhere. Only changed shortcuts are listed; an empty binding switches that key off. Replaces the whole set; {} puts every key back to its default. |

## set_theme

Switch to another theme from your list (or the defaults), by its id.

|  |  |
|---|---|
| Route | [`POST /api/account/theme`](api.md#post-apiaccounttheme) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `theme_id` | string | yes |  | A theme's id. Empty for the first in the list. |

## set_themes

Replace your whole list of color themes (1 to 40) and pick the one you use. themes: null goes back to following the app's defaults. A theme is the Theme schema in the API reference (config/api.json): id, name, mode light|dark, accent, second, and optional colors as #rrggbb. whoami shows the current list.

|  |  |
|---|---|
| Route | [`POST /api/account/themes`](api.md#post-apiaccountthemes) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `builtins` | integer | no |  | Which built-in themes this list has seen: the highest since in Theme.builtIn (static/js/theme.js) when it was saved. Built-in themes added after that join the list the browser shows. Left out: 0, the first ones. |
| `theme_id` | string | no |  | The theme you use. |
| `themes` | array of object or null | no |  | Your whole list, 1 to 40 themes. null goes back to following the app's defaults. |
| `themes[].accent` | string | yes |  | The main color: what you can press. |
| `themes[].background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].button` | object or null | no |  | How the theme's buttons look. A choice left out is the default (rounded, normal, raised, normal). |
| `themes[].button.case` | string | no | one of `normal`, `upper` |  |
| `themes[].button.depth` | string | no | one of `flat`, `raised`, `glossy` |  |
| `themes[].button.shape` | string | no | one of `square`, `soft`, `rounded`, `pill` |  |
| `themes[].button.weight` | string | no | one of `normal`, `medium`, `bold` |  |
| `themes[].dark` | object or null | no |  | A light theme's colors for dark mode, kept while it is light. Switching the theme to dark swaps them in. A light theme's light (or a dark theme's dark) is dropped. |
| `themes[].dark.accent` | string | yes |  | The main color: what you can press. |
| `themes[].dark.background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].dark.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].dark.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].dark.surface` | string | no |  | Cards. Left out: the mode's own. |
| `themes[].id` | string | yes |  | Lower-case letters, digits, - and _, up to 32. |
| `themes[].light` | object or null | no |  | A dark theme's colors for light mode, kept while it is dark. Switching the theme to light swaps them in. |
| `themes[].light.accent` | string | yes |  | The main color: what you can press. |
| `themes[].light.background` | string | no |  | The page. Left out: the mode's own. |
| `themes[].light.overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].light.second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].light.surface` | string | no |  | Cards. Left out: the mode's own. |
| `themes[].mode` | string | yes | one of `light`, `dark` |  |
| `themes[].name` | string | yes |  | 1 to 60 characters. |
| `themes[].overrides` | map of string | no |  | Worked-out colors replaced by name: text, accent, accentFg, accentText, second, secondFg, secondText, link, navLink, background, surface, nested, chrome, bar (the bar across the top), barText, info, infoBorder, border, rowOdd and rowEven (alternating table rows), and for each button variant V (Secondary, Success, Danger, Warning, Info, Light, Dark) btnV, btnVFg and btnVText. Leaving one out puts the readable color back. |
| `themes[].second` | string | yes |  | Says where you are: active nav item, headings, info. |
| `themes[].surface` | string | no |  | Cards. Left out: the mode's own. |

## tag_project

Your own labels for a project you are on, replacing the ones it had.

|  |  |
|---|---|
| Route | [`POST /api/projects/{project_id}/tags`](api.md#post-apiprojectsproject_idtags) |
| Read only | no |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|
| `project_id` | integer | yes | at least 1 | The project's id, from GET /api/projects (list_projects over MCP). |
| `tags` | array of string | yes |  | Up to 10, each 1-30 characters. They are lowercased. An empty list clears them. |

## whoami

The account this API key belongs to, the role it acts with (role -- never higher than the key's own scope), and the account's own role.

|  |  |
|---|---|
| Route | [`GET /api/me`](api.md#get-apime) |
| Read only | yes |
| Destructive | no |

**Arguments**

| Field | Type | Required | Limits | Description |
|---|---|---|---|---|

## HTTP only

Routes an API key may call that have no tool, and why.

| Route | Why |
|---|---|
| [`GET /api/admin/users/{user_id}/keys`](api.md#get-apiadminusersuser_idkeys) | Manages API keys, which no MCP tool does. |
| [`GET /api/api.json`](api.md#get-apiapijson) | The API reference itself; GET /api/mcp.json is the MCP one. |
| [`GET /api/api.md`](api.md#get-apiapimd) | The API reference itself, for people; GET /api/api.json is the same as JSON. |
| [`GET /api/mcp.json`](api.md#get-apimcpjson) | The MCP reference itself; tools/list gives the same list. |
| [`GET /api/mcp.md`](api.md#get-apimcpmd) | The MCP reference itself, for people; tools/list gives the same list. |
| [`POST /api/admin/keys/{key_id}/revoke`](api.md#post-apiadminkeyskey_idrevoke) | Manages API keys, which no MCP tool does. |
| [`POST /api/admin/users/{user_id}/keys`](api.md#post-apiadminusersuser_idkeys) | Makes an API key, which no MCP tool does. |
| [`POST /api/keys/self/rotate`](api.md#post-apikeysselfrotate) | Makes an API key's credential, which no MCP tool does. |
| [`POST /api/logout`](api.md#post-apilogout) | Ends a browser session; a key has none to end. |
| [`POST /api/logs/events`](api.md#post-apilogsevents) | The browser's own activity reports (track.js), not an action. |
| [`POST /mcp`](api.md#post-mcp) | The MCP endpoint itself. |
