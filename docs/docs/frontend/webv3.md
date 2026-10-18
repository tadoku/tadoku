---
title: Webv3
description: The TanStack Start homepage, request-scoped API and sibling links, local goal form, dev-cli overlays, container and preview deployment workflow.
sidebar_position: 3
---

# Webv3

Read this when you run or extend `frontend/apps/webv3`.

Webv3 is a small application foundation. Its homepage follows the
Paper playground composition and reads the current official contest, summary,
standings and next scheduled round from the configured Tadoku API. `/plan`
demonstrates a validated immersion goal kept in the current browser. It does not
sync the goal to a Tadoku account. Native activity logging uses the existing account session and API.

## Stack and ownership

The application uses TanStack Start, Router, Query and Form, Zod, Vite and Nitro's
Node runtime. React remains on the workspace's 18.2 version so the Paper and
legacy applications keep their existing compatibility.

- `frontend/apps/webv3/src/router.tsx` creates a QueryClient per router/request and
  integrates server rendering and hydration.
- `frontend/apps/webv3/src/lib/home.ts` owns response schemas and the homepage
  query. `frontend/apps/webv3/src/lib/home.server.ts` makes public GET requests to
  the configured API; it accepts no browser-supplied upstream URL or credentials.
- `frontend/apps/webv3/src/lib/links.ts` sends unfinished product journeys to the
  configured existing Webv2 application.
- `frontend/apps/webv3/src/components/` owns the Paper screen composition,
  navigation, appearance and local goal interactions.
- The goal form uses `usePaperForm` and bound Paper controls through TanStack
  Form's composition contexts. Its Zod validators own numeric bounds and errors.
  Saving and storage failures remain application-owned effects.

Server responses are parsed before the screen uses them. Production API null
language lists mean unrestricted languages and normalize to empty lists; missing
fields and malformed list members remain invalid. Missing contests,
standings and future rounds have explicit interface states; no sample scores or
invented schedule replace unavailable API data.

The mobile menu, appearance toggle and local goal controls stay disabled until
hydration completes and their browser handlers can respond. Ordinary links work
in the server-rendered HTML.

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
| `TADOKU_WEB_URL` | `https://tadoku.dev.lab` | Runtime Webv2 origin for product links |
| `HOST` | `0.0.0.0` | Local development listener |
| `PORT` | `5183` | Local development port |

The TanStack server function performs upstream reads, so the browser does not
need cross-origin API cookies or a CORS exception. Account sessions are resolved
server-side through the identity API; public homepage reads remain anonymous.
Local runs default to development API data and Webv2 links; the
production image defaults both origins to `https://tadoku.app`.

## Branch routing

The development base is `https://next.tadoku.dev.lab`; a dev-cli branch uses
`https://<route>.next.tadoku.dev.lab`. Run `dev up --owner <owner>` from the
repository root, then `dev url --owner <owner> --host next.tadoku.dev.lab '/'`.
The overlay uses Node22 and Vite on port 3000, builds `paper-ui` initially and
watches its source, styles and assets while Vite provides HMR. Generated
`.output`, `.tanstack` and `dist` files do not enter source synchronization.
The overlay sets `NODE_OPTIONS=--max-old-space-size=768` for the initial Paper declaration
build, within its 1536 MiB pod limit. Its watcher rebuilds JavaScript, styles and
assets without declaration generation or output cleaning, preserving the initial
declaration files and avoiding a declaration worker alongside Vite. Run
`pnpm --filter paper-ui build` after changing a Paper component's public types.
The production Nitro runtime
retains its separately configured heap limit.

`frontend/apps/webv3/src/lib/urls.server.ts` resolves URLs from the configured
runtime origins and the request Host. Development branch reads go through
`https://<route>.tadoku.dev.lab/api/internal`; product links use that same branch's
Webv2 origin. Production branches at `<route>.next.preview.tadoku.app` use
`https://<route>.preview.tadoku.app/api/internal` and the matching Webv2 links.
These public gateways select the branch tenant even when the API itself falls
back to its base deployment. The resolver accepts one valid DNS label beneath
the trusted branch root and rejects malformed or nested branch hosts. Incoming
forwarded-host or branch headers cannot choose an upstream URL.

The root route loader resolves the product-link origin once per request, and SSR
serializes that loader data for hydration. Client navigation and intent preloads
reuse it without another server call. Links are resolved when a screen renders;
they are never frozen into a production build or shared across requests. Opening
Log activity reaches the matching Webv2 branch, whose login flow keeps the same
tenant selection. Canonical `next.tadoku.app` uses production data and the
canonical `tadoku.app` product links.

## Checks

From `frontend/`, with Node 22 and the development server running:

```sh
pnpm --filter webv3 typecheck
pnpm --filter webv3 lint
pnpm --filter webv3 build
pnpm check:paper-boundaries
NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/lab-root-ca.crt pnpm --filter webv3 test:e2e
```

The browser suite in `frontend/apps/webv3/e2e/home.spec.ts` checks configured read-only
data in server-rendered HTML, hydration and mobile containment, local goal
validation/editing/persistence without network writes, unavailable browser
storage, theme persistence, mobile navigation and unknown routes.
It requires reachable contest and standings data and Chromium installed through
Playwright. It preserves HTML reports and screenshots outside the source tree.
`WEBV3_BASE_URL` selects the running server; `PLAYWRIGHT_OUTPUT_DIR` selects the
artifact directory, defaulting to a temporary directory.

`pnpm --filter webv3 build` creates a local Node build in `.output`. This is a build
check; it does not publish or deploy the application.

## Production container and CI

`frontend/apps/webv3/Dockerfile` builds a standalone Nitro Node22 server for Linux
amd64, including traced runtime dependencies. It runs as UID/GID 1000 on port
3000 and supports a read-only root filesystem. `/healthz` returns `200` without
calling the API; use it for readiness and liveness rather than the homepage.
`TADOKU_API_URL` is read at runtime and defaults in the image to
`https://tadoku.app/api/internal`; a runtime environment override selects another
Tadoku API. `TADOKU_WEB_URL` is also read at runtime and defaults in the image to
`https://tadoku.app`. Development deployments override both variables.
Never pass credentials as build arguments. Supply any private CA certificate and
`NODE_EXTRA_CA_CERTS` at runtime when the upstream requires it.

From `frontend/`:

```sh
docker build --platform linux/amd64 -f apps/webv3/Dockerfile -t webv3:check .
pnpm --filter webv3 exec playwright install chromium
PLAYWRIGHT_OUTPUT_DIR=/tmp/webv3-evidence node scripts/smoke-webv3-image.mjs webv3:check
```

The smoke command starts its own labeled container and synthetic GET-only API,
uses nullable language lists by default (`WEBV3_LANGUAGE_FIXTURE=array` selects
nonempty lists), checks server rendering, built assets, non-root/read-only operation and the
existing browser journeys, then disconnects the API and checks health again.
It also requests `/plan` concurrently with canonical and branch Host headers
in production and development runtime modes, checking that product links stay
with their matching origins and that forwarded-host or branch headers cannot
override the request Host. A second container verifies development runtime
overrides against the same production-built image.
It retains a browser report, container log and result with source revision and
fixture boundary. This proves container behavior; it does not verify real API,
authentication, DNS, ingress or deployment. Docker on Linux, Node22 and installed pnpm
are required. The command removes only its own container.

`.github/workflows/build-frontend-webv3.yaml` runs checks and container journeys
on pull requests. Main publishes immutable commit and `latest` tags to
`ghcr.io/tadoku/tadoku/frontend-webv3`. The publication job rebuilds the image;
its digest can differ from the smoke-checked image. No pull request publishes an
image.

The active production GitOps configuration makes `next.tadoku.app` follow
the `latest` digest automatically through Image Updater and Argo CD. No manual
production tag promotion is required for Webv3. Webv3 uses the same production
backend as Webv2; the current homepage reads public data and native logging uses its existing account and activity API. Unfinished journeys
continue to open production Webv2.
Merging source that matches the Webv3 workflow's paths, including the shared
lockfile, therefore republishes `latest` and rolls this production preview through
Image Updater's periodic digest reconciliation and Argo CD's automatic sync.
Require `webv3 / container journeys` on the exact reviewed source head before
landing such changes. An authv2 production hold does not hold this Webv3 pipeline.

Before the first deployment, the package owner must confirm GHCR package
visibility and cluster pull access. Newly created packages may be private;
public unauthenticated pulls require making this package public, otherwise the
cluster needs an authorized image pull credential.

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

Log in opens authv2's `/login` on the sibling account host, `account.<domain>`
or `<route>.account.<domain>`, with the current page as `return_to`. It
preserves the router path and search during SSR and adds the current fragment
after hydration. Without JavaScript the server cannot preserve a fragment.
Authv2 follows only returns on the identity API's
[app host allowlist](../tadoku-api/identity.md#http-boundaries).

Log out is a plain POST form to `/logout`. After the Origin check below, the
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

## Native activity logging

`/logs/new` prompts anonymous visitors to sign in through the existing account
application, then loads configuration and ongoing contest registrations from the
selected tenant. TanStack Form owns the draft. Language and activity determine
the available unit or duration input and eligible contests. Eligible contests are
selected initially; a user can uncheck them to save a personal record.

`frontend/apps/webv3/src/lib/logs.ts` validates responses and form values with
Zod. Its server counterpart forwards only the session cookie to the existing
configuration, registrations, create and detail endpoints. The API owns scoring,
contest eligibility, authorization and persistence; Webv3 does not read the
legacy log-entry feature flag or calculate a provisional score.

A successful save navigates to `/logs/$logId` using the returned ID and displays
the persisted language, measure, tags, score and contests. Reload reads that
record again. An expired session cookie retries a public detail read anonymously. Public detail omits private contest submissions; only the owner
gets the contest management link to Webv2. History, edit, delete and score preview
remain Webv2 journeys.

Invalid input focuses the first invalid control. A pending save disables controls
and repeated submits. Rejected saves keep the draft. Each save checks the current session against the
account that started the draft; an account switch is refused before the API
write, updates the header and offers an explicit new draft for the new account. An expired session offers
login in a new tab so the original form survives. An ambiguous transport failure
asks the user to check their activity before retrying; the API has no idempotency
key, so retrying an ambiguous save can create another record. Drafts are held in
memory and are lost when the form is reloaded or closed.

Run `logging.spec.ts` with `WEBV3_BASE_URL`, `TADOKU_API_URL`,
`E2E_READER_EMAIL` and `E2E_READER_PASSWORD` for the documented development
fixture. Ownership and branch-only suspension cases also use
`E2E_SECOND_EMAIL` and `E2E_SECOND_PASSWORD` for Dev Admin; the suspension case
restores the Reader role in the same branch tenant. Set `NODE_EXTRA_CA_CERTS` to the trusted Lab CA and
`PLAYWRIGHT_OUTPUT_DIR` outside the checkout. Authenticated runs disable traces
and videos; browser-boundary delay and abort cases cover repeated submits and
transport errors without mocking the API's successful writes.
