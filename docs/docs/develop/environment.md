---
title: Development environment
description: How to install dev-cli and run, open, seed, verify and clean up your branch of Tadoku on the shared Talos development cluster.
---

# Development environment

Read this when you want to run your branch of webv2, auth, admin, Tadoku API or its worker
on the shared development cluster, or check that a change works there.

dev-cli deploys live branch overlays of webv2, auth, admin, Tadoku API and its
worker to the `homelab-talos-dev` Kubernetes cluster. Argo CD keeps a shared
base of every service running there even when no developer has a loop running;
see [Development base](../operations/development-base.md). There is no local
Kubernetes cluster or Helm bootstrap to run. For a map of the components, see
[System architecture](../architecture/index.md). Production is deployed from a
private repository and is not covered here.

`.dev/config.yaml` is the committed non-secret configuration and targets
`homelab-talos-dev` directly. Always use that context for development kubectl
commands.

| Setting | Value |
| --- | --- |
| Kubernetes context | `homelab-talos-dev` |
| Hosts | `tadoku.dev.lab`, `account.tadoku.dev.lab`, `admin.tadoku.dev.lab` |
| Overlay image registry | `registry.dev.lab/tadoku-dev-cli` |
| Overlay TTL | 8 hours |

## Prerequisites

- Git access to Tadoku and to the private `antonve/dev-cli` repository.
- Go to install dev-cli, Bazelisk or Bazel at the repository's pinned version,
  and kubectl with authorized access to the `homelab-talos-dev` context.
- Node and pnpm for frontend tools; Docker for local container-based checks.
- Lab network and DNS access, and trust in the Lab CA, including for the
  development registry at `registry.dev.lab`.

The cluster already supplies Postgres, the shared auth providers and routing.
Obtain kube access from the operator. Never commit kubeconfigs, credentials or
private keys.

## Install dev-cli

```sh
GOPRIVATE=github.com/antonve/dev-cli go install github.com/antonve/dev-cli/cmd/dev@v0.7.0
dev version  # must print v0.7.0 or later
```

Go must be able to authenticate to the private repository. If your Git
credentials are SSH-only, rewrite the URL for this one command; no new
credential is required:

```sh
GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=url.git@github.com:.insteadOf \
GIT_CONFIG_VALUE_0=https://github.com/ GOPRIVATE=github.com/antonve/dev-cli \
go install github.com/antonve/dev-cli/cmd/dev@v0.7.0
```

- Put `$(go env GOPATH)/bin`, or your explicit `GOBIN`, on `PATH`. Check
  `command -v dev` and `dev version` so you do not run an older installation.
- To upgrade an existing installation in place, set `GOBIN` to its directory
  on the same command.
- Stop only your own running loops before upgrading.

The minimum version supports lifecycle hooks and resolved configuration
variables. An older CLI must not start these overlays without provisioning their
tenant.

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
dev up --owner alice
```

Keep this terminal running.

- Frontend sources sync into a pnpm-managed Next.js dev server with hot module
  replacement (HMR).
- Go edits rebuild the affected Bazel binary and restart it in the same pod.
  A failed compilation keeps the last working process running.
- A Tadoku API or worker selection starts both workloads against the profile's
  database, scoped to the same branch tenant. Ordinary branches use shared
  `tadoku`; migration branches use an isolated database. The unchanged peer keeps
  its current image; only a changed binary restarts on a live edit.
- dev-cli builds overlay images on demand, pushes them to the development
  registry and deploys them by immutable digest.
- The `tenant` lifecycle hook provisions and seeds the branch before any
  overlay starts, including a frontend-only branch; see
  [Branch tenants](#branch-tenants). No explicit task flags are needed.

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

A frontend-only overlay reads and writes its branch tenant through the base
API. Add `--service tadoku-api` when you need to run changed API code, rather
than to isolate frontend writes. Select the main hostname before any API write,
including one initiated by auth or admin.

Token-reflector is base-only. Paper styleguide has no live overlay; run it
locally with `cd frontend && pnpm paper-styleguide`.

### Make wrappers

| Target | Runs |
| --- | --- |
| `make dev-up` | `dev up` |
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
  uses. Before admin or auth writes to the API, open both that host's link and
  the main-host link in the same browser profile. Otherwise those API calls can
  use canonical base data. Check `X-Dev-Selected` on the actual API response.
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
convergence can serve the base for a while after `dev up` prints a link. A
frontend-only branch intentionally uses the base API with a selected branch
tenant; a base backend header alone does not mean it uses canonical data. For
writes, verify `X-Dev-Selected` and the persisted tenant as well.

### How requests reach your overlay

```text
browser → ingress-nginx → Envoy → webv2 / auth / admin
                             → Oathkeeper → Envoy → Tadoku API
```

- The browser cookie wins over any routing header a client supplies.
- Envoy normalizes the branch selection into an internal header. The
  development Oathkeeper, which uses development-only signing credentials and
  auth providers, signs tenant `tadoku/<route>` from that trusted selection.
  Unselected requests carry `tadoku/prod`.
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

`make dev-seed` runs `scripts/dev/seed-db.sh` against the shared base with
tenant `tadoku/prod`. It:

- runs only against the `homelab-talos-dev` context at `https://omni.lab:8100`
  and waits for Postgres and the base migrations;
- creates the two fixture identities in Kratos, marked as owned by the seed,
  and refuses to touch an existing identity with the same email but no marker;
- resets the fixture passwords to the configured values on every run;
- grants the administrator role in Keto;
- loads the fixtures in `scripts/dev/seed/` into the base `tadoku` database.

The SQL fixtures require a `tenant` psql variable; they set it
transaction-locally and write that tenant explicitly. The canonical tenant keeps
the fixed fixture UUIDs; other tenants derive those UUIDs from their tenant key.
Caller-supplied identity UUIDs remain unchanged.

It is safe to rerun for the same tenant, but it changes shared identities and
base data. See
[Authorization](../architecture/authorization.md#seeding-an-administrator-in-development)
for the role details.

| Account | Email | Role |
| --- | --- | --- |
| Dev Admin | `dev@tadoku.app` | Administrator |
| Dev Reader | `reader@tadoku.app` | Reader |

The development fixture password is `tadoku`. Override the emails and passwords
outside Git with `TADOKU_DEV_ADMIN_EMAIL`, `TADOKU_DEV_ADMIN_PASSWORD`,
`TADOKU_DEV_READER_EMAIL` and `TADOKU_DEV_READER_PASSWORD`.

### Branch tenants

The operator-managed `tdk-dev-data/tadoku-dev-db` is the only Postgres server.
Its `tadoku`, `kratos` and `keto` databases are shared. Ordinary branches use
the base `tadoku` database and a separate test tenant. The tenant key is
`tadoku/<route>`; the route contains the owner, branch and a collision-resistant
hash. Different owners of the same branch get different tenants. Normal
startup creates no database, Postgres cluster, PVC or database dependency Job.

- Before startup, a lifecycle marker records the owner, hooks and resolved
  variables. The `tenant` hook runs a bounded Job in `tdk-dev-data`.
- The Job reads Dev Admin and Dev Reader IDs from the canonical `tadoku/prod`
  fixture user rows in `users`. It fails with a request to run
  `make dev-seed` if they are missing; it never creates or changes identities.
- Provisioning uses `tadoku_owner` to create the test tenant, grants its Keto
  object a parent of `app:tadoku`, and grants Dev Reader direct `testers`
  membership. Dev Admin inherits administrator and access permits.
- Flipt uses environment `test` and namespace `tadoku_<route>`, with display
  name `tadoku/<route>`. Provisioning adds missing resources from this branch's
  `k8s/dev/base/flipt/features.yaml` and preserves existing grants.
- SQL fixtures use the non-owner `tadoku` role so row-level security checks
  every write. The guard requires the route-shaped test tenant, the expected
  role and development database host. It refuses canonical seeding through
  the branch task. Credentials stay Secret references.
- `.dev/config.yaml` resolves `DATABASE` to `tadoku` and `TENANT` to
  `tadoku/${DEV_ROUTE}`. API and worker use `${DEV_VAR_DATABASE}` and their
  `API_BRANCH` / `WORKER_BRANCH` settings use `${DEV_VAR_TENANT}`.
- Provisioning or seeding failure prevents overlay startup. Retrying is safe;
  normal startup and teardown serialize tasks using one Lease per route.

Reseed your branch tenant explicitly with:

```sh
dev task --owner alice tenant
```

This reseeds the selected tenant and repairs missing provider resources; it
does not reset a database or overwrite existing Flipt grants. Never delete a
task Lease to force a retry; first prove that the abandoned holder has stopped.

Kratos accounts remain shared. Do not change shared fixture passwords, profiles
or canonical roles for a branch test. Use the branch's own data, Keto object and
Flipt namespace. This is cooperative development isolation: provider processes
and runtime credentials are shared.

### Migration branches

Schema changes follow the standalone migration release contract in
[Database](../tadoku-api/database.md). Ordinary tenant branches use the base
schema. A changed path under `services/tadoku-api/migrations/`, including an
uncommitted scratch migration, automatically selects profile
`isolated-database`. Startup prints `profile=isolated-database` and resolves
`DATABASE` to `tadoku-<route>` while keeping the same `tadoku/<route>` test tenant.
Its `beforeUp` hooks run `migrate`, then `tenant`; provider lifecycle and worker
ownership hooks are unchanged. Ordinary `dev up` creates no database.

A migration-only change selects no deployable: migration files are not in any
`dev_deployable`'s Bazel dependencies. Add the API explicitly, which also selects
its private worker:

```sh
dev up --owner alice --service tadoku-api
```

Both use the isolated database and the branch tenant. The bounded migration
task runs in `tdk-dev-data` beside its owner Secret. Its database name is the
literal `tadoku-${DEV_ROUTE}`, independent of configuration variables, so it
cannot migrate shared `tadoku`. Database creation requires the route-shaped
name, length limit, owner `tadoku_owner` and matching dev-cli marker, and
serializes with an advisory lock. It applies the base's runtime DML and sequence
default privileges. Migrations use `tadoku_owner`; API, worker and fixtures use
non-owner `tadoku`, with RLS and no DDL or migration-table writes.

A retained database owned by legacy role `tadoku`, or without the expected
marker, fails closed. Inspect it using
[Development base data safety](../operations/development-base.md#data-safety)
and obtain exact-name removal approval; do not transfer ownership, reset it or
drop it to make startup pass. For a dirty branch migration, follow
[Migration recovery](../operations/migration-recovery.md#development-migration-branches).

After the migration merges and deploys to the shared base, rebase onto updated
`main`, run `dev down --owner alice`, then `dev up --owner alice`. A live route
cannot switch profile or resolved variables: rerunning up after removing the
scratch migration or rebasing refuses with `run dev down first`, before new
hooks or overlays. Down uses the recorded isolated database even if current
changes now select `default`. It removes that tenant and provider state while
retaining the marked branch database. The next ordinary up returns to `tadoku`.
Database deletion always requires separate exact-name approval. Migration
overlays are development-only and never deploy to production.

### Leaderboards on shared Valkey

Leaderboard keys follow the parsed request or persisted job tenant:
`tenant:<name>/<id>:leaderboard:…`. A missing tenant is an error before Valkey
access. The base pair uses `tenant:tadoku/prod:leaderboard:…`.

The API persists `tadoku/<route>` on branch jobs, and caches use
`tenant:tadoku/<route>:leaderboard:…` without a separate branch prefix.

- A frontend-only branch uses the base API and base worker, both operating on
  its selected tenant. The base worker claims all tenants without a matching
  component override.
- Selecting a branch API or worker starts both. Before the private worker
  starts, `worker-override-set` inserts its tenant's `tadoku-worker` override;
  the branch worker claims only that tenant and invalidates its caches.
- Teardown waits for the branch worker and its pods to stop, then runs
  `worker-override-clear`. The base can resume remaining branch jobs before
  tenant teardown. Jobs already claimed by the base before an override may
  finish there; compare job tenant and worker evidence when verifying ownership.
- Job producers and consumers share versioned contracts. Follow
  [Jobs and worker](../tadoku-api/jobs.md#migrate-v1-to-v2) before changing a
  payload; pairing workloads does not make incompatible versions safe.

## Clean up

```sh
dev down --owner alice
```

`dev down` stops the local loop and removes only the current owner and branch's
overlays. It waits for their pods to stop, clears worker overrides, then runs
`tenant-teardown`: Keto object tuples, tenant Valkey keys, the test Flipt
namespace, and finally the PostgreSQL tenant row with its cascading data.
Disposable Jobs and the lifecycle marker are removed after successful teardown.
It does not stop the Argo CD base. Ctrl-C alone leaves overlays running.

- `dev cleanup` and startup clean up expired environments across owners,
  replaying their recorded hooks and variables. `dev status` does not run
  lifecycle hooks; it reports marked expired routes as `teardownPending`.
- If teardown fails, its marker and tenant row remain for diagnosis and retry.
  Rerun `dev down`, or `dev cleanup` for expired routes. A missing recorded task
  causes cleanup to skip that route; use compatible configuration rather than
  deleting its marker. Do not force removal of a failed teardown's marker.
- Ordinary branches leave no database to drop. **Migration branch databases
  and previously retained databases remain retained** after down and TTL cleanup.
  Deleting one requires
  inspecting its exact name, marker, owner and active connections, then separate
  authorization.
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
bazel build //frontend:webv2_dev_image //frontend:webv2_dev //services/tadoku-api:dev //services/tadoku-api:worker_dev //.dev:tenant_image
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
