# Activity logging and learner profiles

## Find it

After login, **Log activity** opens main-host `/logs/new`. A profile's Updates
list links to `/logs/<id>`; owner and admin actions lead to `/logs/<id>/edit` and
`/logs/<id>/contests`. The user menu's Profile opens
`/profile/<user-id>/statistics/<year>`; its Updates tab opens
`/profile/<user-id>/updates`. On Statistics, clicking a non-empty heatmap day
opens `/profile/<user-id>/updates/1?date=<YYYY-MM-DD>`, which lists only that
UTC day's updates. A leaderboard participant is another route to a profile.
The heatmap colours days by fixed points-per-day bins shown in its legend.
On your own profile, the header's **Export** menu offers **Download CSV** and
**Download JSON** on both Updates and Statistics. Other profiles and anonymous
visitors do not show this menu.

## Verify

Use a selected branch tenant and a synthetic identity. Frontend-only work can
use the base API; add an API overlay when changed API code is part of the proof.
Check the actual API response's `X-Dev-Selected` before writing, and verify the
persisted tenant when proving isolation. Record the starting profile and
leaderboard values; don't edit seed logs or another run's records.

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

5. On your own profile, download both formats through **Export**. Compare the
   exported log count with the unfiltered Updates total across all pages. Check
   the first and last records, amount-only and time-only entries, tags and text.
   Repeat from Statistics; its year selection does not restrict the export.
   Visit another profile and sign out to confirm the menu is absent. Direct
   export requests by guests must return 401; administrators export only their
   own records.

## Traps

All accounts use the v2 log form and details; editing requires ownership or an
administrator role. Logging and contest submission are separate steps, so
creating a log alone does not prove it reached the expected contest leaderboard.
Check both amount-based legacy logs and duration-based logs when editing changes.
Language-specific units must still override the generic unit with the same name.
The retained feature-flags API must return only `release-log-entry-v2: true`,
including for guests, users without a named grant and unavailable Flipt.
Check the complete response: already-open clients reject a replacement key or
an additional key even when the HTTP status is 200.

Leaderboard updates are asynchronous and branch-cache scoped. A healthy API pod
does not prove the worker processed the write. Keep selected headers and actual
before/after data as evidence. Years, dates and eligible contests affect where a
log appears. An empty list and an API failure are different states.

Heatmap days and the `?date=` filter use UTC days, so a filtered Updates list
can show rows with the neighbouring local date.

The Statistics overall score card's per-day average divides the year's total by
UTC calendar days, including idle days: every day of a past year, or Jan 1 through
today for the current year. It is hidden for future years and zero totals.

Exports use native attachment downloads; inspect the downloaded file rather
than expecting a rendered page. Deleted logs and contest attachments are
excluded. CSV starts with a UTF-8 BOM and prefixes formula-like text cells with
an apostrophe. Tags are sorted and joined with comma and space, so tags containing
commas require JSON for a lossless representation. JSON preserves text and uses
null for absent amount, unit, duration and description values.

## Source anchors

[Log routes](../../../../../frontend/apps/webv2/pages/logs/),
[forms and API hooks](../../../../../frontend/apps/webv2/app/immersion/),
[profile routes](../../../../../frontend/apps/webv2/pages/profile/),
[export menu](../../../../../frontend/apps/webv2/app/immersion/LogExportMenu.tsx),
[export transport](../../../../../services/tadoku-api/transport/http/log_export.go),
[export HTTP cases](../../../../../services/tadoku-api/e2e/log_export_test.go),
[backend log journeys](../../../../../services/tadoku-api/e2e/log_journeys_test.go).
