# Tadoku development with DevCLI

Develop **webv2, auth, admin (Next.js/pnpm)** and **native tadoku-api** on the
Homelab development cluster. The committed `.dev/config.yaml` is the real Homelab
configuration; no hostname substitution is needed.

The fresh `tdk-dev-*` GitOps base is active on `homelab-dev`. The approved cutover
retired the old Tilt namespaces and disposable databases on 2026-09-21. See
[`k8s/dev/base/README.md`](../k8s/dev/base/README.md) before first use.

## Start working

Install [DevCLI v0.4.0](https://github.com/antonve/dev-cli/releases/tag/v0.4.0)
or newer. This release includes YAML configuration, multi-host routing,
dependency/tasks, cold-start waiting and non-blocking TTL heartbeats:

```sh
GOPRIVATE=github.com/antonve/dev-cli go install github.com/antonve/dev-cli/cmd/dev@latest
dev version  # v0.4.0 or newer
dev doctor
make dev-seed  # shared synthetic identities and base fixtures; safe to rerun
# Make service edits, then:
dev up --owner alice --task migrate --task seed
```

Keep that terminal running. Migration/seed tasks prepare your branch database
before API startup. Frontend edits sync into the pnpm Next.js dev server for HMR;
Go edits rebuild the selected binary and restart it in the same pod. Compilation
failure keeps the last working process running. The CLI builds/pushes images to
the configured development registry and injects immutable digests.

Go must be able to authenticate Git access to the private CLI repository. If your
existing credentials use SSH only, add per-command Git URL rewriting to the
installation command (no new credential is required):

```sh
GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=url.git@github.com:.insteadOf \
GIT_CONFIG_VALUE_0=https://github.com/ GOPRIVATE=github.com/antonve/dev-cli \
go install github.com/antonve/dev-cli/cmd/dev@latest
```

Put `$(go env GOPATH)/bin` (or your explicit `GOBIN`) on `PATH`. Check
`command -v dev` and `dev version` to avoid running an older installation.
On T3 the current user-local installation is `/home/t3/.local/bin/dev`; it is
not bundled in the Homelab image. Upgrade that installation with
`GOBIN=/home/t3/.local/bin` on the same command. For a reproducible pin, replace
`@latest` with `@v0.4.0`. Stop only your own existing loops before upgrading.

In another terminal, in the same checkout:

```sh
dev status --owner alice
dev url --owner alice '/'
dev url --owner alice --host account.tadoku.dev.lab '/login'
dev url --owner alice --host admin.tadoku.dev.lab '/'
dev logs --owner alice tadoku-api
dev down --owner alice
```

Open the printed link: Envoy sets a host-only branch cookie, also for deep links.
No application branch menu is needed. Separate browser profiles have independent
selections; tabs in one profile share it. Clear with
`dev url --owner alice --clear '/'`. Owner is a stable developer/worktree label,
not authentication; use different owners for simultaneous checkouts.

Selection is **per hostname**. Open the printed account/admin link to select
that frontend; when testing an API overlay from admin, also open the main-host
link in the same profile. Clearing one hostname does not clear the others.
Kratos stays shared and `/kratos` is never routed to a frontend overlay.

Auth-only or admin-only source edits select only that app through Bazel. Shared
`packages/ui` edits select all three frontends. Token-reflector is deliberately
base-only; Paper styleguide live overlays are deferred (its local pnpm command
still works). See [acceptance.md](acceptance.md) for the live verification gates.

`make dev-up`, `make dev-logs`, and `make dev-down` wrap these commands.
Set `DEV_OWNER` consistently. Frontend-only work can omit the database tasks:
`dev up --owner alice --service webv2` when only the frontend is affected.
`--service` adds to, rather than filters, affected services.

Bazel's graph selects affected services once at startup relative to the merge
base with `origin/main`. Uncommitted edits count; unknown paths conservatively
select all deployables. Use `--base <ref>` to override the comparison. Restart
the loop to add services. `--no-watch` does not provide live updates.

## One Postgres pod, branch databases

The operator-managed `tdk-dev-data/tadoku-dev-db` is the only required
Postgres cluster. Base `tadoku`, `kratos`, and `keto` databases stay shared.
Each native overlay uses `tadoku-<owner-and-branch-route>` on that same server.
The route contains a collision-resistant hash. No branch creates another
Postgres cluster, PVC, or long-running database pod.

A short-lived dependency Job creates the database idempotently; only that Job
references the existing administrator Secret. API/migration/seed use the existing
`tadoku` role. Credentials remain Secret references. This is cooperative
development isolation, not a hostile-tenant boundary: the role, Kratos, Keto,
and Valkey remain shared.

Each branch API runs its own leaderboard outbox worker against its branch
database. Its `dev:${DEV_ROUTE}:` cache prefix keeps every leaderboard key and
startup scan separate from the base and other branches on shared Valkey. The
base API uses unprefixed keys. Keep this prefix unique when changing overlay
routing; disabling a branch worker while keeping cache reads active leaves its
leaderboards stale and its outbox pending.

```sh
dev task --owner alice migrate
dev task --owner alice seed
```

These explicit tasks publish their Bazel images and serialize against the exact
branch database target. They do not reset databases. Never delete a task Lease
to force a retry; first prove the abandoned holder has stopped.

`make dev-seed` uses the existing shared Kratos/Keto seeder for administrator
`dev@tadoku.app` and reader `reader@tadoku.app`; default fixture password
`tadoku`, overrides outside Git. It refuses to take over identities without its
seed ownership marker. Branch fixtures reuse those identities, not new providers.
Base fallback remains independent: deploy related services too when a test needs
consistent data across services otherwise using base data.

`dev down` removes only current-owner/branch overlays and disposable Jobs.
**Branch databases are retained**, also after overlay TTL cleanup. No automatic
SQL-drop operation exists. Deletion requires inspecting the exact database and
ownership, then separate authorization. Never drop shared `tadoku`, `kratos`,
or `keto`.
`make dev-reset` is disabled; any reset needs an explicitly
approved, scoped runbook, not the historical Tilt reset script. The old
`immersion` databases and role were retired from the shared development
cluster under a separate approved runbook.

## Shared setup and routing

Argo CD reconciles the development-only base from `k8s/dev/base/`, including
automatic migration Jobs, base workloads and canonical routing. The platform
Postgres operator and Envoy Gateway must already exist. Homelab owns the Argo
Application and development Image Updater; Tadoku's existing CI publishes the
GHCR base images. Development Image Updater follows `latest` by digest.

```text
browser → ingress-nginx → Envoy → webv2
                             → Oathkeeper → Envoy → native API
```

The development Oathkeeper references development-only signing credentials and auth
providers. It propagates Envoy's normalized branch header internally. Browser
cookie selection wins over supplied routing headers. Each service independently
selects its healthy overlay or base. The canonical hostname is already allowed
by Kratos, so no temporary auth allowlist is required.

Operators must follow the fresh-base bootstrap and explicit cutover gates in
the base README. Full Argo syncs run migrations before dependent workloads;
selective resource sync skips hooks and must not be used for releases.
Do not run Tilt concurrently or apply the historical `k8s/dev/dev-cli/` pilot
manifests. Both can conflict with the new base's canonical routes. Base and branch
frontends send SSR through the same gateway with Lab CA trust, preserving API-only
branch selection. Old resources and data must not be deleted without explicit
approval. Production remains unchanged.
Only the public CA and Secret references are committed, never private keys,
kubeconfigs, tokens or plaintext Secrets.

For overrides use an ignored file with `dev <command> --config <path>`.
Review and move an old ignored `.dev/config.json`: the CLI rejects ambiguous
JSON/YAML defaults rather than silently choosing one.

## Verification and troubleshooting

Doctor checks prerequisites, not end-to-end routing. Inspect status, task logs,
route admission and real responses. Selected routes expose intent in
`X-Dev-Selected`, actual upstream in `X-Dev-Backend`, and outer Oathkeeper in
`X-Dev-Proxy-Backend`. Unselected static base routes need not emit these headers.

Verify frontend HMR without navigation, backend live edits, separate owners and
browser profiles, API-only frontend fallback, switch/clear, spoofed headers,
real Navbar login and rendered leaderboard. Removing one overlay must preserve
the other owner and base. Routing convergence can briefly serve base.
One-off browser verification programs belong outside source control.
Cold-start DNS/health convergence can initially serve base even after `dev up`
prints a link. Check `X-Dev-Backend` before testing overlay behavior. Successful
backend restarts can briefly return 503 while health-based fallback catches up;
this is not zero-downtime deployment. Compile failures leave the old process up.
With `dev up` running, replacement pods receive current source again; without
the loop they start from their image, not the lost writable container layer.

Tilt decommissioning is held for explicit owner approval. Keep its historical
files untouched until that gate; do not run Tilt alongside the GitOps base.

```sh
bazel mod deps --lockfile_mode=error
bazel run //:gazelle -- -mode=diff
bazel build //services/...
# Tests use disposable LOCAL databases, never shared development data.
bazel test //services/...
bazel build //frontend:webv2_dev_image //frontend:webv2_dev //services/tadoku-api:dev //.dev:seed_image
cd frontend
pnpm install --frozen-lockfile
pnpm --filter webv2 exec tsc --noEmit
pnpm --filter webv2 lint
pnpm build
```

The original integration landed in Tadoku #981; live startup corrections landed
in #1007 and #1008. Reusable CLI changes land separately. Merging Tadoku may
trigger its existing production image publication workflow; development cluster
access alone does not authorize that publication or rollout.
