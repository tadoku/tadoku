---
title: Webv3
description: The Tadoku product application on TanStack Start, its navigation and shell, request-scoped API and sibling links, dev-cli overlays, container and preview deployment workflow.
sidebar_position: 3
---

# Webv3

Read this when you run or extend `frontend/apps/webv3`.

Webv3 is the complete Tadoku product application. It serves every Webv2 product
page at the same paths, so cutover is a host switch, and its links leave the host
only for the account and admin applications, the log export download and
external sites. Its homepage follows the Paper playground composition and reads the current official contest, summary,
standings and next scheduled round from the configured Tadoku API. Guests see
the introduction and how Tadoku works; a signed-in learner sees their place and
points in the current round, or whether they have joined it, with the year's
points, today's points, logs and active days beside it. Below come the activity
calendar, their five newest logs and this year's contests, unfinished ones
first, ordered within the first page of the year's registrations. The next
round is its own read with its own failure state, and their place comes from
the leaderboard entry at any rank. The home route's loader prefetches all of
this, so it renders on the server. [Native activity logging](webv3-logging.md), including a saved log's contest submissions, uses the existing account session and API.
The latest official, yearly and all-time [leaderboards](webv3-leaderboards.md)
read their standings from the API, and the Official, Community, Joined and Organizing
[contest lists](webv3-contests.md) list contests as the visitor sees them.
Each [contest page](webv3-contest.md) shows its standings, its activity and the
viewer's actions and opens each participant's contest profile. Signed-in learners
join contests and change their languages there, and organizers and administrators
remove logs from a contest. Permitted accounts [create contests](webv3-contest-admin.md),
and administrators manage a contest's scoring rules.
[Profile statistics and activity](webv3-profile.md) read a learner's yearly records
and logs from the API.
The manual, CMS pages and blog are [content pages](webv3-content.md) on the same host.

## Stack and ownership

The application uses TanStack Start, Router, Query and Form, Zod, Vite and Nitro's
Node runtime. React remains on the workspace's 18.2 version so the Paper and
legacy applications keep their existing compatibility.

- `frontend/apps/webv3/src/router.tsx` creates a QueryClient per router/request and
  integrates server rendering and hydration.
- `frontend/apps/webv3/src/lib/home.ts` owns response schemas and the homepage
  query. `frontend/apps/webv3/src/lib/home.server.ts` makes public GET requests to
  the configured API; it accepts no browser-supplied upstream URL or credentials.
- `frontend/apps/webv3/src/lib/links.ts` links every product page on the same
  host. Only the account and admin applications, the log export download and
  external sites leave it.
- `frontend/apps/webv3/src/components/` owns the Paper screen composition,
  navigation and appearance interactions.

Numbers use one locale, `en-US`, through `frontend/apps/webv3/src/lib/format.ts`,
with English plurals that follow the number as shown, such as `1 page`,
`2.3 pages` and `1.0 points`. Dates are formatted by hand, or pinned to `en-US`
and UTC, and screens read today's date and the current year from their loader's
data rather than the clock while rendering, so server HTML and hydration match
in every browser locale. The UI calls the list of logs "activity" and one entry
a "log".

Server responses are parsed before the screen uses them. Production API null
language lists mean unrestricted languages and normalize to empty lists; missing
fields and malformed list members remain invalid. Missing contests,
standings and future rounds have explicit interface states; no sample scores or
invented schedule replace unavailable API data.

The appearance toggle stays disabled until hydration
completes and their browser handlers can respond. Ordinary links work in the
server-rendered HTML, and so do the mobile and account menus: before hydration
Paper's mobile menu trigger is a link that opens its panel through `:target`,
and the account menu is a native disclosure with its pages and Log out.

## Navigation and shell

`frontend/apps/webv3/src/components/AppShell.tsx` gives every page Webv2's
destinations in Paper's layout. The main navigation links Home, Leaderboard,
Contests, Blog and the Guide. Anonymous visitors see Log in and Sign up, which
open authv2's `/login` and `/register` and return to the current page. A
signed-in account sees Log activity and a menu whose trigger is its
[avatar](#avatars) and whose accessible name is the account's name. The menu
opens under a heading with that name, then Profile,
Admin for administrators, Account settings and Log out; at narrow widths the
menu's entries join the mobile navigation. Admin opens the sibling admin host
and Account settings opens authv2's `/settings` with the current page as
`return_to`. Before hydration the menu is a native disclosure, so it opens
without JavaScript; the appearance switch joins it after hydration.

Links to Webv3 pages go through the router: Paper components take
`renderLink`, and screens use `AppLink`, both from
`frontend/apps/webv3/src/components/router-link.tsx`, which keep the path's
query and fragment and leave other origins as plain links. A few links that
must start a fresh document, such as a new draft for another account, stay
plain. After a client navigation that changes the path,
`frontend/apps/webv3/src/router.tsx`
focuses the new page's `h1`, or `main` without one, and announces the page
title through the shell's status region; the first load leaves focus to the
browser. `navigation.spec.ts` covers a home → contest → profile → leaderboard
journey without a document load, that focus and announcement, the profile's
date field, and both menus without JavaScript.

A suspended account sees the suspended screen, with a Discord contact link and
Log out, in place of every page. An unavailable account service names itself
in the header and leaves public pages readable.

A signed-in account's menu, and its section of the phone menu, holds the
theme switch, named for the theme it switches to; the choice is stored in a
cookie on the domain Webv3 and authv2 share, so both follow it, and visitors
without one follow their system choice. The header stays in view while scrolling. The footer sits on the paper surface with a hairline above it: the brand, the
credit line and Tadoku's Discord, GitHub and Twitter beside grouped links to
About, Guide, Rules and Questions & help, and Privacy and Contact. `frontend/apps/webv3/src/components/Announcements.tsx`
shows the CMS's active announcements above each page after hydration, rendering
their Markdown as the blog does, with Learn more unless the announcement links
to the current page. A dismissal is stored in the browser under Webv2's
`dismissed_announcements` key, so it survives the host switch, and is forgotten
once its announcement ends. An unavailable announcement list shows nothing.

### Avatars

`frontend/apps/webv3/src/components/Avatar.tsx` shows a learner's Gravatar
beside their name wherever webv3 names a learner, except in standings and
inside headings, whose text it would join; the account menu shows it alone. The image comes from the API's
`GET /immersion/users/{userId}/avatar` through the gateway on the public web
origin, so the learner's email and its hash never reach the browser. The
avatar is a circle with the learner's initials, one letter at the
smallest size, and the image covers it once it loads. The image loads lazily,
with the width and height of the size it is shown at; the API serves every
avatar at Gravatar's 128 pixels. It stays hidden until it has loaded, so
without JavaScript the initials show. A learner without a
Gravatar and any other failure leave the initials and no broken image, with or
without JavaScript. The avatar is decorative and hidden
from assistive technology; the name beside it carries the meaning.

`/contests` redirects permanently to `/contests/official`, as Webv2 does.
Unknown paths render the not-found page with a link home and status 404.

## Local development

Use Node 22.12 or newer and pnpm. From `frontend/`:

```sh
pnpm install --frozen-lockfile
NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/lab-root-ca.crt pnpm webv3
```

The development server uses port 5183 by default. `HOST` and `PORT` override the
listener. Start a second instance only after checking ownership of existing
listeners. `predev` builds `paper-ui` before serving the application.
Nitro's development runner uses `self` to keep SSR in the Vite process and
avoid racing worker initialization on page requests.
The separate SSR `noExternal` entry for `@tanstack/query-core` bundles Query's
core into server rendering so the shared workspace builds consistently after
adding authv2. It affects the published build; the development runner option
does not.

Local loopback runs cannot use the account cookie scoped to `tadoku.dev.lab`.
The server uses the scheme of `TADOKU_WEB_URL` for return URLs and mutation
origin checks, so an HTTP loopback run with the default HTTPS configuration
supports anonymous reads only. Use an owned HTTPS branch host for authenticated
journeys and mutations.

The Lab CA path above is specific to Lab hosts. On another machine, supply its
trusted Lab CA certificate through `NODE_EXTRA_CA_CERTS`; do not disable TLS
verification.

| Variable | Local default | Purpose |
| --- | --- | --- |
| `TADOKU_API_URL` | `https://tadoku.dev.lab/api/internal` | Server-only Tadoku API root |
| `TADOKU_WEB_URL` | `https://tadoku.dev.lab` | Runtime public web origin; the account and admin hosts are its siblings |
| `HOST` | `0.0.0.0` | Local development listener |
| `PORT` | `5183` | Local development port |

The TanStack server function performs upstream reads, so the browser does not
need cross-origin API cookies or a CORS exception. Account sessions are resolved
server-side through the identity API; public homepage reads remain anonymous.
Local runs default to development API data and account hosts; the
production image defaults both origins to `https://tadoku.app`.

## Branch routing

The development base is `https://next.tadoku.dev.lab`; a dev-cli branch uses
`https://<route>.next.tadoku.dev.lab`. Run `dev up --owner <owner>` from the
repository root, then `dev url --owner <owner> --host next.tadoku.dev.lab '/'`.
The overlay uses Node22 and Vite on port 3000, builds `paper-ui` initially and
watches its source, token generator and assets while Vite provides HMR. Generated
`.output`, `.tanstack` and `dist` files do not enter source synchronization.
The overlay sets `NODE_OPTIONS=--max-old-space-size=768` for the initial Paper declaration
build, within its 1536 MiB pod limit. Its watcher rebuilds JavaScript, the generated theme and
assets without declaration generation or output cleaning, preserving the initial
declaration files and avoiding a declaration worker alongside Vite. Run
`pnpm --filter paper-ui build` after changing a Paper component's public types.
The production Nitro runtime
retains its separately configured heap limit.

`frontend/apps/webv3/src/lib/urls.server.ts` resolves URLs from the configured
runtime origins and the request Host. Development branch reads go through
`https://<route>.tadoku.dev.lab/api/internal`; account, admin and log export links
use that same branch's sibling hosts. Production branches at
`<route>.next.preview.tadoku.app` use `https://<route>.preview.tadoku.app/api/internal`
and the matching sibling hosts.
These public gateways select the branch tenant even when the API itself falls
back to its base deployment. The resolver accepts one valid DNS label beneath
the trusted branch root and rejects malformed or nested branch hosts. Incoming
forwarded-host or branch headers cannot choose an upstream URL.

The root route loader resolves these origins once per request, and SSR
serializes that loader data for hydration. Client navigation and intent preloads
reuse it without another server call. Links are resolved when a screen renders;
they are never frozen into a production build or shared across requests. Log
activity stays on the same Webv3 host, and its login link opens the matching
branch's account host. Canonical `next.tadoku.app` uses production data and the
canonical `account.tadoku.app` and `admin.tadoku.app` hosts.

## Flagged release

[Flagged release](webv3-flagged-release.md) covers sending named accounts from
Webv2 to Webv3 before the host switch, and how to turn it off.

## Checks

[Webv3 checks and browser tests](webv3-testing.md) lists the checks, the
browser suite and the fixture accounts its specs use.

## Production container and CI

[Webv3 production container](webv3-container.md) covers the image, its runtime
variables, startup checks and shutdown, the image smoke check and publication.

## Account sessions

The root loader prefetches the viewer into the request's QueryClient. Viewer data
is separate from the indefinitely fresh sibling URL loader. It becomes stale after
30 seconds and is read again when a stale viewer is next used, for example on
window focus. A request without a `tadoku_session` or `tadoku_session_<tenant>`
cookie makes no upstream session lookup.
`frontend/apps/webv3/src/lib/session.server.ts` forwards every such cookie,
including empty or repeated ones, to the request's API gateway through the
[generated client](api-client.md):
`GET /identity/session` names the account and `GET /authz/current-user/role`
gives its tenant role. The gateway host selects the tenant; no incoming header
can. Anonymous, signed-in, banned and unavailable states are distinct, and a
failed or malformed answer never shows an account.

Loaders start public reads alongside the viewer lookup and use the viewer only
for cache keys and for final not-found and permission decisions; a read cached
per viewer goes through `ensureViewerData` in `frontend/apps/webv3/src/lib/session.ts`.
`readAsVisitor` reads signed out when the API rejects the visitor's session.
A write checks that the session's account opened the form and reads the role
only for administrator access; the API itself refuses banned accounts.

Log in opens authv2's `/login` on the sibling account host, `account.<domain>`
or `<route>.account.<domain>`, with the current page as `return_to`. It
preserves the router path and search during SSR and adds the current fragment
after hydration. Without JavaScript the server cannot preserve a fragment.
Authv2 follows only returns on the identity API's
[app host allowlist](../tadoku-api/identity.md#http-boundaries).

Log out submits a plain POST form to `/logout` from the account menu. After the Origin check below, the
route sends `POST /identity/logout` with the session cookies and the app's own
Origin, relays the API's cookie-clearing `Set-Cookie` and returns home. Any
other answer, including a timeout, shows an unavailable message and keeps the
browser's cookie, although an unconfirmed request may still have revoked the
session. The identity API refuses logout from a host outside its allowlist. Logout and all
non-GET server functions require an exact Origin and reject non-same-origin
Fetch Metadata. Personalized documents and server-function responses emit
`Cache-Control: private, no-store` and `Vary: Cookie`; development ingress
replaces Cache-Control with its own `no-store` policy. Health keeps its own headers.

The account E2E logs in through authv2 with two native fixture identities and
makes anonymous requests, including thirty interleaved SSR reads. On a branch it
adds an unrelated canonical `tadoku_session` cookie and expects the branch
account to remain signed in. After logout it checks that the API no longer
accepts the old session cookie. Set
`TADOKU_API_URL` to the branch's public API root for that check. Authenticated
tests disable traces and videos and retain no browser storage.

## Server logs

Server failures are recorded as [structured JSON lines](webv3-server-logs.md).
