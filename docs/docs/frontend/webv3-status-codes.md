---
title: Webv3 status codes and failure pages
description: Which HTTP status each Webv3 page answers when its data is missing or its main read fails, what renders then, and which of those responses count toward the 5xx burst alert.
sidebar_position: 3
---

# Webv3 status codes and failure pages

Read this when a Webv3 page or loader decides what to show for missing or
unavailable data, or when you read Webv3's 5xx responses.

| Situation | Status | Renders |
| --- | --- | --- |
| Unknown path | 404 | The not-found page |
| A path that does not decode, such as `/pages/%FF` or a lone `%` | 400 | A plain-text `Bad Request` from `frontend/apps/webv3/server/plugins/malformed-path.ts` |
| A page for one resource, such as a contest, profile, log, blog post or CMS page, whose main read finds nothing | 404 | The route's not-found view |
| That page's main read fails | 503 | The error page, with Try again |
| A list, the home page, or a section beside a page's main read fails | 200 | That section's own failure state |
| A logout the API does not confirm | 503 | Log out is unavailable, with Try again, in the app shell |
| An unexpected server error | 500 | The error page |

A loader fails its page with `pageUnavailable` from
`frontend/apps/webv3/src/lib/page-status.ts`, passing the main read's query. It
drops that cached failure, so Try again reads it afresh, and
`frontend/apps/webv3/src/server.ts` answers the page with 503 rather than the
router's 500. Use `notFound()` for missing data; never fail a page for a
section it can render without.

Each 5xx response writes one `ssr_error` [server log](webv3-server-logs.md)
line. The edge counts every 5xx toward the Webv3 5xx burst alert, which reverts
a release, so an API outage that fails resource pages can trip it.
