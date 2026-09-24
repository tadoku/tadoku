# Activity logging and learner profiles

## Find it

After login, **Log activity** opens main-host `/logs/new`. A profile's Updates
list links to `/logs/<id>`; owner actions lead to `/logs/<id>/edit` and
`/logs/<id>/contests`. The user menu's Profile opens
`/profile/<user-id>/statistics/<year>`; its Updates tab opens
`/profile/<user-id>/updates`. A leaderboard participant is another route to a profile.

## Verify

Use a selected branch API and a synthetic identity. Record the starting profile
and leaderboard values; don't edit seed logs or another run's records.

1. Create an owned log with a unique description, language and activity. Exercise
   the applicable amount/unit or time fields, tags and contest selection. Check
   invalid/missing input before a successful submission.
2. Follow the result to log details. Confirm amount/time, language, score,
   description and tags; reload to prove persistence. Check the same log in Profile
   → Updates and the applicable Statistics year. Inspect real API errors, not only
   toast notifications.
3. If editing is affected, edit that log and verify the detail and profile again.
   For contest submission, inspect `/logs/<id>/contests` and the chosen contest's
   updates/ranking. Wait for the observable worker result, not a fixed sleep.
4. Test owner vs another user vs anonymous access for the changed action. If the
   task covers deletion, delete only this run's log through the supported UI and
   verify its removal from the relevant lists/score. Otherwise record retained IDs.

## Traps

`release-log-entry-v2` selects between two forms. Its default is false, and the
decision is latched while the form is mounted. Record the active variant; don't
claim both were covered by one visit. For variant changes, use an authorized test
identity's feature access and remount the route. Shared flag-policy changes need
separate scope. A disabled or unavailable variant is an explicit coverage gap.
The newer flow separates logging from contest submission; creating a log alone
does not prove it reached the expected contest leaderboard.

Leaderboard updates are asynchronous and branch-cache scoped. A healthy API pod
does not prove the worker processed the write. Keep selected headers and actual
before/after data as evidence. Years, dates and eligible contests affect where a
log appears. An empty list and an API failure are different states.

## Source anchors

[Log routes](../../../../../frontend/apps/webv2/pages/logs/),
[forms and API hooks](../../../../../frontend/apps/webv2/app/immersion/),
[profile routes](../../../../../frontend/apps/webv2/pages/profile/),
[flag registry](../../../../../frontend/apps/webv2/app/feature-flags/registry.ts),
[backend user journeys](../../../../../services/tadoku-api/e2e/user_journeys_test.go).
