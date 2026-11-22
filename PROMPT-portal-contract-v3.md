Paste this into a new Claude session opened in `~/.claude/skills`:

---

Read `PLAN-portal-contract-v3.md` in this repo and carry out **phase 1** in both
the go-webserver and python-webserver skills.

Work on `master` only. Don't commit: leave that to `gitPushUpdate.sh`.

Before you start:
- Read `~/WORKSPACE/PROJECTS/ProjectPortal/CROSS_PROJECT.md`, for the v2
  contract and how it went wrong.
- Read Rook's `server/routes/admin_games.go` (`RequireAccountAdmin`) and
  `server/routes/account_admin_test.go`, for the fix that motivated item 1.

Do the items in order:
- 1, 3, 4, 7 and 8 first: they're fixes.
- Then 2, 5 and 6: they're upgrades.
- Then 9.

Each item needs tests in both templates, and each test must fail before its
change. After each item, run both test suites and regenerate the API docs.

The wire format in the plan is binding. If something in it can't work as
written, stop and ask before changing it.

At the end, report:
- what changed per item;
- the `UPGRADING.md` entries you added;
- anything left for phase 2 or 3.
