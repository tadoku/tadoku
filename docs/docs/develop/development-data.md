---
title: Development data
description: Which databases and tenants the development base and your branch use, how to seed fixture accounts, how migration branches get an isolated database, and how leaderboard caches and jobs are scoped on shared Valkey.
---

# Development data

Read this when you seed the shared base, need to know which database and tenant
your branch uses, run a branch that changes a migration, or check leaderboard
caches and jobs on a branch. [Development environment](./environment.md) covers
installing dev-cli and starting, opening and cleaning up a branch.

## Shared base data

Argo CD runs the base migrations automatically during full syncs; see
[Development base](../operations/development-base.md#automatic-migrations).
Selective resource sync skips those migration hooks, so never use it for a
release. Kratos and Keto are shared by the base and every branch.

## Seed data

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

## Branch tenants

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
  membership. Dev Admin inherits administrator permits.
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

## Migration branches

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

## Leaderboards on shared Valkey

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
  [Migrate a job contract from v1 to v2](../tadoku-api/jobs-versioning.md)
  before changing a payload; pairing workloads does not make incompatible
  versions safe.
- To check your worker's readiness, metrics and job rows, follow
  [Run jobs on your development branch](../tadoku-api/jobs-operations.md#run-jobs-on-your-development-branch).
