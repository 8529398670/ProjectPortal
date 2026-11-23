# Security audit: ProjectPortal + TodoBoard (2026-11-22)

**Scope:**
- ProjectPortal (this repo) and TodoBoard at commit 57, run side by side in a throwaway lab: separate state and ports, now stopped.
- Code review of the portal-specific routes and of TodoBoard's auth core.

**Method:**
- Live testing with curl against the lab.
- Two independent code reviews, each confirmed with scratch Go tests that were never written into the repos.

**Bottom line:**
- Plain users and viewers are well boxed in. I found no way for a normal user to reach a project they aren't on, raise their role, or sign in as someone else.
- The weak spot is new: **API keys can turn into browser sessions, or into durable access, in ways the design says they can't.** Most of these bugs live in the go-webserver template (`assets/template/...`), so every Go app has them. The python-webserver template has not been checked yet.

---

## Status (2026-11-22)
- **H1, H2, M1 and M2 are fixed in both webserver templates.** The skills' UPGRADING.md calls this update K1-K5. K3 is an extra fix found while doing H2: a key's *expiry* now also ends the sessions its links opened. It is not committed or rolled out to the apps yet.
- **H3, H4 (portal half) and M3 are fixed in ProjectPortal.** Tests are in `server/routes/portal_keys_test.go`. Nothing is committed or deployed yet.
  - **H3:** Join and Open through an API key need Sign-in links switched on for that key. Otherwise the answer is 403.
  - **H4:** making or promoting an *admin* account on a project needs a browser session; an API key gets 403 `session_only`. A login link for someone else's account on a project is browser-only, so the MCP tool `admin_remote_login_link` is gone. Accounts below admin, and invite links, still work with a key.
  - **M3:** these follow the owner ladder:
    - acting on everyone's project accounts at once;
    - taking someone off a project;
    - cutting off access with `on_project` set to disable or revoke;
    - linking a project account.

    Linking someone to an account that is already another person's gives 409.
- **Still open:**
  - H4's app half: capping service accounts at `user` on the apps. That would break making admins from the portal's Manage project screen, so it is left as a decision.
  - The Low items.

## High

### H1. An admin key can mint itself an unlimited, never-expiring login link, skipping MaySignIn (template; confirmed)
- **Where:** `server/routes/admin.go` `reissueLogin`, and `models.mayActOnTx`, which lets an actor act on themselves.
- **How:**
  - `POST /api/account/login-link` correctly answers 403 for a key without Sign-in links.
  - `POST /api/admin/users/{own_id}/reissue-login {"max_uses":0,"expires_in_minutes":0}` answers 200 with an unlimited link that never expires. The MCP tool `admin_reissue_login` does the same.
- **Impact:**
  - Any admin-role key becomes a full browser session, which opens everything kept session-only: minting keys, setting a password or 2FA, and the owner screens if the key is the owner's.
  - The service account is refused, and 2FA on the account blocks this.
- **Fix:** refuse self-targeting in `reissueLogin`, or send it through `createMyLoginLink`'s rules (MaySignIn, 1 use, 15 minutes).

### H2. Keys made from a key-born session survive revoking the key (template; confirmed)
- **Where:** `server/routes/api_keys.go` `createMyAPIKey` passes an empty `minted_by_key`. `Session.MintedByKey` is never carried onto what the session creates.
- **How:**
  1. Sign in through a link from a key: either via MaySignIn, or via H1.
  2. `POST /api/keys {"role":"admin","expires_in_days":0}`.
  3. Revoke the original key. The session dies, but the new key lives forever. A password set from that session also survives.
- **Impact:** "revoke the leaked key and everything it did goes too" does not hold.
- **Fix:** do one of these:
  - stamp keys created in a session that has `MintedByKey` with that key id, so the revoke cascade catches them;
  - or refuse key, password and 2FA changes from a key-born session.

### H3. A user-scoped portal API key can sign in to every catalog app as its account (ProjectPortal; confirmed by code)
- **Where:** `server/routes/projects.go` `openProject` / `joinProject` (MCP tools `open_project`, `join_project`). Neither route looks at the calling key's MaySignIn.
- **How:**
  - Take any portal key of a person, even one narrowed to role `user`.
  - Calling `open_project` on each project returns login links for that person's account there.
  - Where the person attached their own key, or is linked to an admin or owner account there, that is an admin or owner session on that app.
- **Impact:** narrowing a portal key does not limit it. One leaked low-value key becomes a session on every app the person can use.
- **Fix:** require `MaySignIn` on the calling key for open and join, or make both session-only.

### H4. A portal admin key can leave admin accounts behind on every app (ProjectPortal + apps; confirmed)
- **Where:**
  - `admin_manage.go`: `createRemoteUser` with role admin, `setRemoteRole`, and `remoteLoginLink`. All are reachable by key and over MCP.
  - Upstream, on each app, the service account is an admin, and admins may create admin accounts.
- **What I confirmed live on TodoBoard:**
  - The service account's key creates an `admin` account and gets its login link straight back.
- **What a reviewer confirmed with a test:**
  - Combined with H2, that link gives a backdoor that outlives Disconnect.
- **Context:**
  - The portal already makes `mintRemoteKey` session-only for exactly this reason ("a leaked portal key must not be able to leave keys behind"). These routes defeat that.
- **Fix, in both places:**
  - Portal: make admin-role create, role-to-admin, and remote login links session-only.
  - Apps: let the `kind: service` account create and promote only up to `user`.

## Medium

### M1. Minted keys outlive their parent's expiry (template; confirmed live)
- **What I confirmed live:**
  - I connected the portal with `days: 30`.
  - The portal key then minted a 365-day admin key for its own service account, and a 365-day key for a user.
- **Why it matters:**
  - The revoke cascade covers these keys, but nothing ties them to the parent's *expiry*.
  - Also, the child's account can later set the child to never expire.
- **Fix:** cap the child's `ExpiresAt` at the parent's. Or have `LoadAPIKey` refuse a key whose `MintedByKey` parent is no longer live.

### M2. An admin can clear their own account expiry (template; confirmed)
- **Where:** `models/expiry.go` `SetUserExpiry` refuses only an owner target.
- **How:** owner-set expiry on admin Bob can be removed by Bob, or by Bob's key.
- **Fix:** refuse a self-target for non-owners, as `setUserRole` and `deleteUser` already do.

### M3. A non-owner portal admin can act on the owner's and other admins' project accounts (ProjectPortal; confirmed by code)
- **Where:**
  - `actOnProjectAccounts` (disable or revoke everywhere), with `user_id` set to the owner.
  - `removeMember` and `setAccess` with `on_project`.
  - `linkRemote`, which doesn't check whether the remote account already belongs to someone else. An admin could link themselves to the owner's account on an app, then Open it.
- **Why it matters:** the portal's own owner/admin ladder is not applied to project accounts.
- **Fix:** apply `MaySuperActOn`, and refuse a link to a remote account already linked to another person.

## Low
- **L1. Raw dial errors reach users.**
  - The portal accepts private, loopback and `http://` project addresses.
  - Raw dial errors (`dial tcp 10.0.0.5:22: connection refused`) reach `CheckNote`, and reach plain users in 424 replies.
  - Fix: give users a generic message, and warn on http:// or private hosts.
- **L2. Key can go to a newly edited address in a narrow race.**
  - In `admin_manage.go` `needs()`, `*project = *checked` can pick up a concurrently edited `BaseURL` while still holding the old key.
  - Fix: re-read the key after `needs`, or refuse if the URL changed.
- **L3. `NormalizeBaseURL` builds from the decoded path.**
  - `%23` turns into `#`. This breaks things rather than letting anyone in.
  - Fix: use `EscapedPath()`.
- **L4. Rotation or health-check races can mark a freshly rotated key invalid.**
  - Affects availability only, until the next check.
- **L5. PIN guessing is slow but not impossible.**
  - The lockout is 4 digits, 5 tries per 15 minutes, kept in memory per account: about 10 days on average, and a restart resets it.
- **L6. Promoted invite-joiners can still be claimed.**
  - An invite-joined viewer promoted to user, with no PIN, can still be claimed through the invite link.
- **L7. Editing an invite link re-tags it to the newest key.**
  - Editing an invite link's limits overwrites its `MintedByKey`.
- **L8. Share links survive revoking the key that made them.**
  - A TodoBoard view share link made by a key outlives revoking that key.

## Held up (tested or traced)
- **Bearer handling:**
  - Any `Authorization` header means key-or-nobody, with no cookie fallback.
  - Comparisons are constant-time, and login links use bcrypt.
- **CSRF:** enforced on every non-public write. Login needs same-origin JSON. Cookies are HttpOnly and SameSite=Strict, and a fresh session is made at sign-in.
- **Owner protection on apps (live):**
  - Reissuing the owner's login link: refused.
  - Seeing the owner's sessions: refused.
  - Minting a key for the owner: refused.
  - Minting a key for another admin: owner only.
  - Service key calling `/api/keys`: session-only.
  - Service key making its own sign-in link: refused.
- **Key roles:** `NarrowRole` is applied per request. Minted keys never get MayMintKeys or MaySignIn, and MaySignIn links are forced to 1 use and 15 minutes.
- **Portal user isolation:**
  - Remote ids come from stored memberships only, so there is no IDOR.
  - Join never matches by name.
  - Unlisted projects are admin-only.
  - Grant can't send `admin`.
- **Stored keys:**
  - They never appear in replies, logs or MCP output.
  - Redirects aren't followed, so `Authorization` can't leak to another host.
  - Editing an address drops the keys.
- **Login and invite paths:** must start with `/login/` or `/join/`, and the UI renders them with `Dom.el`, so no `javascript:` link and no open redirect.
- **Viewer and expired accounts:** viewers are blocked globally, MCP included. Expired and disabled accounts are refused in both the session and key loaders.

## Suggested order
1. In the go-webserver template, fix H1, H2, M1 and M2, then roll them out with `webserver-upgrade`. These are the widest. Check python-webserver for the same four.
2. In the portal, fix H3, H4 (the portal half) and M3.
3. On the app side, limit service accounts to `user` (H4).
4. Fix the Low items as convenient.
