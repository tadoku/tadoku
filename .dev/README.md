# Tadoku development with DevCLI

Develop **webv2 (Next.js/pnpm)** and **native tadoku-api** at
**https://tadoku.dev.lab**. The committed `.dev/config.yaml` is the real Homelab
configuration; no hostname substitution is needed.

## Start working

Install a CLI revision containing YAML support (merged in antonve/dev-cli#16):

```sh
GOPRIVATE=github.com/antonve/dev-cli go install github.com/antonve/dev-cli/cmd/dev@d9f9aed7f2c381d772db366217afca2627889cbf
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

In another terminal, in the same checkout:

```sh
dev status --owner alice
dev url --owner alice '/'
dev logs --owner alice tadoku-api
dev down --owner alice
```

Open the printed link: Envoy sets a host-only branch cookie, also for deep links.
No application branch menu is needed. Separate browser profiles have independent
selections; tabs in one profile share it. Clear with
`dev url --owner alice --clear '/'`. Owner is a stable developer/worktree label,
not authentication; use different owners for simultaneous checkouts.

`make dev-up`, `make dev-logs`, and `make dev-down` wrap these commands.
Set `DEV_OWNER` consistently. Frontend-only work can omit the database tasks:
`dev up --owner alice --service webv2` when only the frontend is affected.
`--service` adds to, rather than filters, affected services.

Bazel's graph selects affected services once at startup relative to the merge
base with `origin/main`. Uncommitted edits count; unknown paths conservatively
select all deployables. Use `--base <ref>` to override the comparison. Restart
the loop to add services. `--no-watch` does not provide live updates.

## One Postgres pod, branch databases

The existing operator-managed `default/tadoku-dev-db` is the only required
Postgres cluster. Base `immersion`, `kratos`, and `keto` databases stay shared.
Each native overlay uses `immersion-<owner-and-branch-route>` on that same server.
The route contains a collision-resistant hash. No branch creates another
Postgres cluster, PVC, or long-running database pod.

A short-lived dependency Job creates the database idempotently; only that Job
references the existing administrator Secret. API/migration/seed use the existing
`immersion` role. Credentials remain Secret references. This is cooperative
development isolation, not a hostile-tenant boundary: the role, Kratos, Keto,
Valkey and unchanged legacy APIs remain shared.

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
ownership, then separate authorization. Never drop shared `immersion`,
`kratos`, or `keto`. `make dev-reset` is the old destructive whole-stack reset,
not part of this workflow.

## Shared setup and routing

The existing base workloads, Postgres operator, Secrets and Envoy Gateway must
already exist. `k8s/dev/dev-cli/` configures their canonical development
entrypoint, reusing base frontend/API Services and the existing TLS Secret.

```text
browser → ingress-nginx → Envoy → webv2
                             → Oathkeeper → Envoy → native API
```

The development Oathkeeper references existing signing credentials and auth
providers. It propagates Envoy's normalized branch header internally. Browser
cookie selection wins over supplied routing headers. Each service independently
selects its healthy overlay or base. The canonical hostname is already allowed
by Kratos, so no temporary auth allowlist is required.

Operators stage backends before handing over the existing canonical Ingresses:

```sh
kubectl --context homelab-dev apply --dry-run=server -k k8s/dev/dev-cli
kubectl --context homelab-dev apply -f k8s/dev/dev-cli/base.yaml
kubectl --context homelab-dev -n default rollout status deployment/tadoku-cli-oathkeeper
kubectl --context homelab-dev -n tadoku-dev-cli get httproute
# Require current-generation Accepted/ResolvedRefs and test gateway traffic.
kubectl --context homelab-dev apply -f k8s/dev/dev-cli/ingress.yaml
kubectl --context homelab-dev -n default patch deployment frontend-webv2 --type=strategic --patch-file k8s/dev/dev-cli/frontend-base-patch.yaml
kubectl --context homelab-dev -n default rollout status deployment/frontend-webv2
```

Do not run Tilt concurrently: its old Ingress definitions can overwrite these
entrypoints and base environment. The versioned frontend patch preserves its
published image while sending SSR through the same gateway with Lab CA trust;
without it, API-only selections would use base API data during SSR.
Legacy Tilt files still describe the existing base; full retirement
and Argo ownership are separate work. Production remains unchanged.
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

Tadoku integration remains one PR; reusable CLI changes land separately.
Merging Tadoku may trigger its existing production image publication workflow.
Development verification does not authorize that publication or rollout.
