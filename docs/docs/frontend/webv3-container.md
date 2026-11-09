---
title: Webv3 production container
description: The Webv3 container image, its runtime variables and startup checks, how it shuts down, the image smoke check and how main publishes and rolls it out.
sidebar_position: 3
---

# Webv3 production container

Read this when you change Webv3's Dockerfile, runtime variables, startup or
shutdown, or its image smoke check and publication.

`frontend/apps/webv3/Dockerfile` builds a standalone Nitro Node22 server for Linux
amd64, including traced runtime dependencies. It runs as UID/GID 1000 on port
3000 and supports a read-only root filesystem. `/healthz` returns `200` without
calling the API; use it for readiness and liveness rather than the homepage.
`TADOKU_API_URL` is read at runtime and defaults in the image to
`https://tadoku.app/api/internal`; a runtime environment override selects another
Tadoku API. `TADOKU_WEB_URL` is also read at runtime and defaults in the image to
`https://tadoku.app`. Development deployments override both variables.
`frontend/apps/webv3/server/plugins/runtime-env.ts` checks these URLs, `PORT`,
`HOST`, `SERVER_SHUTDOWN_TIMEOUT` and any `NODE_EXTRA_CA_CERTS` file before the
server listens. A bad value makes the process exit with status 1 and a stderr
line naming the variable, so a misconfigured pod never becomes ready. The line
never prints the URL variables' values, which can carry credentials.
On `SIGTERM` the server stops accepting connections and finishes in-flight
requests for up to `SERVER_SHUTDOWN_TIMEOUT` seconds, 15 in the image, which is
longer than the 8-second upstream deadline. Because it stops accepting at once,
the Deployment needs a `preStop` delay so the edge removes the pod before it
receives `SIGTERM`.
The [production config contract](../develop/ci-workflows.md#production-config-contract)
fails CI when these variables, the canonical hosts or `/healthz` are renamed or removed.
Never pass credentials as build arguments. Supply any private CA certificate and
`NODE_EXTRA_CA_CERTS` at runtime when the upstream requires it.

From `frontend/`:

```sh
docker build --platform linux/amd64 -f apps/webv3/Dockerfile -t webv3:check .
pnpm --filter webv3 exec playwright install chromium
PLAYWRIGHT_OUTPUT_DIR=/tmp/webv3-evidence node scripts/smoke-webv3-image.mjs webv3:check
```

The smoke command starts its own labeled container and synthetic GET-only API
(`frontend/scripts/smoke-webv3-fixture.mjs`), including one contest with its
logs and one standings learner's profile year and contest profile,
uses nullable language lists by default (`WEBV3_LANGUAGE_FIXTURE=array` selects
nonempty lists), checks server rendering, built assets, non-root/read-only operation and the
existing browser journeys, then disconnects the API and checks health again.
It also checks that a container with a bad `TADOKU_API_URL`, or with
`NODE_EXTRA_CA_CERTS` naming a directory, exits non-zero naming the variable,
and that `docker stop` lets a page waiting 6.5 seconds on the API finish.
It also requests the guide, `/pages/manual`, concurrently with canonical and branch Host headers
in production and development runtime modes, checking that account links stay
with their matching origins, native links stay on the requested host and
forwarded-host or branch headers cannot override the request Host. A second container verifies development runtime
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
the `prod` digest automatically through Image Updater and Argo CD. Webv3 uses the same production
backend as Webv2; the homepage, leaderboards, contest lists and contest pages read contest data, content pages read the CMS and blog, native logging, contest registration and log moderation use the existing account, activity and contest API, and contest creation and scoring use its contest and scoring API. No product
journey opens Webv2.
Merging source that matches the Webv3 workflow's paths, including the shared
lockfile, therefore republishes `latest` and `prod` and rolls this production
preview through Image Updater and Argo CD's automatic sync.
Require `webv3 / container journeys` on the exact reviewed source head before
landing such changes. An authv2 production hold does not hold this Webv3 pipeline.

Before the first deployment, the package owner must confirm GHCR package
visibility and cluster pull access. Newly created packages may be private;
public unauthenticated pulls require making this package public, otherwise the
cluster needs an authorized image pull credential.
