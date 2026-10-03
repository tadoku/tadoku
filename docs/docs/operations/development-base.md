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

Each component runs in its own `tdk-dev-*` namespace:

| Namespace | Contents |
| --- | --- |
| `tdk-dev-frontend-webv2`, `tdk-dev-frontend-auth`, `tdk-dev-frontend-admin` | The three frontends |
| `tdk-dev-tadoku-api` | Tadoku API and the private asynchronous worker |
| `tdk-dev-kratos`, `tdk-dev-keto`, `tdk-dev-oathkeeper` | Auth providers |
| `tdk-dev-flipt` | Feature flags |
| `tdk-dev-token-reflector` | Token-reflector |
| `tdk-dev-data` | One operator-managed Postgres server (`tadoku-dev-db`), disposable Valkey and Mailhog |
| `tdk-dev-routing` | Application routes attached to the platform Envoy Gateway |

The Postgres server holds the base `tadoku`, `kratos` and `keto` databases and
previously retained `tadoku-<route>` branch databases. Ordinary branches use
their own test tenant, `tadoku/<route>`, in `tadoku`; canonical base data uses
`tadoku/prod`. Row-level security and transaction-local tenant context isolate
their application rows. No ordinary branch database is created. The server is
provisioned with the Zalando `postgresql` custom resource; do not add hand-rolled Postgres Deployments or Helm
releases. Styleguides and optional admin tools are not deployed. The base holds no production data, external backups or
notification configuration.

## Images and Image Updater

- The repository's existing CI publishes the GHCR runtime and migration images.
  Do not add another build or push pipeline.
- The `images` entries in `k8s/dev/base/kustomization.yaml` select `latest`.
  The development Image Updater uses the **digest** strategy and writes the
  immutable resolutions back to that file.
- The worker has its own GHCR image and Kustomize image entry. Image Updater
  tracks its digest independently of the API image.
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
before starting Next. The public Lab CA is mounted for server-side HTTPS.

- It is a deployment adapter, not a source-sync server or reusable CLI runtime.
- It calls a private Next startup API. Reverify it whenever the frontend
  framework version changes.
- Branch pods use the dev-cli-managed pnpm Next.js dev server, not this adapter.

## Asynchronous worker ownership

The private `tadoku-worker` Deployment consumes `jobs` in the base
`tadoku` database across tenants without a matching `tadoku-worker` override.
It derives leaderboard keys from each persisted job tenant. It has one
replica, Recreate rollout, a separate image digest, and no Service or public
route. Its resource requests and limits are in
`k8s/dev/base/services/tadoku-worker.yaml`.

The worker consumes `jobs` and invalidates leaderboard caches. Verify worker
ownership and reads:

1. Confirm the base `tadoku-worker` is ready and its `/readyz` endpoint remains
   healthy while tasks retry or fail.
2. Inspect due and failed `jobs` rows; a write followed by a
   leaderboard read must still succeed.

dev-cli pairs each branch API and worker against the shared database with
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

## Automatic migrations

Full Argo CD syncs apply these waves:

| Wave | Resources and gate |
| --- | --- |
| -30 | Development namespaces |
| -20 | Postgres custom resource and generated configuration |
| -15 | Transfer existing Tadoku objects to `tadoku_owner` and establish runtime grants |
| -10 | Tadoku API migrations as `tadoku_owner`; Kratos and Keto migration Sync hooks; each waits for authenticated database connectivity |
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

Follow the repository's migration-first release contract: schema changes land
and deploy on their own before dependent runtime code. Digest tracking is not a
substitute for backward-compatible schema and application releases.

Base seeding is separate from migrations: `make dev-seed` creates the marked
synthetic identities and base fixtures on `homelab-talos-dev` for tenant
`tadoku/prod`. SQL seeds require the supplied tenant, set it for each
transaction and write that tenant explicitly. The canonical tenant retains the
fixture UUIDs; other tenants derive fixture UUIDs from their tenant key.
Caller-supplied identity UUIDs are unchanged.

Branch startup uses the `tenant` hook to provision and seed its test tenant;
see [Development environment](../develop/environment.md#branch-tenants).
The Job reuses the marked shared identities, grants the Keto object a canonical
parent and Dev Reader tester, and provisions missing Flipt resources in
`test/tadoku_<route>`. It runs in `tdk-dev-data`; Flipt's ingress policy admits
only labeled tenant-lifecycle pods from that namespace for these direct
management calls.

## Credentials

There are no plaintext Secret manifests or private keys in `k8s/dev/base/`.

- The Postgres operator generates `tadoku_owner`, `tadoku`, `kratos` and `keto`
  credentials in `tdk-dev-data`. The owner Secret is named
  `tadoku-owner.tadoku-dev-db.credentials.postgresql.acid.zalan.do`; the
  operator replaces underscores in Secret names. The base Tadoku migrations
  use this owner Secret. Tenant lifecycle Jobs use it for tenant and override
  writes. API, worker and SQL fixture writes keep the `tadoku` credentials.
- The Postgres administrator Secret stays in `tdk-dev-data`. The ownership
  hooks and short-lived branch database creation Job use it. Branch migration
  tasks use `tadoku_owner`; ordinary tenant tasks do not use administrator
  credentials. API, worker and fixture seeding use `tadoku`. This
  is cooperative isolation, not hostile multi-tenancy.
- `scripts/dev/bootstrap-gitops-secrets.sh` copies only the required
  credentials, including the owner credential for branch migrations, into
  consumer namespaces. It targets only `homelab-talos-dev` at
  `https://omni.lab:8100`. It generates the development-only Kratos runtime
  secret, Oathkeeper JWKS and Oathkeeper authorization token once, and
  preserves existing keys on rerun. It needs `kubectl`, `jq` and `node`.
- The script fails closed on another API server, on namespaces not labeled as
  part of the development base, and on Secrets with unexpected ownership. It
  never reads credentials from outside the development GitOps base.
- Run it only with explicit credential or bootstrap approval.
- Keep generated keys only in development Kubernetes Secrets or an approved
  encrypted backup, never in Git or a worktree. Commit only the public Lab CA
  and Secret references, never private keys, kubeconfigs, tokens or plaintext
  Secrets.

When operator credentials rotate, rerun the bootstrap script, then perform an
explicitly approved restart of the consumers. Kubernetes does not refresh
environment variables in running processes.

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
5. Wait for the migration Jobs and base workloads to become Healthy, then seed.

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
- `dev down` and TTL cleanup remove the selected test tenant's Keto tuples,
  Valkey keys and Flipt namespace before deleting its PostgreSQL tenant row and
  cascading application rows. They stop branch pods and clear their worker
  overrides first. Canonical `tadoku/prod` is refused by the lifecycle tool.
- Teardown failures retain the lifecycle marker for retry. The tenant row is
  deleted last so an external-provider failure preserves an inventory of the
  remaining tenant. Do not delete its marker or task Lease to bypass a failure.
- Previously retained branch databases remain retained on down and TTL
  cleanup. Deleting one requires an exact-name inventory and explicit approval:
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
- `infra/dev/ory/` holds the Kratos schema and Keto namespace fixtures that Bazel
  backend tests load; they are not obsolete deployment files. Preserve the shared
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

- It records the source revision, worktree status, rendered-manifest hash,
  exact image digests, commands and results in `report.json`, with individual
  logs.
- It proves fresh and existing ownership transfer, runtime DML and sequence
  access, denied DDL and migration-table writes, and branch provisioning.
- It also proves fresh migrations, no-op reruns, an intentional connection
  failure and recovery, plus HTTP readiness, development runtime URLs and
  compiled assets in all three published frontend images with external
  networking disabled. It also checks the rendered leaderboard worker ownership.
- It cleans up only its labeled fixtures. Synthetic credentials in its fixture
  logs are not live credentials.

The offline gate **does not prove** Argo CD sequencing, a failed schema
migration blocking a live rollout, Image Updater commits, ingress and
authentication, or branch isolation. Those remain mandatory live gates after
activation: verify the real browser login and leaderboard, two owners,
independent frontend and API fallback, HMR, Go binary replacement, migration
failure gating and scoped cleanup. `.dev/acceptance.md` lists the dev-cli gates.

For tenant lifecycle changes, also prove frontend-only writes, separate owners,
base and branch worker ownership, access permits, in-flight teardown, complete
provider cleanup and TTL cleanup with unchanged canonical row counts and
database names.

Retain browser traces and screenshots outside source control, and identify the
exact tested revision and any substituted routing boundary.

## Recovery

| Symptom | Action |
| --- | --- |
| A migration Job failed and later waves did not start | Diagnose the retained Job, fix forward and run another full sync. For a dirty Tadoku API schema, follow [Database migration recovery](./migration-recovery.md). |
| The initial sync waits at the auth-provider wave | Run the bootstrap script once the namespaces and operator Secrets exist. |
| Consumers fail after operator credentials rotated | Rerun the bootstrap script, then perform an explicitly approved consumer restart. |
| Resources are healthy but a host serves another workload | Look for competing Ingresses or HTTPRoutes on the canonical hosts. |
| Leaderboards are stale | Run the [asynchronous worker checks](#asynchronous-worker-ownership). |

Do not restart shared services or delete databases as a troubleshooting
shortcut, and do not delete old resources or data without explicit approval.
