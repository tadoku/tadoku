# Contests and leaderboards

## Find it

Main navigation **Contests** has Official contests, User contests and My contests
tabs (`/contests/official`, `/contests/user-contests`, `/contests/my-contests`).
Create contest opens `/contests/new`. A contest links to its `/contests/<id>/leaderboard`,
`/updates`, `/scoring`, `/registration` and participant `/profile/<user-id>` pages.
Main navigation **Leaderboard** opens `/leaderboard/latest`; also cover
`/leaderboard/yearly/<year>` and `/leaderboard/all-time` when aggregation changes.

## Verify

- Read-only starting point: open the seeded Dev Tadoku Round leaderboard and
  confirm real participant rows such as Dev Reader. Follow a participant and
  scoring/update tabs. Test a pagination transition if the selected data has
  multiple pages; don't manufacture a pagination pass from a single page.
- For participant activity charts, check a long contest at phone, tablet and
  desktop widths. The page and chart card stay within the viewport, score cards
  remain readable, and desktop widths (1024px and above) show the full date range
  and both score axes without horizontal scrolling. On narrower screens,
  horizontal scrolling inside the chart reaches both the first and last contest
  dates. Resize across the desktop layout breakpoint in both directions and
  confirm the score column retains its one-fifth width on desktop.
- For creation/registration changes, use a branch API and uniquely named contest.
  Verify allowed languages/activities, dates and privacy; exercise invalid input,
  save, reload and visibility from another identity. Register an eligible user
  and verify the registration state and My contests list. Cover closed/ineligible
  registration only with an owned suitable fixture, not by rewriting shared dates.
- For scoring/leaderboard changes, create a qualifying log through the
  [activity journey](activity.md), submit it to the contest, and verify the
  participant/contest score and profile agree. Check a second read after reload.
  Contrast a separate base context to prove the branch write did not alter base.
- Distinguish no contests/participants from failed requests. Verify anonymous,
  participant and owner/admin boundaries relevant to the change, including the
  private contest case when privacy is in scope.

## Traps

“Latest” depends on current contest dates and is not a stable fixture ID. Use the
explicit seeded contest ID from the index to diagnose missing data. Time windows,
language restrictions, activity units, contest eligibility and registrations affect
scores; don't assume every log counts everywhere. A log may need explicit contest
submission. Cached scores depend on the outbox worker, not just database writes.
Avoid comparing exact shared seed totals after another test may have edited them.

## Source anchors

[Contest routes](../../../../../frontend/apps/webv2/pages/contests/),
[leaderboard routes](../../../../../frontend/apps/webv2/pages/leaderboard/),
[contest/log components](../../../../../frontend/apps/webv2/app/immersion/),
[Tadoku API features](../../../../../services/tadoku-api/features/),
[HTTP journeys](../../../../../services/tadoku-api/e2e/user_journeys_test.go).
