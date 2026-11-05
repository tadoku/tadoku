---
title: Webv3 profile statistics
description: The Webv3 profile statistics page, its year selection, per-part loading, owner and visitor differences, calendar and its browser journey.
sidebar_position: 3
---

# Webv3 profile statistics

Read this when you change a Webv3 profile page or its data.

`/profile/$id/statistics/{-$year}` shows a learner's year as Webv2's statistics
page does: overall score with the per-day average over UTC calendar days, updates
and active days, score by language and by activity, the daily activity calendar
and contest history. Without a year segment it shows the current UTC year. The
year control steps from the profile's first year to the current one; an unknown
or malformed ID or year is not found. The parts are read together but fail on their own, so a failed part
reports itself while the rest of the year shows. A year without updates shows an
empty state, and contest history remains visible.

`frontend/apps/webv3/src/lib/profile.ts` owns the server function, its result
variants and query key; `frontend/apps/webv3/src/lib/profile.server.ts` reads the
profile, yearly activity, scores, activity split and contest registrations
through the generated client. Contest history is read with the visitor's session,
because the API includes private contests only for their participant and
administrators; a rejected cookie falls back to the public history. More than 50
contests page through `?contests=<page>`.

Paper's `HeatmapChart` colours days by fixed points-per-day bands named in its
legend, and links each day with activity to Webv2's updates for that UTC day. On
narrow screens the calendar scrolls within its labelled region. Only the
profile's own account sees **Download CSV** and **Download JSON**, which download
its logs from the export operation through Webv2's API path.
`frontend/apps/webv3/src/components/ProfileHeader.tsx` renders the name and the
Overview and Activity views; Activity opens Webv2's updates.

`profile.spec.ts` creates a log through the form, compares each total, breakdown
and contest with the API for two years, switches years with the year control,
and checks owner tools and private contests, the empty year, not-found addresses
and a 390px layout. It reuses the shared Dev Reader session. Without fixture
credentials only the visitor and not-found journeys run: a visitor opens a
homepage standings profile without JavaScript and compares it with the API. The
container smoke's synthetic API serves that learner's profile year.
