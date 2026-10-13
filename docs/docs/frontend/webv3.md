---
title: Webv3
description: The unpublished TanStack Start homepage foundation, its read-only development API, local goal form and development checks.
sidebar_position: 3
---

# Webv3

Read this when you run or extend `frontend/apps/webv3`.

Webv3 is a small, unpublished application foundation. Its homepage follows the
Paper playground composition and reads the current official contest, summary,
standings and next scheduled round from the existing development API. `/plan`
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
  existing development Webv2 application.
- `frontend/apps/webv3/src/components/` owns the Paper screen composition,
  navigation, appearance and local goal interactions.
- The goal form uses TanStack Form with native controls and Paper CSS recipes.
  Existing Paper form components retain their React Hook Form integration.

Server responses are parsed before the screen uses them. Missing contests,
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

| Variable | Default | Purpose |
| --- | --- | --- |
| `TADOKU_API_URL` | `https://tadoku.dev.lab/api/internal` | Server-only public API root |
| `VITE_TADOKU_WEB_URL` | `https://tadoku.dev.lab` | Existing Webv2 origin for product links |
| `HOST` | `0.0.0.0` | Local development listener |
| `PORT` | `5183` | Local development port |

The TanStack server function performs upstream reads, so the browser does not
need cross-origin API cookies or a CORS exception. The app has no account session,
API mutation, deployment image, Kubernetes manifest or publication workflow.

## Checks

From `frontend/`, with Node 22 and the development server running:

```sh
pnpm --filter webv3 typecheck
pnpm --filter webv3 lint
pnpm --filter webv3 build
pnpm check:paper-boundaries
NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/lab-root-ca.crt pnpm --filter webv3 test:e2e
```

The browser suite in `frontend/apps/webv3/e2e/home.spec.ts` checks real development
data in server-rendered HTML, hydration and mobile containment, local goal
validation/editing/persistence without network writes, unavailable browser
storage, theme persistence, mobile navigation and unknown routes.
It requires reachable development fixtures and Chromium installed through
Playwright. It preserves HTML reports and screenshots outside the source tree.
`WEBV3_BASE_URL` selects the running server; `PLAYWRIGHT_OUTPUT_DIR` selects the
artifact directory, defaulting to a temporary directory.

`pnpm --filter webv3 build` creates a local Node build in `.output`. This is a build
check; it does not publish or deploy the application.
