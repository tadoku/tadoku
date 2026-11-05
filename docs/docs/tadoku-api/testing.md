---
title: Testing
description: What to test in Tadoku API and how, from test principles and shared infrastructure to repository and package tests, and where HTTP end-to-end tests and user journeys are documented.
sidebar_position: 5
---

# Testing

Read this when you add or change a Tadoku API test. HTTP golden cases are in
[HTTP end-to-end tests](./http-e2e.md) and chained scenarios in
[User journeys](./user-journeys.md).

## Principles

- Test meaningful behavior and risk, not every change. Add coverage where a
  regression would matter: production authentication and HTTP contracts,
  persistence and data integrity, and nontrivial domain rules.
- There is no test-per-method or test-per-change requirement. No new test is a
  valid choice for trivial, documentation-only, generated-output-only and
  test-removal changes.
- Do not add low-value tests: tautological assertions that mirror the
  implementation, trivial constant or schema equality checks, pass-through or
  error-wrapper assertions without distinct behavior, or coverage already
  provided at a shared boundary.
- Do not export internals, broaden visibility, or add dependencies or
  abstractions solely to enable such assertions. Broadening visibility so a
  test fixture builds a dependency the way production does is not such an
  assertion; record it in [Import boundaries](./import-boundaries.md). When a
  test is explicitly removed, do not recreate equivalent coverage elsewhere
  unless asked.
- Prefer real-database repository tests plus HTTP E2Es over isolated
  feature-service tests:
  - Repository tests are highly recommended for query behavior, row mapping,
    constraints and persistence.
  - HTTP E2Es cover feature-service orchestration and the HTTP contract. Do not
    add database-backed feature-service tests.
  - Unit tests cover pure parameter validation and domain rules.
  - Cover shared dependency failures and route ownership once, at their own
    boundaries, not again for every operation that uses them.
- Use Go's standard `testing` package for new or rewritten tests: ordinary
  comparisons, `t.Fatalf` for failed prerequisites, `t.Errorf` for independent
  checks, `t.Cleanup` for resource cleanup, and `errors.Is` or `errors.As` for
  errors. Do not add Testify, another assertion framework or a homegrown
  assertion DSL. Existing tests need no bulk rewrite; convert them when you
  rework their code or in a separately scoped mechanical change.

## Running tests

```sh
bazel test //services/tadoku-api/... --test_output=errors
bazel test //services/tadoku-api/e2e:e2e_test --test_output=errors
bazel test //services/tadoku-api/e2e:e2e_test --test_filter=TestAuthentication
```

The first command runs everything, the second the HTTP E2Es and user journeys,
and the third one test function. CI runs the database suites in the normal build
of every change and in the daily race build; see
[Race tests](../develop/race-tests.md).

## Test infrastructure

Tests never skip. Missing or unsafe configuration fails before any connection.
Never point tests at shared development or production services.

- **PostgreSQL:** `TADOKU_TEST_POSTGRES_URL` must point to an explicit loopback
  port, database `postgres`, credentials `postgres:postgres` and exactly
  `sslmode=disable`. Fixtures clone random disposable databases from a template
  holding the complete migration history. The first fixture on a server builds
  the template, named after the up-migration contents, while other test
  processes wait on an advisory lock; templates are retained and refuse
  connections. The admin pool is capped at two connections;
  owner and restricted application pools each allow four.

  `testpostgres.Database.Pool` is the owner pool for fixture reset and
  verification SQL. `AppPool` uses a separate randomly named per-database
  login with no superuser, `bypassrls`, ownership or role memberships. It has
  application DML, sequence usage and read-only migration metadata access.
  Owner-only baseline snapshots are not granted to it. Application-pool
  shutdown precedes database and role cleanup; partial setup also cleans up
  the owned role and reports cleanup failures.

  `SeedTenant` runs a fixture transaction bound to one test tenant through a
  second per-database login, capped at two connections. The login holds DML
  only on tables whose every row-level security policy is exactly the
  `tadoku.tenant` binding, derived from the catalog, plus their sequences. It
  has nothing on the tenant registry, overrides, `jobs`, shared reference
  tables or migration metadata, and the same pool shutdown, role cleanup and
  partial-setup handling. It fails if the transaction ends bound to another
  tenant or to all tenants.
- **PgBouncer:** PostgreSQL transport tests require
  `TADOKU_TEST_PGBOUNCER_URL` with the same loopback, synthetic credential and
  database guard. Run PgBouncer 1.25.2 in transaction mode, with a wildcard
  database mapping to the disposable PostgreSQL instance, SCRAM authentication
  and nonzero `max_prepared_statements`. The dedicated target exercises 300
  goroutines, 12,000 tenant reads and 6,000 writes. Its test-only
  session-setting negative control must observe cross-client leaks. The normal and race CI
  workflows print both workload and negative-control results.
- **Identity accounts:** ordinary HTTP golden and journey resets seed native
  tenant-owned accounts after case SQL, preserving existing user/account state
  on primary-key conflicts. Cast and gateway fixtures obtain sessions through
  real login with test-cost bcrypt hashes; registration/password-policy
  workflows retain production cost 12. See
  [Local account fixtures](./http-e2e.md#local-account-fixtures).
- **Valkey:** raw Valkey and application lifecycle tests require
  `TADOKU_TEST_VALKEY_URL` in the exact form `redis://127.0.0.1:<port>` (or
  `localhost`) for a disposable Valkey 9 service. They never flush the shared
  instance. Logical databases are reserved per suite: HTTP E2E shards use
  15, 14, 12, 11 and 10, and the leaderboard cache fence and worker tests use 13.
  Unsharded HTTP E2Es use 15. Invalid or incomplete Bazel shard coordinates,
  including more than five shards, fail before opening fixtures. The HTTP E2E fixtures lease their database. Tests sharing database 13
  use unique parsed `e2e/<id>` tenants: the fence and cross-tenant tests create
  and delete only their named tenant keys; worker cache tests register a unique
  test tenant, persist it on jobs and delete only that tenant's matching keys.
  The worker's extra marker outside its tenant prefix also embeds the unique
  key and is deleted by exact name. No test sharing database 13 flushes it.
  The leased HTTP E2E database refuses pre-existing leaderboard or identity
  throttle keys. It resets only legacy/tenant-derived leaderboard patterns and
  tenant-derived identity login, registration and recovery patterns while holding its
  lease, including leaderboard keys created by arbitrary test-tenant jobs.

## HTTP shard coverage

The normal HTTP E2E target declares three native Bazel shards. Race builds use
one E2E process while retaining the complete suite. Overall test-target
concurrency still needs a runner-appropriate limit.

Use native Bazel `shard_count` to select top-level tests; do not add an
application test selector. Within a process, only scenario tests that meet the
[parallel scenario](./scenario-fixtures.md#parallel-scenarios) rules call
`t.Parallel`, at most two at a time; cases that share fixture state stay serial.
`--test_sharding_strategy=disabled` runs the same target in one process for
comparison. Keep `--nocache_test_results` on measured runs.

Retain the complete successful test XML from an unfiltered unsharded run and
all shards of an equivalent run. Compare exact `(classname, name)` identities,
including parent tests and subtests, with
`services/tadoku-api/e2e/check_shard_inventory.py`:

```sh
python3 services/tadoku-api/e2e/check_shard_inventory.py \
  --baseline "$EVIDENCE/unsharded/test.xml" \
  --shards "$EVIDENCE"/sharded/shard_*/test.xml
```

Run `python3 services/tadoku-api/e2e/check_shard_inventory_test.py` to exercise
the checker's positive and negative controls.

The comparison fails on empty inventories, unsuccessful cases, duplicates,
missing names or unexpected names. A case count alone cannot prove completeness.
Use the same revision and no test filter for both runs. Preserve an inventory
from before fixture changes as well, so review can establish that existing cases
remain covered. Retain commands, revision, setup and resource samples alongside
XML; CI artifacts must include successful shard XML as well as failures.

Limit concurrent test processes using Bazel's `--local_test_jobs` and the
runner's measured memory, CPU and database capacity. The number of declared
shards is not the number that must run simultaneously. Measure normal and race
builds separately; instrumentation changes memory and CPU demand. Do not infer
CI latency from a host with more available resources.

## Repository and package tests

- Test code follows [Go style](./go-style.md), including arrange, act and
  assert grouping and one test case per table entry.
- Database tests in feature packages follow the
  [feature package layout](./conventions.md#feature-package-layout).
- Database helpers take contexts and return errors, with explicit `Close`
  cleanup instead of depending on `testing.TB`. Suite teardown preserves test
  failures and reports cleanup failures; partial setup also cleans up.
- Direct repository, service and worker tests pass an explicit
  `tenant.WithKey(ctx, tenant.Production())` context. HTTP E2Es receive the
  tenant through their signed fixture tokens. Fixture reset sets the canonical tenant
  transaction-locally before cleanup and seed inserts.
- Raw owner-pool fixture inserts name the tenant column explicitly, including
  every row and `insert ... select` projection. Canonical fixtures use
  `tadoku/prod`; parent and child rows use the same tenant. A context value
  alone does not scope raw pool SQL. Reads through a restricted application role use
  the PostgreSQL executor with an explicit tenant context, preserving the same
  transaction-local scope as application calls. Never supply a session, role,
  database or pool startup tenant setting to make fixtures pass.
- Repository and transaction tests keep their own independent databases and may
  run in parallel. If local test targets contend with the provider fixtures,
  use `--local_test_jobs=1` to serialize targets while retaining all assertions.
- Parallel worker tests create their fixtures after `t.Parallel`, own their
  databases and provider keys, and stop every worker goroutine before fixture
  cleanup. Tests that change process-wide state, including `timex.TheWorld`,
  `uuid.SetRand`, `jwt.TimeFunc` or environment variables, and their parents
  stay serial. Bound in-process concurrency with `-test.parallel` as well as
  target concurrency when measuring against CPU, memory and connection limits.
- Split a slow target whose tests must stay serial with native Bazel
  `shard_count`; shards are separate processes, so process-wide state and
  `t.Setenv` stay safe. Tests that create real Flipt SDK clients stay serial
  within a process: concurrent clients race in the SDK's WebAssembly runtime.
  Before sharding, compare `-test.v` pass inventories of a sharded and an
  `--test_sharding_strategy=disabled` run.
- Keep the pool-closing failure test isolated. It opens a second pool on the
  shared DSN and closes it, without creating another migrated database.
- The `services/tadoku-api/infra/postgres/` helper tests have their own setup in
  `services/tadoku-api/infra/postgres/README.md`.

## Property tests

- Write a
  [property-based test](https://github.com/flyingmutant/rapid#what-is-property-based-testing)
  with [rapid](https://pkg.go.dev/pgregory.net/rapid) when a rule must hold
  over a combinatorial input space and an independent model states the
  expected result. Keep example tests for the cases a reader should see.
  Never require random cases to reach an outcome; check each outcome a
  property must cover on a fixed case through the same assertions.
- Draw from small pools so values collide, and build them through the public
  constructors. Generate invalid input only when the property is about
  rejecting it. Rapid seeds, shrinks and prints counterexamples; do not do any
  of that by hand.
- Test a stateful component with
  [`T.Repeat`](https://pkg.go.dev/pgregory.net/rapid#T.Repeat). Each action
  calls the real component, and the invariant compares its state with an
  in-memory model of the expected state. The model is an oracle, not a fake of
  a port; storage stays real.
- A database property keeps one disposable database and deletes its rows at
  the start of each case.
- Every run draws a fresh seed. A target with property tests is tagged
  `no-cache`, so CI never reuses its result, and passes `-rapid.nofailfile` in
  its `args`, because a fail file stays in the Bazel sandbox. Locally,
  `--nocache_test_results` reruns an unchanged target. Replay a failure with
  the seed it prints:
  `bazel test <target> --test_filter=<Test> --test_arg=-rapid.seed=<seed>`.
  `--test_arg=-rapid.checks=<n>` runs more cases and `--test_arg=-test.short`
  fewer. Keep a counterexample worth reading as an example test.

## Tenancy guard and isolation

`services/tadoku-api/app/tenantlifecycle/` tests the owner command's
orchestration against disposable PostgreSQL and Valkey. Its complete
tenant-column inventory includes registry overrides; every ordinary owned table
has a nonempty teardown fixture. Canonical and another test tenant's complete
row snapshots remain unchanged. The Flipt HTTP fixture records resource
creation, while an injected outage verifies that the registry remains retryable.
Use `TADOKU_TEST_POSTGRES_URL` and `TADOKU_TEST_VALKEY_URL` for this suite, with
the same loopback guards as the API tests.

`services/tadoku-api/infra/postgres/tenancy_schema_test.go` migrates a
disposable database and checks every ordinary table for RLS, a non-null tenant
column, the strict transaction tenant default and its canonical policy.
Registry, shared reference, migration metadata and the helper's baseline tables
have their explicit exceptions. The guard checks shared reference policies and
rejects database/role startup tenant settings. A new table therefore fails the
guard until it has the required protections.

All HTTP E2E application repositories and the real journey worker use
`AppPool`; reset, fixture mutations and `verify.sql` keep `Pool`.
[Scenario fixture](./scenario-fixtures.md) seeds use `SeedTenant`.

Tests using `testing/synctest` create and close every database pool used in the
bubble inside that same bubble. Copy the fixture pool configuration to preserve
the owner or restricted application role, then use the new pools for writes,
worker services and verification SQL. Shared pools can open connections inside
the bubble and abort the process when their synchronization crosses its boundary.
Worker timer and retry scenarios advance virtual time while retaining real
PostgreSQL, SMTP and worker execution; keep SMTP listeners outside the bubble
and all database work inside its owned pools.

`services/tadoku-api/e2e/tenant_isolation_test.go` runs signed requests through
the production router with real JWT verification, PostgreSQL and Valkey. It
covers two-way log visibility, duplicate page slugs and synchronized identities,
shared-table write denial, base and branch job claims, actual delegated
provider tenant contexts, and test-tenant deletion including a replay chain
without changing canonical rows. Pool tests also prove fresh and reused
unscoped connections fail closed. These checks cover database/context
isolation. The leaderboard Valkey tests also prove that invalidating one tenant
keeps the other tenant's warm page and that missing-tenant invalidation writes
no key.

Development seed SQL requires `psql` variables including `tenant`. Exercise
the files in `scripts/dev/seed/` against a disposable migrated database,
with canonical and test keys, and retain the command and row/ID comparison.
If the test runner does not provide `psql`, attach that bounded manual run
as verification evidence rather than skipping it silently.
