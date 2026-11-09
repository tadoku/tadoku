---
title: Webv3 profile
description: The Webv3 profile statistics and activity pages, year selection, per-part loading, owner and visitor differences, calendar, log list, pagination and their browser journeys.
sidebar_position: 3
---

# Webv3 profile

Read this when you change a Webv3 profile page or its data.

`/profile/$id/statistics/{-$year}` shows a learner's year as Webv2's statistics
page does: overall score with the per-day average over UTC calendar days, updates
and active days, score by language and by activity, the daily activity calendar
and contest history. Without a year segment it shows the current UTC year. The
year tabs run from the current year back to the profile's first year; an unknown
or malformed ID or year is not found. The parts are read together but fail on their own, so a failed part
reports itself while the rest of the year shows. A year without updates shows an
empty state, and contest history remains visible.

`frontend/apps/webv3/src/lib/profile.ts` owns the server function, its result
variants and query key; `frontend/apps/webv3/src/lib/profile.server.ts` reads the
profile, yearly activity, scores, activity split and contest registrations
through the generated client. Contest history is read with the visitor's session,
because the API includes private contests only for their participant and
administrators; a rejected cookie falls back to the public history. Each contest
lists its dates and the languages the learner registered. More than 50 contests
page through `?contests=<page>`.

Paper's `HeatmapChart` colours days by fixed points-per-day bands named in its
legend, and links each day with logs, including a day whose logs scored nothing, to the profile's activity for that UTC day. On
narrow screens the calendar scrolls within its labelled region, keeps its
weekday labels pinned and opens the current year with today in view. A date field below the calendar, Open a day, opens any day's
activity too, so a day never depends on hitting its small cell; before
hydration it submits as a plain GET form to the same address. Only the profile's own account sees the CSV
and JSON export links, which download its logs from the export operation
through Webv2's API path.
`frontend/apps/webv3/src/components/ProfileHeader.tsx` renders the learner's
[avatar](webv3.md#avatars), hidden on phones, their name and the Overview and
Activity views.

`profile.spec.ts` creates a log through the form, compares each total, breakdown
and contest with the API for two years, switches years with the year tabs,
and checks owner tools and private contests, the empty year, not-found addresses
and a 390px layout. It reuses the shared Dev Reader session. Without fixture
credentials only the visitor and not-found journeys run: a visitor opens a
homepage standings profile without JavaScript and compares it with the API. The
container smoke's synthetic API serves that learner's profile year.

## Activity

`/profile/$id/updates/{-$page}` lists a profile's logs as Webv2's updates page
does: newest first, 50 per page with page links above and below the list,
without deleted logs, each day's date shown once above its logs, each linking to its
[native detail page](webv3-logging.md). A missing page segment
reads as the first page, and a malformed one too, as in Webv2. `?date=YYYY-MM-DD`
lists one UTC day's logs, with a way back to all of them; a malformed date lists
every log. Entries show their UTC creation day, description, language, activity,
amount in its unit and duration, and score. A description tagged NSFW is
[masked as on its detail page](webv3-logging.md), and the entry is titled by its
language and activity. A page past the end links back to the first page. A profile without
logs shows an empty state that offers its owner the activity form; only the owner
sees the downloads.

`frontend/apps/webv3/src/lib/profile.ts` owns the server function, its result and
query key; `readProfileActivity` in `frontend/apps/webv3/src/lib/profile.server.ts`
reads the profile and the page of logs together. The list fails on its own while
the profile header still shows. Search, filters and their applied-filter
summary in the Paper design need API support and are not built.
`frontend/apps/webv3/src/components/ProfileActivity.tsx` renders the page around
the log list in `frontend/apps/webv3/src/components/LogEntries.tsx`.

`profile-activity.spec.ts` follows a homepage standings profile to its activity,
compares each page with the API, pages forward and opens an entry. With fixture
credentials it tops Dev Reader up to more than one page through the API's create
operation, checks the owner's pages, a page past the end, a calendar day and its
way back, and registers a fresh account in the branch tenant to check the owner
and visitor empty states. Registration has its own budget of ten per tenant and
client IP in 15 minutes, separate from logins. It also checks not-found addresses
and a 390px layout. The container smoke's synthetic API serves 51 logs for its
standings learner, including one tagged NSFW, and their detail pages.
