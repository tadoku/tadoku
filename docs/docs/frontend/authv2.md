---
title: Authv2
description: The TanStack Start and Paper account skeleton, request-scoped session selection, local smoke test, dev-cli overlays and container runtime.
---

# Authv2

Read this when you run or extend `frontend/apps/authv2`.

Authv2 is the account application foundation on the [Webv3 stack](webv3.md):
TanStack Start, Router, Query and Form, Zod, Vite, Nitro and Paper. `/` and
`/login` render the skeleton. Registration, settings, recovery and account
mutations are not implemented here. The existing `auth` app remains available.

## Session and account selection

`frontend/apps/authv2/src/lib/session.server.ts` requests `GET /identity/session`
from the configured Tadoku API root, forwarding only the incoming Cookie and
`X-Dev-Branch` request headers. Zod validates the response shape; the API owns
account field constraints, so legacy display names and email formats do not block
the v1 redirect. Session data
is request-scoped; both the upstream read and rendered response disable caching.
The server function in `frontend/apps/authv2/src/lib/session.ts` keeps server
configuration and request credentials out of the browser bundle.

The API chooses the account app for the request's tenant. A `v1` answer redirects
to the existing account origin with the current path and query. A `v2` answer
renders the skeleton with the session's signed-in or anonymous state. A missing
endpoint, non-success response, timeout or invalid payload shows an unavailable
message and a link to the existing account app. An unavailable answer does not
invent an account-app selection or an authenticated account.

`frontend/apps/authv2/src/lib/urls.server.ts` resolves origins per request.
Development uses `authv2.tadoku.dev.lab` and
`<route>.authv2.tadoku.dev.lab`; production uses `authv2.tadoku.app` and
`<route>.authv2.preview.tadoku.app`. A valid branch label selects the sibling
Webv2 API gateway and existing account origin. Other hosts retain the configured
origins. Forwarded host headers cannot select an upstream.

## Local development

Use Node 22.12 or newer and pnpm. From `frontend/`:

```sh
pnpm install --frozen-lockfile
NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/lab-root-ca.crt pnpm authv2
```

`predev` builds `paper-ui`; Vite listens on port 5184. Check existing listener
ownership before starting another instance.

| Variable | Default | Purpose |
| --- | --- | --- |
| `TADOKU_WEB_URL` | `https://tadoku.dev.lab` | Webv2 origin and trusted sibling-host root |
| `TADOKU_API_URL` | `https://tadoku.dev.lab/api/internal` | Server-only API root |
| `HOST` | `0.0.0.0` | Listener address |
| `PORT` | `5184` | Vite listener port |

## Branch overlays and base

Use [Development environment](../develop/environment.md) to start a branch and
obtain its authv2 link. Authv2 source changes select its deployable; shared
`paper-ui` changes select both authv2 and Webv3. The Paper runtime uses Node 22,
builds shared declarations initially, watches Paper assets and runs the selected
app's Vite server on port 3000. Public Paper type changes require another full
`paper-ui` build. API and tenant provisioning remain separate responsibilities.

The base lives in `k8s/dev/base/frontend-authv2/`, with a native Nitro server,
Lab CA trust, API-independent `/healthz` probes and canonical gateway routes
that strip client-supplied branch headers. The development base becomes
available only after its manifests and compatible image have been deployed.

## Checks and manual browser smoke

From `frontend/`:

```sh
pnpm --filter authv2 exec tsc --noEmit
pnpm --filter authv2 lint
pnpm --filter authv2 build
pnpm check:paper-boundaries
AUTHV2_BASE_URL=http://127.0.0.1:5184 pnpm --filter authv2 test:e2e
```

The single Playwright smoke test is run by hand, never in CI. Use a tenant on
v2, or a local server whose session endpoint is absent. It checks the rendered
skeleton, mobile containment and health. See
`frontend/apps/authv2/e2e/README.md` for the loopback session fixture that also
checks header forwarding, v1 redirection and failure responses. That fixture
proves application HTTP/SSR behavior; it does not prove real authentication,
Oathkeeper, tenant isolation or shared-cluster routing. TLS verification remains
enabled. Reports and screenshots go outside source control.

## Images and deployment

`frontend/apps/authv2/Dockerfile` builds the standalone Node 22 Nitro runtime.
It runs as UID/GID 1000 on port 3000 with a read-only root filesystem and no
Kubernetes API token. `/healthz` remains healthy when the identity API is absent.
The production image defaults runtime origins to `https://tadoku.app` and
`https://tadoku.app/api/internal`; development overrides them through ConfigMaps.

`frontend/scripts/publish-image.sh` and `//frontend:authv2_branch_push` publish
explicitly requested tags from committed source. The authv2 build workflow
checks types, lint, boundaries and the container without running Playwright.
Main publication creates commit and `latest` tags; only an owner-approved manual
workflow dispatch with `publish_production` creates `prod`. The maintainer owns
production manifests, DNS, certificate readiness and deployment in the separate
GitOps repository. Preparing manifests does not activate the application.
