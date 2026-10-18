---
title: Development environment
description: How to install dev-cli and run, open, verify and clean up your branch of Tadoku on the shared Talos development cluster.
---

# Development environment

Read this when you want to run your branch of webv2, webv3, auth, admin, Tadoku API or its worker
on the shared development cluster, or check that a change works there.

dev-cli deploys live branch overlays of webv2, webv3, auth, admin, Tadoku API and its
worker to the `homelab-talos-dev` Kubernetes cluster. Argo CD keeps a shared
base of every service running there even when no developer has a loop running;
see [Development base](../operations/development-base.md). There is no local
Kubernetes cluster or Helm bootstrap to run. For a map of the components, see
[System architecture](../architecture/index.md). Production is deployed from a
private repository; running a branch beside production is covered in
[Production branches](./production-branches.md).
[Development data](./development-data.md) covers seed data, branch tenants,
migration branches and leaderboard caches.

`.dev/config.yaml` is the committed non-secret configuration and targets
`homelab-talos-dev` directly. Always use that context for development kubectl
commands.

| Setting | Value |
| --- | --- |
| Kubernetes context | `homelab-talos-dev` |
| Hosts | `tadoku.dev.lab`, `next.tadoku.dev.lab`, `account.tadoku.dev.lab`, `admin.tadoku.dev.lab` |
| Branch hosts | `<route>.tadoku.dev.lab`, `<route>.next.tadoku.dev.lab`, `<route>.account.tadoku.dev.lab`, `<route>.admin.tadoku.dev.lab` |
| Certificate issuer | `lab-ca-acme` |
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
GOPRIVATE=github.com/antonve/dev-cli go install github.com/antonve/dev-cli/cmd/dev@v0.9.0
dev version  # must print v0.9.0 or later
```

Go must be able to authenticate to the private repository. If your Git
credentials are SSH-only, rewrite the URL for this one command; no new
credential is required:

```sh
GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=url.git@github.com:.insteadOf \
GIT_CONFIG_VALUE_0=https://github.com/ GOPRIVATE=github.com/antonve/dev-cli \
go install github.com/antonve/dev-cli/cmd/dev@v0.9.0
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
[Seed data](./development-data.md#seed-data)).

```sh
dev up --owner alice
```

Keep this terminal running.

- Frontend sources sync into pnpm-managed dev servers with hot module
  replacement (HMR). Webv2, auth and admin use Next.js; Webv3 uses Node22 and
  Vite, with a `paper-ui` build watcher.
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
  [Branch tenants](./development-data.md#branch-tenants). No explicit task
  flags are needed.

**Owner** is a stable label for a developer or worktree, not authentication.
Use a different owner for each checkout you run at the same time.

### Which services start

dev-cli asks Bazel which deployables changed relative to the merge base with
`origin/main`, including uncommitted edits. It starts overlays only for those;
every other service keeps using the base.

- Auth-only or admin-only edits select only that app. Shared `packages/ui`
  edits select the three legacy frontends. Webv3-only or shared `packages/paper-ui`
  edits select Webv3.
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
than to isolate frontend writes. Branch-host links select the same route across
apps, including API writes initiated by auth or admin.

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
dev url --owner alice --host next.tadoku.dev.lab '/'
dev url --owner alice --host account.tadoku.dev.lab '/login'
dev url --owner alice --host admin.tadoku.dev.lab '/'
dev logs --owner alice tadoku-api
dev logs --owner alice tadoku-worker
```

Flags come before positional paths and service names.

Open the printed branch-host link, including for deep links. One link keeps
navigation across app, account and admin on the same branch. Each branch has
its own origin, so two owners can use one browser profile without selection
leaking across tabs. A printed link does not create an overlay or prove readiness.

- Base URLs are `https://tadoku.dev.lab`, `https://next.tadoku.dev.lab`, `https://account.tadoku.dev.lab` and
  `https://admin.tadoku.dev.lab`.
- Each service independently uses its healthy overlay or the base fallback.
- Authentication is shared between base and branches. Kratos remains at
  `https://account.tadoku.dev.lab/kratos`; it is never routed to a frontend overlay.

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
browser → ingress-nginx → Envoy → webv2 / webv3 / auth / admin
                             → Oathkeeper → Envoy → Tadoku API
```

- An owner-scoped Ingress and Lab CA certificate in `tdk-dev-routing` expose
  the branch hosts. The hostname wins over any routing header a client supplies.
- Envoy normalizes the branch selection into an internal header. The
  development Oathkeeper, which uses development-only signing credentials and
  auth providers, signs tenant `tadoku/<route>` from that trusted selection.
  Unselected requests carry `tadoku/prod`.
- Base and branch frontends send server-side rendering requests through the
  same public branch host and gateway with Lab CA trust, so an API-only
  selection also applies to server-rendered pages.
- Kratos and Oathkeeper allow the development hosts and every
  `*.tadoku.dev.lab` branch host; no temporary auth allowlist is needed.

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

- For TLS issuance failures, inspect `kubectl --context homelab-talos-dev -n tdk-dev-routing describe certificate <name>`. Certificate readiness is required; fix issuance before retrying `dev up`.
- A branch hostname returning 404 means its environment is not running; check owner, route and status.

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
frontend fallback, branch-host isolation, spoofed headers, a real Navbar login
and a rendered leaderboard. Keep one-off browser verification programs outside
source control.

Merging to `main` can publish new images through the path-filtered CI
workflows. Development cluster access does not authorize publishing or rolling
out a release. dev-cli changes land in the dev-cli repository, not here.
