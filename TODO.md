# TODO: portal contract v2 upgrades still not done

Apps on 7291HPEliteDesk that do not yet speak portal contract v2 (see
`CROSS_PROJECT.md`, "Portal contract v2"). Each is upgraded with the
`webserver-upgrade` skill, one app at a time. Nothing is committed by the
skill: review, back up the app's database, then run `gitPushUpdate.sh`.

## Upgraded but not yet committed or deployed (as of 2026-11-22)

- [ ] **MRIExperimentalDesignPlanner** (`~/WORKSPACE/PROJECTS/fMRI/MRIExperimentalDesignPlanner`):
  check the Appearance screen (OSU, WSU) in a browser, back up the database, commit.
- [ ] **AIS Counter** (on 7291-M1, `~/PROJECTS/AIS_Counter`): back up the database,
  commit and push from the M1; the EliteDesk deploys from GitHub.
- [ ] **TodoBoard** (`~/WORKSPACE/PROJECTS/TodoBoard`): check it in a browser
  (admin row menu, invite links, a viewer account, themes), back up the database, commit.

## Not started

- [ ] **iPB / IOSWebAppPushNotifications** (`~/WORKSPACE/PROJECTS/IOSWebAppPushNotifications`)
  - Left at contract v1 on purpose. It is 32 skill commits behind (b286bc2 -> ea9f39c).
  - Its API key code exists but is switched off in the live config
    (`auth.api_keys: false`), so the portal cannot reach it with a key.
    Switching it on is a config change you make yourself.
  - To get key minting, add `api_keys` to `features:` in `.skill-origin.yaml`.
  - Taking every upgrade means about 75 merge conflicts. Taking only contract v2
    plus `keys-manage-accounts` also means cutting 30 upgrades out of 53 files by hand.
  - Decide which before starting.
- [ ] **VideoToSlides**: unstamped. It needs `baseline`, then your OK on the
  starting commit and feature list, then `stamp`, then an update.
- [ ] **QRCodeOrganizer**: `prepare` refuses while its uncommitted AI-feature work
  (22 files) is in the working tree. Commit that first, then upgrade.

## After deploying

- [ ] ZoteroWebDAV, VocabTrainer, Rook and PaperReference have no invite links,
  so the Manage project tab will still show "Can't do from here: invite links"
  for them. That is expected.
- [ ] Check each upgraded app on the portal's Manage project tab.
