---
title: Development environment
description: How to install dev-cli and run, open, seed, verify and clean up your branch of Tadoku on the shared homelab-dev cluster.
---

# Development environment

Read this when you want to run your branch of webv2, auth, admin, Tadoku API or its worker
on the shared development cluster, or check that a change works there.

dev-cli deploys live branch overlays of webv2, auth, admin, Tadoku API and its worker to the
`homelab-dev` Kubernetes cluster. Argo CD keeps a shared base of every service
running there even when no developer has a loop running; see
[Development base](../operations/development-base.md). There is no local
Kubernetes cluster or Helm bootstrap to run. For a map of the components, see
[System architecture](../architecture/index.md). Production is deployed from a private
repository and is not covered here.

`.dev/config.yaml` is the real, committed, non-secret configuration; no hostname
substitution is needed.

| Setting | Value |
| --- | --- |
| Kubernetes context | `homelab-dev` |
| Hosts | `tadoku.dev.lab`, `account.tadoku.dev.lab`, `admin.tadoku.dev.lab` |
| Overlay image registry | `registry.dev.lab/tadoku-dev-cli` |
| Overlay TTL | 8 hours |

## Prerequisites

- Git access to Tadoku and to the private `antonve/dev-cli` repository.
- Go to install dev-cli, Bazelisk or Bazel at the repository's pinned version,
  and kubectl with authorized access to the `homelab-dev` context.
- Node and pnpm for frontend tools; Docker for local container-based checks.
- Lab network and DNS access, and trust in the Lab CA, including for the
  development registry at `registry.dev.lab`.

The cluster already supplies Postgres, the shared auth providers and routing.
Obtain kube access from the operator. Never commit kubeconfigs, credentials or
private keys.

## Install dev-cli

Install a dev-cli release that supports route-free `worker` deployables and
`selectionGroup` companions. Version 0.4.0 supports the YAML configuration,
multi-host routing and dependency/task workflow, but does not support the
worker declaration in this repository.

```sh
GOPRIVATE=github.com/antonve/dev-cli go install github.com/antonve/dev-cli/cmd/dev@latest
dev version  # must include worker and selectionGroup support
```

Go must be able to authenticate to the private repository. If your Git
credentials are SSH-only, rewrite the URL for this one command; no new
credential is required:

```sh
GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=url.git@github.com:.insteadOf \
GIT_CONFIG_VALUE_0=https://github.com/ GOPRIVATE=github.com/antonve/dev-cli \
go install github.com/antonve/dev-cli/cmd/dev@latest
```

- Put `$(go env GOPATH)/bin`, or your explicit `GOBIN`, on `PATH`. Check
  `command -v dev` and `dev version` so you do not run an older installation.
- To upgrade an existing installation in place, set `GOBIN` to its directory
  on the same command.
- For a reproducible pin, replace `@latest` with the released version that
  includes worker and selectionGroup support.
- Stop only your own running loops before upgrading.

Then check the prerequisites:

```sh
git fetch origin main
dev doctor
```

`dev doctor` is a read-only prerequisite check. It deploys nothing and does not
test routing or login.

## Start a branch

Create a feature branch and make your service edits before starting the loop.
If the shared fixture accounts are missing, run `make dev-seed` first (see
[Seed data](#seed-data)).

```sh
dev up --owner alice --task migrate --task seed
```

Keep this terminal running.

- Frontend sources sync into a pnpm-managed Next.js dev server with hot module
  replacement (HMR).
- Go edits rebuild the affected Bazel binary and restart it in the same pod.
  A failed compilation keeps the last working process running.
- A Tadoku API or worker selection starts both workloads against the same
  branch database and cache prefix. The unchanged peer keeps its current image;
  only a changed binary restarts on a live edit.
- dev-cli builds overlay images on demand, pushes them to the development
  registry and deploys them by immutable digest.
- The `migrate` and `seed` tasks prepare your branch database before the API
  starts; see [Branch databases](#branch-databases).

**Owner** is a stable label for a developer or worktree, not authentication.
Use a different owner for each checkout you run at the same time.

### Which services start

dev-cli asks Bazel which deployables changed relative to the merge base with
`origin/main`, including uncommitted edits. It starts overlays only for those;
every other service keeps using the base.

- Auth-only or admin-only edits select only that app. Shared `packages/ui`
  edits select all three frontends.
- Unknown paths conservatively select every deployable.
- `--base <ref>` changes the comparison ref.
- `--service <name>` adds a deployable to the affected set; it does not filter
  the set.
- Tadoku API and tadoku-worker form one selection group. Either one starts both
  workloads; the worker is private and has no browser route.
- Discovery happens once at startup. Restart the loop to add a service.
- `--no-watch` does not provide live updates.

Frontend-only work can omit the tasks: `dev up --owner alice`. A frontend-only
overlay reads and writes base data through the base API. To keep writes in a
branch database, add the API:
`dev up --owner alice --service tadoku-api --task migrate --task seed`.

Token-reflector is base-only. Paper styleguide has no live overlay; run it
locally with `cd frontend && pnpm paper-styleguide`.

### Make wrappers

| Target | Runs |
| --- | --- |
| `make dev-up` | `dev up --task migrate --task seed` |
| `make dev-logs` | `dev logs` |
| `make dev-down` | `dev down` |
| `make dev-seed` | `scripts/dev/seed-db.sh` |
| `make dev-reset` | Nothing; it is disabled and exits with an error |

Set `DEV_OWNER` consistently for the wrappers, for example
`DEV_OWNER=alice make dev-up`.

## Open and inspect your branch

In another terminal, in the same checkout:

```sh
dev status --owner alice
dev url --owner alice '/'
dev url --owner alice --host account.tadoku.dev.lab '/login'
dev url --owner alice --host admin.tadoku.dev.lab '/'
dev logs --owner alice tadoku-api
dev logs --owner alice tadoku-worker
```

Flags come before positional paths and service names.

Open the printed link, including for deep links. Its `dev-branch` query
parameter makes Envoy set a host-only branch cookie; no application branch menu
or copied branch string is needed. The parameter stays in the URL and wins over
an older cookie. A printed link does not create an overlay or prove that it is
ready.

- **Selection is per hostname.** Open the printed link for each host your test
  uses. To test an API overlay from admin, open both the admin link and the
  main-host link in the same browser profile.
- Tabs in one browser profile share the selection; separate profiles have
  independent selections.
- Clear a selection with `dev url --owner alice --clear '/'`, adding `--host`
  for the account or admin host. Clearing one hostname does not clear others.
- Without a selection, browsers use the base at `https://tadoku.dev.lab`,
  `https://account.tadoku.dev.lab` and `https://admin.tadoku.dev.lab`.
- Each service independently uses your overlay when it is selected and
  healthy, and the base otherwise.
- Kratos stays shared. `/kratos` is never routed to a frontend overlay.

### Routing headers

Selected routes report their routing in response headers:

| Header | Meaning |
| --- | --- |
| `X-Dev-Selected` | The requested selection (intent only) |
| `X-Dev-Backend` | The upstream that actually served the request |
| `X-Dev-Proxy-Backend` | The outer Oathkeeper hop, not necessarily the final API |

Unselected static base routes need not emit these headers. Check
`X-Dev-Backend` before testing overlay behavior: cold-start DNS and health
convergence can serve the base for a while after `dev up` prints a link.

### How requests reach your overlay

```text
browser → ingress-nginx → Envoy → webv2 / auth / admin
                             → Oathkeeper → Envoy → Tadoku API
```

- The browser cookie wins over any routing header a client supplies.
- Envoy normalizes the branch selection into an internal header. The
  development Oathkeeper, which uses development-only signing credentials and
  auth providers, propagates it to the API hop.
- Base and branch frontends send server-side rendering requests through the
  same gateway with Lab CA trust, so an API-only selection also applies to
  server-rendered pages.
- Kratos already allows the development hostnames; no temporary auth allowlist
  is needed.

## Databases, migrations and seed data

### Shared base data

Argo CD runs the base migrations automatically during full syncs; see
[Development base](../operations/development-base.md#automatic-migrations).
Selective resource sync skips those migration hooks, so never use it for a
release. Kratos and Keto are shared by the base and every branch.

### Seed data

`make dev-seed` runs `scripts/dev/seed-db.sh` against the shared base. It:

- runs only against the `homelab-dev` context and waits for Postgres and the
  base migrations;
- creates the two fixture identities in Kratos, marked as owned by the seed,
  and refuses to touch an existing identity with the same email but no marker;
- resets the fixture passwords to the configured values on every run;
- grants the administrator role in Keto;
- loads the fixtures in `scripts/dev/seed/` into the base `tadoku` database.

It is safe to rerun, but it changes shared identities and base data. See
[Authorization](../architecture/authorization.md#seeding-an-administrator-in-development)
for the role details.

| Account | Email | Role |
| --- | --- | --- |
| Dev Admin | `dev@tadoku.app` | Administrator |
| Dev Reader | `reader@tadoku.app` | Reader |

The development fixture password is `tadoku`. Override the emails and passwords
outside Git with `TADOKU_DEV_ADMIN_EMAIL`, `TADOKU_DEV_ADMIN_PASSWORD`,
`TADOKU_DEV_READER_EMAIL` and `TADOKU_DEV_READER_PASSWORD`.

### Branch databases

The operator-managed `tdk-dev-data/tadoku-dev-db` is the only Postgres server.
Its `tadoku`, `kratos` and `keto` databases are shared. Each Tadoku API overlay
uses its own `tadoku-<route>` database on the same server. The route contains
the owner and branch plus a collision-resistant hash. No branch creates another
Postgres cluster, PVC or long-running database pod.

- A short-lived dependency Job creates the database idempotently, owned by the
  `tadoku` role and marked with its route. It refuses to adopt an existing
  database without that marker. Only this Job references the Postgres
  administrator Secret.
- The API, `migrate` and `seed` use the existing `tadoku` role. Credentials
  stay Secret references.
- `dev up --task migrate --task seed` creates the database first; a task
  failure prevents overlay startup. Plain `dev up` runs no tasks.
- The branch `seed` task reuses the shared fixture identities and fails until
  `make dev-seed` has created them. It writes only into an owned branch database.

Rerun the tasks explicitly with:

```sh
dev task --owner alice migrate
dev task --owner alice seed
```

Tasks publish their Bazel images and serialize against the exact branch
database. They never reset a database. Never delete a task Lease to force a
retry; first prove that the abandoned holder has stopped.

This is cooperative development isolation, not a hostile-tenant boundary: the
`tadoku` role, Kratos, Keto and Valkey are shared. Services without an overlay
keep using base data, so deploy related services together when a test needs
consistent data across them.

### Leaderboards on shared Valkey

Each branch API runs its own leaderboard outbox worker against its branch
database. `.dev/tadoku-api.yaml` sets `API_LEADERBOARD_CACHE_PREFIX` to
`dev:${DEV_ROUTE}:`, which keeps every branch cache key and startup scan
separate from the base and from other branches. The base API uses unprefixed
keys.

- Keep the prefix unique per route when you change overlay routing.
- Tadoku API rejects a cache prefix unless the outbox worker is enabled.
- Do not disable a branch worker while its cache reads stay active: its
  leaderboards go stale and its outbox stays pending.

## Clean up

```sh
dev down --owner alice
```

`dev down` stops the local loop and removes only the current owner and branch's
overlays and disposable Jobs. It does not stop the Argo CD base. Stopping the
loop with Ctrl-C alone leaves the overlays running.

- Overlays expire after their TTL. `dev cleanup` removes expired overlays for
  every owner, not only yours; `dev status` and `dev up` can also maintain
  expired overlays.
- **Branch databases are retained** after `dev down` and after TTL cleanup.
  No automatic drop exists. Deleting one requires inspecting the exact database
  and its ownership, then separate authorization.
- Never drop the shared `tadoku`, `kratos` or `keto` databases, and never
  run namespace-wide deletion or cleanup.
- `make dev-reset` is disabled. Any reset needs an explicitly approved, scoped
  runbook.

## Troubleshooting

- Doctor checks prerequisites, not end-to-end routing. Inspect `dev status`,
  task logs, route admission and real responses; use service-filtered
  `dev logs` for sync and build errors.
- Check the selected hostname and `X-Dev-Backend` before diagnosing stale
  content.
- Successful backend restarts can briefly return 503 while health-based
  fallback catches up; this is not zero-downtime deployment. Compile failures
  leave the old process up.
- With `dev up` running, replacement pods receive the current source again.
  Without the loop they start from their image, not from the lost writable
  container layer.
- Removing one overlay preserves other owners and the base. Routing
  convergence can briefly serve the base.
- If an ignored `.dev/config.json` exists next to `.dev/config.yaml`, dev-cli
  rejects the ambiguous defaults. Review it and move it into an override file.
- For a shared-base failure, inspect the Argo CD application and follow
  [Development base](../operations/development-base.md). Do not restart or
  delete shared databases as a troubleshooting shortcut.

## Overrides

Put local overrides in an ignored file and pass it with
`dev <command> --config <path>`. Never commit credentials, private keys,
kubeconfigs or tokens.

## Verify a change

Before opening a pull request:

```sh
bazel mod deps --lockfile_mode=error
bazel run //:gazelle -- -mode=diff
bazel build //services/...
# Tests use disposable local databases, never shared development data.
bazel test //services/...
bazel build //frontend:webv2_dev_image //frontend:webv2_dev //services/tadoku-api:dev //.dev:seed_image
cd frontend
pnpm install --frozen-lockfile
pnpm --filter webv2 exec tsc --noEmit
pnpm --filter webv2 lint
pnpm build
```

For browser verification on your branch, follow the verification skill in
`.agents/skills/verify-tadoku/SKILL.md`. Changes to routing or synchronization
must also pass the live gates in `.dev/acceptance.md`: frontend HMR without
navigation, backend live edits, separate owners and browser profiles, API-only
frontend fallback, switching and clearing, spoofed headers, a real Navbar login
and a rendered leaderboard. Keep one-off browser verification programs outside
source control.

Merging to `main` can publish new images through the path-filtered CI
workflows. Development cluster access does not authorize publishing or rolling
out a release. dev-cli changes land in the dev-cli repository, not here.
