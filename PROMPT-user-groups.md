# Prompt: user groups in ProjectPortal (paste into plan mode in this repo)

I want **user groups** in ProjectPortal, matching a groups feature being added to
the `go-webserver` and `python-webserver` skills (and so to every app in the
catalog). I'm not sure I'm describing it accurately, so treat what follows as the
idea, not the spec. Before you write the plan, run the `interview` skill and ask me
questions until the design is pinned down.

## The idea

A group is another level of separation, above accounts and roles. Each group is
isolated: its people, and what it can see, are invisible to anyone outside it, not
even admins of other groups see it. An admin sees and manages only the groups they
are in.

In the portal that means two things that may or may not be the same:
1. **The portal's own groups**: people, projects, memberships, invite links, audit
   log split by group, so a group admin runs their own corner of the portal.
2. **Groups inside the apps**: when the portal adds someone to TodoBoard (Join),
   it can put them in a particular TodoBoard group, create that group, and show
   which group each membership is in. A portal group might map to one group on
   each app, or not.

## What to read first

- `CROSS_PROJECT.md`, especially "The contract the portal relies on" and
  "Portal contract v2": capabilities in `/api/me` turn features on per app.
- `server/remote/remote.go` (calls to apps), `server/models/project.go`,
  `membership.go`, `user.go`, `owner.go`, `invite_link.go`,
  `server/routes/admin_projects.go`, `admin_manage.go`, `join.go`, `team.go`,
  `projects.go`, and the frontend screens for the catalog, Manage project and Team.
- `~/.claude/skills/PLAN-user-groups.md` once it exists: the webserver plan owns
  the app side and the contract v3 wire format. Copy its field names exactly. If
  it doesn't exist yet, say so and plan against the open questions instead.
- ProjectPortal was built from go-webserver, so its own groups may come from the
  skill's template through `webserver-upgrade` rather than being written here.
  Find out which parts are template and which are portal-only.

## Things the interview has to settle (not a complete list)

- Does the portal itself get groups, or only learn to drive groups on the apps?
  Or both, and is a portal group tied to app groups?
- The portal owner: sees every group, or can be kept out?
- Can a person be in several portal groups, with a role per group, and how do they
  switch?
- What's scoped per group: people, projects in the catalog, memberships, invite
  links, audit log, jobs/health checks, settings. Can a project (an app) be shared
  by several portal groups, each seeing only its own group on that app?
- The project's admin key: one per project, or one per app group? A key scoped to
  one app group can't see others, which may be exactly what we want.
- Join, Open, reissue login, Manage project (roles, expiry, keys, invite links,
  disable/delete): what each does with a group, and how it looks on screen.
- Apps without the `groups` capability: everything keeps working as today.
- Migration: existing people, projects and memberships go into a default group.
- Isolation tests: nothing from portal group A is reachable from B by any route,
  key, MCP tool or screen.

## What the plan should produce

- The data model and rules, every route, screen, MCP tool and `language.yaml`
  string that changes, Bootstrap 5 and mobile first.
- What comes from the webserver template versus what's written here.
- The `remote.go` changes for contract v3, gated on the capability.
- Tests, and a manual check against TodoBoard once it has groups.
- Entries for `CROSS_PROJECT.md`.
- An order of work, and what has to land in the skills first.

Rules: all on `master`, no branches or worktrees, commits and pushes left to
`gitPushUpdate.sh`. Write the final plan to `PLAN-user-groups.md` in this repo.
