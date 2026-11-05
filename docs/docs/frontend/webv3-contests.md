---
title: Webv3 contest lists
description: Webv3's official, community and My contests lists, their paths, page numbers, viewer-dependent contents, statuses and create action.
sidebar_position: 3
---

# Webv3 contest lists

Read this when you change a Webv3 contest list.

`/contests/official`, `/contests/user-contests` and `/contests/my-contests` keep
Webv2's paths, including an optional trailing page number such as
`/contests/official/2`. A missing or unreadable page is page 1, as on the
[leaderboards](webv3-leaderboards.md). Pages hold 25 contests in the API's
order, newest first.

- `frontend/apps/webv3/src/lib/contests.ts` owns the server functions, query
  options and route loader: which list a collection shows the viewer, one page
  of it, and whether the viewer may create a contest.
  `frontend/apps/webv3/src/lib/contests.server.ts` reads them through the
  generated client with the visitor's session cookies.
- `frontend/apps/webv3/src/components/ContestsPage.tsx` composes the collection
  tabs, rows, pagination and the empty, failed, past-the-end and signed-out
  states.

The API decides which unlisted contests a list includes: a signed-in account's
own, and every one for administrators. Lists are therefore read as the visitor
and cached per viewer; a rejected session cookie falls back to the public list.
My contests lists the community contests the signed-in account organizes and
asks anonymous visitors to log in. As in Webv2, community and My contests lists
name each contest's languages and activities for signed-in visitors only, and
the create action, which opens Webv3's [contest form](webv3-contest-admin.md),
appears when the API's create-permission check answers 200.

Each row shows the round's status (upcoming, live or ended) and, until it ends,
whether registration is open, both decided by the UTC day the list was read.
Contest dates are UTC calendar days and the page says so. Rows link to the
native [contest page](webv3-contest.md). Search, date filters and a list
of joined contests are not available.

`frontend/apps/webv3/e2e/contests.spec.ts` compares the server-rendered lists
with the API, follows tabs and page links, checks the signed-out prompt and
containment at 390px and 1280px. With the Dev Admin credentials it gives that
account more than one page of community contests in the selected tenant, then
pages through its My contests; with Dev Reader it checks that the seeded
unlisted contest appears for its owner only.
