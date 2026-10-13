---
title: Webv3
description: The TanStack Start homepage foundation, read-only API, local goal form, production container and verification workflow.
sidebar_position: 3
---

# Webv3

Read this when you run or extend `frontend/apps/webv3`.

Webv3 is a small application foundation. Its homepage follows the
Paper playground composition and reads the current official contest, summary,
standings and next scheduled round from the configured read-only API. `/plan`
demonstrates a validated immersion goal kept in the current browser. It does not
create logs or change a Tadoku account.

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
- The goal form uses TanStack Form with native controls and Paper CSS recipes.
  Existing Paper form components retain their React Hook Form integration.

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

The Lab CA path above is specific to Lab hosts. On another machine, supply its
trusted Lab CA certificate through `NODE_EXTRA_CA_CERTS`; do not disable TLS
verification.

| Variable | Local default | Purpose |
| --- | --- | --- |
| `TADOKU_API_URL` | `https://tadoku.dev.lab/api/internal` | Server-only public API root |
| `VITE_TADOKU_WEB_URL` | `https://tadoku.dev.lab` | Existing Webv2 origin for product links |
| `HOST` | `0.0.0.0` | Local development listener |
| `PORT` | `5183` | Local development port |

The TanStack server function performs upstream reads, so the browser does not
need cross-origin API cookies or a CORS exception. The app has no account session or
API mutation. Local runs default to development API data and Webv2 links; the
production image defaults both origins to `https://tadoku.app`.

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
read-only API. `VITE_TADOKU_WEB_URL` is a public build
argument, defaulting to `https://tadoku.app`; development defaults remain unchanged.
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
It retains a browser report, container log and result with source revision and
fixture boundary. This proves container behavior; it does not verify real API,
authentication, DNS, ingress or deployment. Docker on Linux, Node22 and installed pnpm
are required. The command removes only its own container.

`.github/workflows/build-frontend-webv3.yaml` runs checks and container journeys
on pull requests. Main publishes immutable commit and `latest` tags to
`ghcr.io/tadoku/tadoku/frontend-webv3`. An owner-approved manual dispatch on main
with `publish_production: true` also promotes the image built from that verified commit to `prod`. The publication
job rebuilds the image; its digest can differ from the smoke-checked image.
No pull request publishes an image. Publication does not deploy it.

Before the first deployment, the package owner must confirm GHCR package
visibility and cluster pull access. Newly created packages may be private;
public unauthenticated pulls require making this package public, otherwise the
cluster needs an authorized image pull credential.
