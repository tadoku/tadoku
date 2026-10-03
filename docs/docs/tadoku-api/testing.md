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
  abstractions solely to enable such assertions. When a test is explicitly
  removed, do not recreate equivalent coverage elsewhere unless asked.
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
and the third one test function. CI runs the database suites in normal and race
builds.

## Test infrastructure

Tests never skip. Missing or unsafe configuration fails before any connection.
Never point tests at shared development or production services.

- **PostgreSQL:** `TADOKU_TEST_POSTGRES_URL` must point to an explicit loopback
  port, database `postgres`, credentials `postgres:postgres` and exactly
  `sslmode=disable`. Fixtures create random disposable databases and apply the
  complete migration history.

  `testpostgres.Database.Pool` is the owner pool for fixture reset and
  verification SQL. `AppPool` uses a separate randomly named per-database
  login with no superuser, `bypassrls`, ownership or role memberships. It has
  application DML, sequence usage and read-only migration metadata access.
  Owner-only baseline snapshots are not granted to it. Application-pool
  shutdown precedes database and role cleanup; partial setup also cleans up
  the owned role and reports cleanup failures.
- **PgBouncer:** PostgreSQL transport tests require
  `TADOKU_TEST_PGBOUNCER_URL` with the same loopback, synthetic credential and
  database guard. Run PgBouncer 1.25.2 in transaction mode, with a wildcard
  database mapping to the disposable PostgreSQL instance, SCRAM authentication
  and nonzero `max_prepared_statements`. The dedicated target exercises 300
  goroutines, 12,000 tenant reads and 6,000 writes. Its test-only
  session-setting negative control must observe cross-client leaks. Normal and race CI jobs
  print both workload and negative-control results.
- **Keto:** relationship scenarios start a pinned, official Linux x86-64 Keto
  v25.4.0 SQLite-enabled executable under Bazel. Each helper owns an in-memory,
  loopback-only process using `infra/dev/ory/namespaces.keto.ts` and never
  accepts an external target.
- **Kratos:** HTTP E2Es own a pinned Kratos process; see
  [Kratos fixture](./http-e2e.md#kratos-fixture). No Docker service, external Kratos URL or
  Kratos environment variable is needed.
- **Valkey:** raw Valkey and application lifecycle tests require
  `TADOKU_TEST_VALKEY_URL` in the exact form `redis://127.0.0.1:<port>` (or
  `localhost`) for a disposable Valkey 9 service. They never flush the shared
  instance. Logical databases are reserved per suite: 15 for HTTP E2Es, 14 for
  their nested cleanup probe, and 13 for the leaderboard cache fence and worker
  tests. The HTTP E2E fixtures lease their database. Tests sharing database 13
  use unique parsed `e2e/<id>` tenants: the fence and cross-tenant tests create
  and delete only their named tenant keys; worker cache tests register a unique
  test tenant, persist it on jobs and delete only that tenant's matching keys.
  The worker's extra marker outside its tenant prefix also embeds the unique
  key and is deleted by exact name. No test sharing database 13 flushes it.
  The leased HTTP E2E database refuses pre-existing leaderboard keys and resets
  only legacy and tenant-derived leaderboard patterns while holding its lease,
  including keys created by arbitrary test-tenant jobs.

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
- Keep the pool-closing failure test isolated. It opens a second pool on the
  shared DSN and closes it, without creating another migrated database.
- The `services/tadoku-api/infra/postgres/` helper tests have their own setup in
  `services/tadoku-api/infra/postgres/README.md`.

## Tenancy guard and isolation

`services/tadoku-api/app/tenantlifecycle/` tests the owner command's orchestration
against disposable PostgreSQL, Valkey and the pinned Keto process. Its complete
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

`services/tadoku-api/e2e/tenant_isolation_test.go` runs signed requests through
the production JWT/Keto router and real PostgreSQL and Valkey. It covers
two-way log visibility, duplicate page slugs and synchronized identities,
shared-table write denial, base and branch job claims, actual delegated
provider tenant contexts, and test-tenant deletion including a replay chain
without changing canonical rows. Pool tests also prove fresh and reused
unscoped connections fail closed. These checks cover database/context
isolation. The leaderboard Valkey tests also prove that invalidating one tenant
keeps the other tenant's warm page and that missing-tenant invalidation writes
no key.

Development seed SQL requires `psql` variables including `tenant`. Exercise
the three files in `scripts/dev/seed/` against a disposable migrated database,
with canonical and test keys, and retain the command and row/ID comparison.
If the test runner does not provide `psql`, attach that bounded manual run
as verification evidence rather than skipping it silently.
