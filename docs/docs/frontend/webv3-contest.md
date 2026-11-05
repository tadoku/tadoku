---
title: Webv3 contest page and registration
description: Webv3's contest page with its standings and viewer actions, the registration form, their deadlines, permissions, failure states and checks.
sidebar_position: 3
---

# Webv3 contest page and registration

Read this when you change a Webv3 contest page or contest registration.

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
  latest official leaderboard: notices, standings, conditions, summary and recent
  logs. `frontend/apps/webv3/src/components/ContestRegistration.tsx` holds the
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
contest ends for participants, Log activity while it runs for participants, and
Scoring rules for administrators. Anonymous visitors see a login prompt; joining
sends them to log in and back to the registration. Suspended accounts get no join
action. Scoring rules opens the native [scoring page](webv3-contest-admin.md);
participant rows and contest updates open Webv2, and the organizer opens their
Webv3 profile.

The registration form accepts one to three distinct languages: the contest's
allowed languages, or every configured language when it has no restriction.
Removing a registered language warns that the contest logs in that language
will be removed. A save is refused before the API call when the session ended or
another account signed in since the form opened. After the deadline the page
explains that registration has closed instead of offering the form; an API
refusal after the deadline shows Webv2's message and keeps the form. A saved
registration returns to the contest page.

## Checks

`frontend/apps/webv3/e2e/contest.spec.ts` compares the server-rendered page with
the API, follows page segments, checks unknown contests, the anonymous login
redirect and containment at 390px and 1280px. With the Dev Reader and Dev Admin
credentials it creates community contests in the selected tenant, registers Dev
Reader, changes its languages and checks a refused registration after the
deadline.
