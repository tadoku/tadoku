---
title: Webv3 leaderboards
description: Webv3's latest official, yearly and all-time leaderboards, their paths and page numbers, data owners, reusable standings table and scope labels.
sidebar_position: 3
---

# Webv3 leaderboards

Read this when you change a Webv3 leaderboard or reuse its standings table.

`/leaderboard/latest`, `/leaderboard/yearly/$year` and `/leaderboard/all-time`
keep Webv2's paths, including an optional trailing page number such as
`/leaderboard/all-time/2`. Like Webv2, a page or year segment reads its leading
integer, a missing or unreadable page is page 1 and an unreadable year is the
current UTC year. Pages hold 50 participants.

- `frontend/apps/webv3/src/lib/leaderboard.ts` owns the server functions and
  query options: standings for a `StandingsScope` (one contest, one year or all
  time), the latest official contest's board, and whether the signed-in viewer
  has joined a contest. `frontend/apps/webv3/src/lib/leaderboard.server.ts`
  reads standings, contests, summaries and recent logs anonymously through the
  generated client; only the registration read forwards the session cookies.
- `frontend/apps/webv3/src/components/Standings.tsx` renders one page of any
  scope: tied ranks as `T`, the signed-in viewer's row, the participant count,
  pagination and the empty, failed and past-the-end states. The screen supplies
  the scope caption and the page and profile links.
- `frontend/apps/webv3/src/components/LeaderboardPage.tsx` holds the shared title
  and period tabs and the yearly and all-time boards, which count official
  contests only and list the years since 2020.
  `frontend/apps/webv3/src/components/ContestLeaderboard.tsx` composes the latest
  official contest with its join or logging action, conditions, summary and,
  while it runs, its ten most recent logs. The
  [contest page](webv3-contest.md) shares this board.

Every board names its scope: the contest, the year or all time. The contest's
phase follows the UTC day the board was read, as on the homepage. Yearly and
all-time participants link to their Webv3 [profile statistics](webv3-profile.md);
the latest contest's participants open their native
[contest profile](webv3-contest.md#updates-and-participant-profiles). Contest
details and Join contest open the native [contest page and registration](webv3-contest.md). When no official contest
exists, the latest board says so; the API's 404 is not an error. Name search,
"Find my row" and a separate personal position are not available.
