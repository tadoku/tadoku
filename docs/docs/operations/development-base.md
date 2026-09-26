---
title: Development base
description: How Argo CD runs the always-on Tadoku development base on homelab-dev, and how operators bootstrap, migrate, verify and recover it.
---

# Development base

Read this when you operate the Argo CD development base in `k8s/dev/base/`:
first activation, credentials, migrations, image updates, worker ownership or
recovery.

`k8s/dev/base/` is exclusively for the `homelab-dev` cluster (node `ct190`).
**Never apply it to another Kubernetes context.** Production is deployed from a
private repository and is not affected by these manifests. To run and verify
your own branch, use [Development environment](../develop/environment.md)
instead; for a map of the components, see
[System architecture](../architecture/index.md).

## Ownership

| Owner | Owns |
| --- | --- |
| Argo CD | The always-running base defined in `k8s/dev/base/`, including automatic migration Jobs, base workloads and canonical routing |
| dev-cli | Branch overlays, routing overrides, short-lived tasks and owner/branch database provisioning |
| Homelab infrastructure | The Argo CD Application and the development Image Updater, not copies of these workload manifests |
| Platform | The Postgres operator and Envoy Gateway, which must exist before the base syncs |

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
every `tadoku-<route>` branch database. It is provisioned with the Zalando
`postgresql` custom resource; do not add hand-rolled Postgres Deployments or Helm
releases. Styleguides and optional admin tools are not deployed. The base holds no production data, external backups or
notification configuration.

## Images and Image Updater

- The repository's existing CI publishes the GHCR runtime and migration images.
  Do not add another build or push pipeline.
- The `images` entries in `k8s/dev/base/kustomization.yaml` select `latest`.
  The development Image Updater uses the **digest** strategy and writes the
  immutable resolutions back to that file.
- The worker has its own GHCR image and Kustomize image entry. Its first digest
  must be written by Image Updater after the worker image is published; confirm
  the external Image Updater tracks the new image before activating the worker.
  An API image update must leave the worker image digest and Pod template intact.
- Hook migration images need Image Updater's `force-update`, because successful
  Jobs are removed from the live resource list.
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

The private `tadoku-worker` Deployment consumes `async_outbox` in the base
`tadoku` database and uses unprefixed leaderboard cache keys. It has one
replica, Recreate rollout, a separate image digest, and no Service or public
route. Its CPU and memory limits are initial values; verify them with four
active tasks and startup reconciliation before treating them as settled.

During the outbox transfer, the API's embedded worker remains enabled and
consumes only the old `leaderboard_outbox` table. The new worker consumes only
`async_outbox`. API cache reads retain the embedded worker’s existing readiness
behavior. Verify both queues and owners independently:

1. Confirm the base `tadoku-worker` is ready and its `/readyz` endpoint remains
   healthy while tasks retry or fail.
2. Confirm one base API replica has `API_LEADERBOARD_OUTBOX_ENABLED=true` and
   logs `leaderboard outbox ready` for the old table.
3. Inspect due and failed `async_outbox` rows and old pending rows; a write
   followed by a leaderboard read must still succeed.

DevCLI pairs each branch API and worker against `tadoku-${DEV_ROUTE}` and the
`dev:${DEV_ROUTE}:` Valkey prefix. Selecting either workload starts both;
unchanged peers use their current image. Verify branch writes, worker
completion, and leaderboard reads without changing base or another branch.
The offline E2E (see [Verification](#verification)) checks rendered ownership,
isolation, image separation and lack of worker routing. Tadoku API's backend
E2Es check separate PostgreSQL databases with shared Valkey cache namespaces.

## Automatic migrations

Full Argo CD syncs apply these waves:

| Wave | Resources and gate |
| --- | --- |
| -30 | Development namespaces |
| -20 | Postgres custom resource and generated configuration |
| -10 | Tadoku API, Kratos and Keto migration Sync hooks; each waits for authenticated database connectivity |
| 0 | Auth providers, cache, Flipt, token-reflector and Gateway routes |
| 10 | Oathkeeper, which publishes its JWKS before the APIs start |
| 20 | Tadoku API |
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
synthetic identities and base fixtures. Branch databases are migrated and
seeded by dev-cli tasks; see
[Development environment](../develop/environment.md#branch-databases).

## Credentials

There are no plaintext Secret manifests or private keys in `k8s/dev/base/`.

- The Postgres operator generates the `tadoku`, `kratos` and `keto` credentials
  in `tdk-dev-data`. Tadoku API uses the `tadoku` role and database.
- The Postgres administrator Secret stays in `tdk-dev-data`. Only the
  short-lived branch database creation Job uses it; application pods and tasks
  use the `tadoku` role. This is cooperative isolation, not hostile
  multi-tenancy.
- `scripts/dev/bootstrap-gitops-secrets.sh` copies only the required
  credentials into consumer namespaces. It generates the development-only
  Kratos runtime secret, Oathkeeper JWKS and Oathkeeper authorization token
  once, and preserves existing keys on rerun. It needs `kubectl`, `jq` and
  `node`.
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
- Branch databases are retained on `dev down` and TTL cleanup. Deleting one
  requires exact ownership checks and permission.
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
kubectl --context homelab-dev kustomize k8s/dev/base |
  kubeconform -strict -summary -ignore-missing-schemas
node k8s/dev/base/e2e.cjs /tmp/tadoku-base-evidence
```

Install the frontend's locked pnpm dependencies first; the E2E harness reuses
its YAML parser. The harness runs the real published migration images in
isolated, resource-bounded Docker containers, not a nested Kubernetes cluster.

- It records the source revision, worktree status, rendered-manifest hash,
  exact image digests, commands and results in `report.json`, with individual
  logs.
- It proves fresh migrations, no-op reruns, an intentional connection failure
  and recovery, plus HTTP readiness, development runtime URLs and compiled
  assets in all three published frontend images with external networking
  disabled. It also checks the rendered leaderboard worker ownership.
- It cleans up only its labeled fixtures. Synthetic credentials in its fixture
  logs are not live credentials.

The offline gate **does not prove** Argo CD sequencing, a failed schema
migration blocking a live rollout, Image Updater commits, ingress and
authentication, or branch isolation. Those remain mandatory live gates after
activation: verify the real browser login and leaderboard, two owners,
independent frontend and API fallback, HMR, Go binary replacement, migration
failure gating and scoped cleanup. `.dev/acceptance.md` lists the dev-cli gates.
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
