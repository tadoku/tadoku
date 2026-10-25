---
title: Development base
description: How Argo CD runs the always-on Tadoku development base on homelab-talos-dev, and how operators bootstrap, migrate, verify and recover it.
---

# Development base

Read this when you operate the Argo CD development base in `k8s/dev/base/`:
first activation, credentials, migrations, image updates, worker ownership or
recovery.

`k8s/dev/base/` is exclusively for the `homelab-talos-dev` cluster.
**Never apply it to another Kubernetes context.** Production is deployed from a
private repository and is not affected by these manifests. To run and verify
your own branch, use [Development environment](../develop/environment.md)
instead; for a map of the components, see
[System architecture](../architecture/index.md).

## Ownership

| Owner | Owns |
| --- | --- |
| Argo CD | The always-running base defined in `k8s/dev/base/`, including automatic migration Jobs, base workloads and canonical routing |
| dev-cli | Branch overlays, routing overrides, lifecycle markers and short-lived tenant provisioning, worker-override and teardown tasks |
| Homelab infrastructure | The Argo CD Application and the development Image Updater, not copies of these workload manifests |
| Platform | The Postgres operator and Envoy Gateway, which must exist before the base syncs |

The `tadoku_owner` role owns the base and branch databases and their public
tables. Migrations use that owner role. API, worker and seed processes use
`tadoku`, a non-owner without `bypassrls`, with table DML and sequence grants.
The runtime role cannot create tables or write `schema_migrations`.

## Topology

Components run in `tdk-dev-*` namespaces; authv2 runs in the account namespace, `tdk-dev-frontend-auth`:

| Namespace | Contents |
| --- | --- |
| `tdk-dev-frontend-webv2`, `tdk-dev-frontend-webv3`, `tdk-dev-frontend-auth`, `tdk-dev-frontend-admin` | Webv2, Webv3, authv2 and admin frontends |
| `tdk-dev-tadoku-api` | Tadoku API and the private asynchronous worker |
| `tdk-dev-keto`, `tdk-dev-oathkeeper` | Auth providers |
| `tdk-dev-flipt` | Feature flags |
| `tdk-dev-token-reflector` | Token-reflector |
| `tdk-dev-data` | One operator-managed Postgres server (`tadoku-dev-db`) and disposable Valkey |
| `tdk-dev-routing` | Application routes attached to the platform Envoy Gateway |

The Postgres server holds the base `tadoku` and `keto` databases and
retained `tadoku-<route>` migration branch databases. Ordinary branches use
their own test tenant, `tadoku/<route>`, in `tadoku`; canonical base data uses
`tadoku/prod`. Row-level security and transaction-local tenant context isolate
their application rows. No ordinary branch database is created. Changed
`services/tadoku-api/migrations/` paths automatically select the
`isolated-database` profile: a bounded creation Job establishes the owner and
default runtime privileges, then migration and tenant Jobs run in `tdk-dev-data`.
The paired branch API/worker use that database and the same test tenant; the
base worker does not consume jobs from a branch database. See
[Migration branches](../develop/development-data.md#migration-branches). The server is
provisioned with the Zalando `postgresql` custom resource; do not add hand-rolled Postgres Deployments or Helm
releases. Styleguides and optional admin tools are not deployed. The base holds no production data, external backups or
notification configuration.

## Images and Image Updater

- The repository's existing CI publishes the GHCR runtime and migration images.
  Do not add another build or push pipeline.
- Backend images, Webv3 and authv2 select `latest` in
  `k8s/dev/base/kustomization.yaml`; Webv2 and admin select reviewed
  immutable commit tags. The development Image Updater uses the **digest**
  strategy and writes the immutable resolutions back to that file.
- The `image-list` annotation of the homelab repository's `tadoku-dev` Argo CD
  Application (clusters/talos/dev/root/tadoku-dev.yaml) owns the Webv2 and
  admin tag selections. Change them there; Image Updater reverts a
  selection edited only in the Kustomization.
- The worker has its own GHCR image and Kustomize image entry. Image Updater
  tracks its digest independently of the API image.
- Webv3 tracks its published `latest` digest independently at
  `https://next.tadoku.dev.lab`. Its deployment keeps the canonical Webv2 host
  available for product journeys linked from the preview.
- Authv2 follows [merge → publish → public pulls → base readiness → Image Updater](../frontend/authv2.md#images-and-deployment).
  Delayed public visibility can block its Pod and later sync waves.
- Hook migration images need Image Updater's `force-update`, because successful
  Jobs are removed from the live resource list.
- Kustomize mirrors the migration image digest into a top-level annotation on
  the Tadoku API Deployment, so a migration-only update triggers a full Argo CD
  sync without changing the API Pod template.
- Image Updater write-back commits touch only the Kustomization, so they do not
  match the path filters of the CI image-publication workflows.

## Frontend start adapter

The CI-published Next.js standalone images freeze their build-time URLs in
`server.js`. `k8s/dev/base/frontend-start.cjs` keeps their compiled assets and
build configuration unchanged, but supplies development-only runtime endpoints
before starting Next. The browser receives the public
`https://tadoku.dev.lab/api/internal` endpoint. Server-side requests use
`http://oathkeeper-proxy.tdk-dev-oathkeeper:4455/api/internal`, the private
Oathkeeper entry for session and feature-flag reads, and forward the incoming
X-Dev-Branch. The public Lab CA is mounted for server-side HTTPS.

- It is a deployment adapter, not a source-sync server or reusable CLI runtime.
- It calls a private Next startup API. Reverify it whenever the frontend
  framework version changes.
- Branch pods use the dev-cli-managed pnpm Next.js dev server, not this adapter.

Webv3 runs its published Node 22 Nitro server directly, with a non-root identity
and read-only root filesystem. Its `/healthz` readiness and liveness probes do
not depend on API availability. The stable `tadoku-cli-webv3` ConfigMap supplies
`TADOKU_WEB_URL=https://tadoku.dev.lab` and
`TADOKU_API_URL=https://tadoku.dev.lab/api/internal`; `tadoku-cli-ca` mounts the
public Lab CA at `/etc/dev-ca/ca.crt` for server-side HTTPS. Branch pods reuse
those ConfigMaps and run the source-synced Vite server. Per-request hostname
resolution keeps Webv3's product links on its matching Webv2 branch host.

Canonical Webv3 requests use the shared Envoy Gateway through a Lab CA
ingress. `/api/internal` uses Oathkeeper before the API, while page and asset
requests use Webv3. Existing Tadoku subdomain admission and shared account
cookies cover the preview and branch hosts.

Authv2 runs in `tdk-dev-frontend-auth` and uses
`tadoku-cli-authv2`; overlays use inline URLs and `tadoku-cli-ca`. Canonical
routes strip branch headers. Anonymous public identity rules sign a guest JWT
and forward Cookie to API auth/CSRF; other routes keep session checks.
See [Authv2](../frontend/authv2.md).

## Identity session callback

[Identity gateway configuration](identity-gateway.md) covers credentials,
logging, private SSR and activation.

## Asynchronous worker ownership

The private `tadoku-worker` Deployment consumes `jobs` in the base
`tadoku` database across tenants without a matching `tadoku-worker` override.
It derives leaderboard keys from each persisted job tenant. It has one
replica, Recreate rollout, a separate image digest, and no Service or public
route. Its resource requests and limits are in
`k8s/dev/base/services/tadoku-worker.yaml`.

[Operate jobs and the worker](../tadoku-api/jobs-operations.md#leaderboard-freshness-and-worker-health)
describes its readiness, metrics, failures and replay.

dev-cli pairs each branch API and worker against the profile's database with
`API_BRANCH` and `WORKER_BRANCH` set to `tadoku/<route>`. Selecting either
workload starts both; unchanged peers use their current image. The
`worker-override-set` hook runs before its Deployment starts. After the worker
and its pods stop, `worker-override-clear` removes that override so the base
can resume remaining jobs. A frontend-only branch needs no worker overlay:
the base API and worker serve its selected test tenant. Both paths use
`tenant:tadoku/<route>:leaderboard:…` keys. Verify branch writes, worker
completion, and leaderboard reads without changing canonical or another tenant.
The offline E2E (see [Verification](#verification)) checks rendered ownership,
isolation, image separation and lack of worker routing. Tadoku API's backend
E2Es check tenant-scoped PostgreSQL work with shared Valkey.

## Development email

The base sends no email: base and branch workers bind
`WORKER_EMAIL_ENABLED=false`, overriding the `ENABLED` fallback. Manifests alone
do not prove a cluster sends no email; that needs restarted worker Pods with
the binding.

## Automatic migrations

Full Argo CD syncs apply these waves:

| Wave | Resources and gate |
| --- | --- |
| -30 | Development namespaces |
| -20 | Postgres custom resource and generated configuration |
| -15 | Transfer existing Tadoku objects to `tadoku_owner` and establish runtime grants |
| -10 | Tadoku API migrations as `tadoku_owner` and the Keto migration Sync hook; each waits for authenticated database connectivity |
| -5 | Reapply runtime grants and revoke writes to the newly created `schema_migrations` table |
| 0 | Auth providers, cache, Flipt, token-reflector and Gateway routes |
| 10 | Oathkeeper, which publishes its JWKS before the APIs start |
| 20 | Tadoku API and its private worker |
| 30 | Frontends |
| 50 | Browser Ingresses |

Migration Jobs have a bounded deadline, `backoffLimit: 0` and the
`BeforeHookCreation,HookSucceeded` delete policy.

- A failed Job remains for diagnosis and prevents the later application waves
  from starting. Existing healthy application replicas are not deleted.
- No-op reruns succeed. Subsequent full syncs need no manual migration step.
- A failure does not authorize a database reset, schema downgrade or rollback.
  Fix forward and run another full sync. If a Tadoku API migration left
  `schema_migrations` dirty, follow
  [Database migration recovery](./migration-recovery.md).
- **Never use selective resource sync to release an application**: it skips
  Argo CD hooks. Do not configure `ApplyOutOfSyncOnly=true`.

Follow the migration-first release contract: schema changes land and deploy on
their own before dependent runtime code. Digest tracking does not replace
backward-compatible schema and application releases.

Base seeding is separate from migrations. On `homelab-talos-dev`, `make dev-seed`
seeds tenant `tadoku/prod` with base fixtures and, through the canonical
identity provider, synthetic fixture accounts. SQL seeds require the supplied
tenant, set it per transaction and write it explicitly. The canonical tenant
retains the fixture UUIDs; other tenants derive them from their tenant key.
Caller-supplied identity UUIDs are unchanged.

Branch startup's `tenant` hook provisions and seeds its test tenant; see
[Branch tenants](../develop/development-data.md#branch-tenants).
The Job creates native fixture accounts with the canonical fixture UUIDs, grants
Dev Admin `admins` in the branch tenant and provisions missing Flipt resources
in `test/tadoku_<route>`. It runs in `tdk-dev-data`; Flipt's ingress policy
admits only labeled tenant-lifecycle pods from that namespace for these direct
management calls.

## Credentials

[Development credentials](development-credentials.md) covers generated database
roles, Secret bootstrap, provider keys and rotation. Bootstrap requires explicit
approval; keep private keys and Secret contents out of Git.

## Activation and bootstrap

Provisioning creates fresh development credentials and auth identities; it
imports no production data or credentials.

1. Obtain approval for provisioning.
2. Inventory the canonical hosts and confirm there are no competing Ingresses
   or HTTPRoutes. A conflicting route can keep serving another workload even
   when these resources are healthy.
3. Make sure `k8s/dev/base/` is on `main`, configure development repository
   access for the Argo CD Application, then start a **full** sync.
4. Once the namespaces and operator-generated Secrets exist, run the bootstrap.
   The initial sync can wait at the auth-provider wave until it supplies the
   runtime Secrets.
5. Once the migration Jobs and base workloads are Healthy, seed.

```sh
bash scripts/dev/bootstrap-gitops-secrets.sh
# Wait for migration Jobs and base workloads to become Healthy, then:
make dev-seed
```

After activation, run the live acceptance gates listed under
[Verification](#verification).

## Data safety

- The Postgres custom resource and the `tdk-dev-data` namespace have
  `Prune=false,Delete=false`. Ordinary Argo CD pruning or removing the
  Application must not destroy data.
- `dev down` and TTL cleanup remove the selected test tenant's Valkey keys and
  Flipt namespace before deleting its PostgreSQL tenant row and cascading
  application rows, including its permission relationships. They stop branch pods and clear their worker
  overrides first. Canonical `tadoku/prod` is refused by the lifecycle tool.
- Teardown failures retain the lifecycle marker for retry. The tenant row is
  deleted last so an external-provider failure preserves an inventory of the
  remaining tenant. Do not delete its marker or task Lease to bypass a failure.
- Migration branch databases remain retained on down and TTL cleanup, along
  with previously retained databases. Teardown uses the lifecycle marker's
  recorded database even after a rebase changes the selected profile. Startup
  reuses only databases owned by `tadoku_owner` with the matching marker;
  legacy `tadoku` ownership fails closed without an implicit ownership transfer.
  Deleting a retained database requires an exact-name inventory and explicit approval:
  require the `tadoku-<route>` name, matching
  `dev-cli branch database route=<route>` database marker, owner `tadoku` or
  `tadoku_owner`, no active connections, and no live
  API/worker route or lifecycle marker using that database. Report every
  mismatch and leave it alone. Never use a forced database drop.
- Database or namespace deletion is never an implicit setup or cleanup step.
  Inspect the exact resources and obtain explicit approval for any destructive
  operation. Never run namespace-wide deletion.
- `make dev-reset` is disabled. Any reset needs an explicitly approved, scoped
  runbook.
- `infra/dev/ory/` holds the Keto namespace fixture that Bazel backend tests
  load; it is not an obsolete deployment file. Preserve the shared
  SQL fixtures in `scripts/dev/seed/`, which base and branch seeding both use.
- Any future rebuild of the environment needs its own data-disposition
  decision. A previous deletion approval is not permission to erase later data.

## Verification

Offline checks:

```sh
kubectl --context homelab-talos-dev kustomize k8s/dev/base |
  kubeconform -strict -summary -ignore-missing-schemas
node k8s/dev/base/e2e.cjs /tmp/tadoku-base-evidence
```

Install the frontend's locked pnpm dependencies first; the E2E harness reuses
its YAML parser. The harness runs the real published migration images in
isolated, resource-bounded Docker containers, not a nested Kubernetes cluster.
The `verify-dev-base` workflow runs the same harness for changes to development
manifests and migration inputs, retaining `development-base-evidence` even on
failure. It needs no cluster credentials.

- It records the source revision, worktree status, rendered-manifest hash,
  exact image digests, commands and results in `report.json`, with individual
  logs.
- It proves fresh and existing ownership transfer, runtime DML and sequence
  access, denied DDL and migration-table writes, failure of the ownership hook
  when its validation fails, and branch provisioning. It
  checks the automatic migration profile, declared variables and hooks, literal
  branch-only migration targets, and refusal to provision legacy-owned databases.
- It also proves fresh migrations, no-op reruns, an intentional connection
  failure and recovery, plus HTTP readiness, development runtime URLs and
  compiled assets in the Webv2 and admin Next.js images and Webv3 with external
  networking disabled. For base and branch hosts, each Next.js image's
  server-rendered session read must match a rendered private Oathkeeper rule,
  forward the exact Cookie and only the incoming X-Dev-Branch, open no
  connection to the public API, and give the browser only public API URLs.
  Loopback fixtures stand in for Oathkeeper and the API. Webv3 additionally proves non-root, read-only Nitro
  startup, Lab CA loading, SSR and compiled assets on canonical and branch
  hosts, development product links and healthy probes during an API outage.
  Authv2, built from the tested revision, proves its non-root image, Lab CA and
  compiled assets, and that during an API outage canonical and branch hosts show
  a retry page with a branch-preserving home link and no account-app choice.
  It also checks the rendered leaderboard worker ownership and the
  [development email](#development-email) binding.
- It cleans up only its labeled fixtures. Synthetic credentials in its fixture
  logs are not live credentials.

The offline gate **does not prove** Argo CD sequencing, a failed schema
migration blocking a live rollout, Image Updater commits, ingress and
authentication, or branch isolation. Those remain mandatory live gates after
activation: verify the real browser login and leaderboard, two owners,
independent frontend and API fallback, HMR, Go binary replacement, migration
failure gating and scoped cleanup. `.dev/acceptance.md` lists the dev-cli gates.

For tenant lifecycle changes, also prove frontend-only writes, separate owners,
base and branch worker ownership, open admission, in-flight teardown, complete
provider cleanup and TTL cleanup with unchanged canonical row counts and
database names.

For the isolated migration profile, prove the scratch migration version, owner
and marker, tenant-scoped API persistence, refusal to switch a live profile,
recorded-database teardown, provider cleanup and retained database, as listed
in `.dev/acceptance.md`.

Retain browser traces and screenshots outside source control, and identify the
exact tested revision and any substituted routing boundary.

## Recovery

| Symptom | Action |
| --- | --- |
| A migration Job failed and later waves did not start | Diagnose the retained Job, fix forward and run another full sync. For a dirty Tadoku API schema, follow [Database migration recovery](./migration-recovery.md). |
| The initial sync waits at the auth-provider wave | Run the bootstrap script once the namespaces and operator Secrets exist. |
| Consumers fail after operator credentials rotated | Rerun the bootstrap script, then perform an explicitly approved consumer restart. |
| Resources are healthy but a host serves another workload | Look for competing Ingresses or HTTPRoutes on the canonical hosts. Owner-scoped dev-cli branch Ingresses labelled `dev-cli.io/managed-by=dev-cli` in `tdk-dev-routing` are expected. |
| Leaderboards are stale | Run the [asynchronous worker checks](#asynchronous-worker-ownership). |

Do not restart shared services or delete databases as a troubleshooting
shortcut, and do not delete old resources or data without explicit approval.
