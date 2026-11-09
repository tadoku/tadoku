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
  the language options and `useSaveRegistration()`, which refreshes the
  registration, the contest's board and standings and the activity form's
  contests after a save. `frontend/apps/webv3/src/lib/registration.server.ts`
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

`/contests/$id/profile/$userId` shows one participant in one contest: their
overall score and the score for each registered language, each active UTC day's
score per language with the running total, and their contest logs, paged the
same way. Webv2 draws the daily scores as a chart; Paper has no bar chart, so
Webv3 lists them in a table. The API answers 404 for an unknown contest or a
learner who has not joined it, and the page is not found. The daily scores fail
on their own while the scores and logs still show. View profile opens the
learner's Webv3 profile.

The contest's organizer and administrators see Remove from contest on each log
in both lists. It asks for a reason of up to 1,000 characters before it sends
`ImmersionContestModerationDetachLog`, which enforces the same permission. A
removal is refused before the API call when the session ended or another account
signed in since the page opened. The log stays in the learner's personal record;
the contest's logs, board, standings and participant profiles are read again.

- `frontend/apps/webv3/src/lib/contest-activity.ts` owns the contest log and
  participant queries and `useDetachLog()`;
  `frontend/apps/webv3/src/lib/contest-activity.server.ts` reads them anonymously
  and removes a log with the visitor's session.
- `frontend/apps/webv3/src/components/ContestUpdates.tsx` renders a contest's
  logs and the moderation dialog, `frontend/apps/webv3/src/components/ContestProfile.tsx`
  the participant profile, and `frontend/apps/webv3/src/components/LogEntries.tsx`
  the log list it shares with [profile activity](webv3-profile.md#activity).

## Checks

`frontend/apps/webv3/e2e/rank.spec.ts` checks that a participant below the first
standings page sees their rank on the home page and their participant page.

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
