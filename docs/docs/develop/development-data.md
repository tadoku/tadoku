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
release. V1 identities and Keto providers are shared; v2 accounts and sessions
belong to their tenant.

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
  fixture user rows in `users`. If they are missing, ask the operator to run
  `make dev-seed` first; it changes shared identities and base data, and the
  branch task never seeds the base. Before loading its profiles, it
  creates native v2 accounts with those same UUIDs and the names Dev Admin and
  Dev Reader in its own tenant. Passwords and emails use the same
  `TADOKU_DEV_ADMIN_*` / `TADOKU_DEV_READER_*` source and defaults as the base
  seed. Custom values must be supplied privately in the tenant Job environment,
  through an ignored task-template override or Secret references.
- Provisioning uses `tadoku_owner` to create the test tenant, grants its Keto
  object a parent of `app:tadoku`, and grants Dev Reader direct `testers`
  membership. Dev Admin inherits administrator permits.
- Flipt uses environment `test` and namespace `tadoku_<route>`, with display
  name `tadoku/<route>`. Provisioning adds missing resources from this branch's
  `k8s/dev/base/flipt/features.yaml`, plus a branch-only `identity-v2` flag with
  a 100% true rollout. Canonical definitions/imports and the safe false default
  are unchanged. Existing flag values and grants are preserved on retries.
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

Native branch accounts use the existing documented fixture password, including
its legacy short length; only the development fixture command permits this.
Public registration retains its normal password policy. Seeding inserts missing
accounts by the exact tenant and fixture UUID. Existing accounts, including
branch-changed names, emails, passwords, state, timestamps and sessions, stay
unchanged. If a missing fixture's configured email belongs to a different UUID
in that tenant, the whole fixture pair fails atomically. Other tenants may use
the same UUIDs and emails with independent credentials. New fixture users retain
the canonical fixture user's creation time when available, and new accounts
retain their local user's creation time, preserving account-age rules. It creates
no session; sign in through the configured branch API to create a native session.
The branch SQL sets `seed_users=false` before loading immersion fixtures because
the native fixture command owns initialization; reseeding still populates its
profiles, contests and content without resetting user names. The human base seed
keeps its existing user refresh behavior.

Dev Admin keeps its base UUID, so the branch Keto parent preserves inherited
administrator permits without changing canonical tuples. Dev Reader has direct
tester membership. Native sessions use the tenant-specific cookie; v1 Kratos
sessions do not authenticate native accounts. Do not change shared fixture
passwords, profiles or canonical roles. Down deletes the branch accounts,
sessions, recovery tokens, emails and access tokens through the tenant cascade.
Provider processes and runtime credentials remain shared.

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
