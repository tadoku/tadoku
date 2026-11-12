---
title: Webv3 contest page and registration
description: Webv3's contest page with its standings, activity and viewer actions, participant contest profiles, log moderation, the registration form, their deadlines, permissions, failure states and checks.
sidebar_position: 3
---

# Webv3 contest page and registration

Read this when you change a Webv3 contest page, its activity or participant
profiles, log moderation or contest registration.

`/contests/$id/leaderboard` is a contest's main page, with an optional trailing
page number such as `/contests/<id>/leaderboard/2`, read like the
[leaderboards](webv3-leaderboards.md)' page segments. `/contests/$id/registration`
joins the contest or changes a registration's languages. Both keep Webv2's paths.
An unknown or malformed contest ID answers 404 with a not-found page.

- `frontend/apps/webv3/src/lib/leaderboard.ts` owns the contest board query, read
  as the visitor because administrators can also open deleted contests, and the
  contest's phase on the UTC day the board was read.
- `frontend/apps/webv3/src/lib/registration.ts` owns the viewer's registration,
  the language options and `useSaveRegistration()`, which refreshes every read
  a save changes, by query-key prefix, including the Joined list and the
  account's contest history. `frontend/apps/webv3/src/lib/registration.server.ts`
  calls the API with the visitor's session cookies.
- `frontend/apps/webv3/src/components/ContestPage.tsx` composes the page header
  and actions around the board that
  `frontend/apps/webv3/src/components/ContestLeaderboard.tsx` shares with the
  latest official leaderboard: notices, standings and conditions; the Activity
  tab lists the contest's logs.
  The selected Leaderboard tab names the standings, so the contest page shows no
  separate standings heading. `frontend/apps/webv3/src/components/ContestRegistration.tsx` holds the
  registration states and form.

## Rules

The round's status (upcoming, live or ended) and registration are separate: a
live contest can be closed to new participants. Contest dates are UTC calendar
days. A contest accepts new participants through its registration end day;
registered participants can change their languages until the contest's last day,
as the API enforces. An unlisted contest is left out of contest lists but anyone
with its link can view and join it.

The page shows only the actions Webv2 allows the viewer: Join contest while
registration is open and the viewer hasn't joined, Update registration until the
contest ends for participants, and Scoring rules for administrators. Join contest
is the page's only button. The other two are items in a Contest options menu
(an icon-only "⋯" trigger), which is left out when it would be empty. On phones the
trigger sits at the end of the back link's row, so the title keeps the full width.
Logging uses the header's Log activity action. Anonymous visitors see a login prompt; joining
sends them to log in and back to the registration. Suspended accounts get no join
action. Scoring rules opens the native [scoring page](webv3-contest-admin.md),
participant rows open their contest profile and the organizer opens their Webv3
profile.

The registration form accepts one to three distinct languages: the contest's
allowed languages, or every configured language when it has no restriction.
Removing a registered language warns that the contest logs in that language
will be removed. A save is refused before the API call when the session ended or
another account signed in since the form opened. After the deadline the page
explains that registration has closed instead of offering the form; an API
refusal after the deadline shows Webv2's message and keeps the form. A saved
registration returns to the contest page.

## Activity and participant profiles

The Leaderboard and Activity tabs share the page header and actions.
`/contests/$id/updates` lists every participant's contest logs newest first, as
Webv2's updates page does, 25 to a page, without deleted logs, each naming its
learner beside their [avatar](webv3.md#avatars) and opening its
[native detail page](webv3-logging.md). A missing or
malformed page segment reads as the first page; a page past the end links back
to it.

`/contests/$id/profile/$userId` shows one participant in one contest, paged the
same way. The header gives their place and overall score and links View profile
to the learner's Webv3 profile. The Activity section has four figures, score,
active days, logs and best day, and a chart of points per UTC day stacked by
activity in the activity colours, with the running total, from the contest's
first day to today or its last day, whichever is earlier. Below it the Logs list
sits beside Score by language, which appears only with more than one registered
language, Score by activity and Score by tag. Tags overlap, so their rows show
points and logs but no share; the top six are listed and the rest follow on one
line. Below the tablet breakpoint the three lists fold into one disclosure
between the chart and Logs. The API answers 404 for an unknown contest or a
learner who has not joined it, and the page is not found. A failed summary read
leaves the scores and logs on the page.

### Filters

The Filters menu beside the Activity heading filters by language (only with
more than one registered language), activity and tag, following the
[Paper filter pattern](paper-composition.md#filters). Every figure on the page
follows the filters: the summary, the chart, Logs and the three lists, each list
counted under every filter but its own. Values in one filter match any of them,
and filters combine. Filtered, the score figure reads as a share of the
participant's contest score instead of their place, and the applied values
show as removable chips with Clear all and the filtered log count. When nothing
matches, Clear filters replaces the summary.

Filters are repeated search parameters, `?language=jpn&activity=listening&tag=manga`,
holding language codes, lower-case activity names and tags.
`frontend/apps/webv3/src/lib/contest-filters.ts` parses them at the route: it
drops malformed language codes, activity names `frontend/apps/webv3/src/lib/activity.ts`
doesn't know, empty or overlong tags, repeats and values past the API's limits,
and the server redirects to the cleaned address. A language code the
participant didn't register matches nothing. `frontend/apps/webv3/src/lib/search.ts`
writes the router's list parameters as repeated keys, so the page's links and a
submitted form give the same address. Changing a filter returns to the first
page, page links keep the filters, and while the new figures load the page keeps
showing the current ones. Without JavaScript the Filters button is a disclosure
holding a GET form with one checkbox group per filter and Apply; hydration
replaces it with the menu.

The contest's organizer and administrators see Remove from contest on each log
in both lists. It asks for a reason of up to 1,000 characters before it sends
`ImmersionContestModerationDetachLog`, which enforces the same permission. A
removal is refused before the API call when the session ended or another account
signed in since the page opened. The log stays in the learner's personal record;
the contest's logs, participant profiles and places, the standings and the
log's submissions are read again.

- `frontend/apps/webv3/src/lib/contest-activity.ts` owns the contest log,
  participant and summary queries and `useDetachLog()`;
  `frontend/apps/webv3/src/lib/contest-activity.server.ts` reads them anonymously
  and removes a log with the visitor's session.
- `frontend/apps/webv3/src/components/ContestUpdates.tsx` renders a contest's
  logs and the moderation dialog, `frontend/apps/webv3/src/components/ContestProfile.tsx`
  the participant profile with its filters from
  `frontend/apps/webv3/src/components/ContestProfileFilters.tsx`, and
  `frontend/apps/webv3/src/components/LogEntries.tsx` the log list it shares with
  [profile activity](webv3-profile.md#activity).
  `frontend/apps/webv3/src/components/Breakdown.tsx` draws the score lists both
  profiles share.

## Checks

`frontend/apps/webv3/e2e/rank.spec.ts` checks that a participant below the first
standings page sees their rank on the home page and their participant page.

`frontend/apps/webv3/e2e/viewer-change.spec.ts` signs in from another tab and
switches accounts, then checks that the contest, registration and log pages show
their loading state while the new viewer's data loads, never their error state.

`frontend/apps/webv3/e2e/contest.spec.ts` compares the server-rendered page with
the API, follows page segments, checks unknown contests, the anonymous login
redirect and containment at 390px and 1280px. With the Dev Reader and Dev Admin
credentials it creates community contests in the selected tenant, registers Dev
Reader, changes its languages and checks a refused registration after the
deadline.

`frontend/apps/webv3/e2e/contest-activity.spec.ts` compares the server-rendered
activity and a standings participant's contest profile with the API, pages through
them, opens a log, checks not-found addresses and a 390px layout. With both
fixture credentials it submits a Dev Reader log to a contest Dev Admin creates,
checks that Dev Reader sees no moderation action, and removes the log as Dev
Admin with a reason; a contest Dev Reader organizes checks the organizer's
removal. The container smoke's synthetic API serves 51 contest logs and the
standings learner's contest profile.

`frontend/apps/webv3/e2e/contest-profile.spec.ts` seeds a Dev Reader profile with
two languages, both activities and overlapping tags in a contest Dev Admin
creates, and checks the filters against the API's filtered summary and log list:
choosing from the menu, tags matching either and an activity narrowing them,
removing chips and Clear all returning focus to Filters, a filtered address, a
filter change from page 2, invalid values, the form without JavaScript and the
full-screen sheet at 390px.
