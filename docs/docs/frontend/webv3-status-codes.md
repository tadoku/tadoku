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
| A server function call whose payload does not decode | 400 | A plain-text `Bad Request` |
| A call to a server function this build doesn't have, such as from a tab loaded before a release | 409 | Plain text, with `X-Tadoku-Outdated-Build: true`; the tab reloads once |
| An unexpected server error | 500 | The error page |

A loader fails its page with `pageUnavailable` from
`frontend/apps/webv3/src/lib/page-status.ts`, passing the main read's query. It
drops that cached failure, so Try again reads it afresh, and
`frontend/apps/webv3/src/server.ts` answers the page with 503 rather than the
router's 500. Use `notFound()` for missing data; never fail a page for a
section it can render without.

`frontend/apps/webv3/e2e/outages.spec.ts` checks a failed, a malformed and a
slow profile section, an unreadable schedule beside a working home board, and a
contest page whose main read fails or answers malformed, then recovery, through
a local server whose API reads it can fail.

`outdatedServerFunctions` in `frontend/apps/webv3/src/start.ts` answers an
unknown server function ID with 409, so open tabs calling renamed or removed
functions after a release never count toward the burst alert.
`serverFunctionFetch` in `frontend/apps/webv3/src/lib/outdated-build.ts` then
reloads that tab to load the current build. A tab that reloaded for this in the
last five minutes, or cannot use `sessionStorage`, shows the call's failure
instead, so a lasting mismatch cannot reload in a loop.
`frontend/apps/webv3/e2e/stale-build.spec.ts` checks the 409, the single reload
and the loop guard.

`undecodableServerFunctions` in `frontend/apps/webv3/src/start.ts` answers a
server function call with 400 when TanStack Start fails it before any function
middleware runs, which happens only when its payload does not decode; the
origin check in `mutations` never sees such a call. A call that decoded keeps
its status: one its input validator rejects answers 200 with the serialized
error the client throws. A route answers 404 for any path parameter its server
function's input schema would reject, as the profile statistics route does by
parsing with that schema; on the server, the rejection would fail the page with
500.
`frontend/apps/webv3/e2e/server-function-input.spec.ts` checks malformed GET
and POST payloads.

Each 5xx response writes one `ssr_error` [server log](webv3-server-logs.md)
line. The edge counts every 5xx toward the Webv3 5xx burst alert, which opens
an incident for the owner, so an API outage that fails resource pages can trip it.
