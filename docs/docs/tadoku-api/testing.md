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
  their nested cleanup probe, 13 for the leaderboard cache fence test and 12 for
  the enabled worker lifecycle test. The E2E and lifecycle fixtures lease their
  database; the cache fence test uses unique scoped keys. Every fixture refuses
  to overwrite pre-existing fixture keys and deletes only its own keys during
  cleanup.

## Repository and package tests

- Database tests in feature packages follow the
  [feature package layout](./conventions.md#feature-package-layout).
- Database helpers take contexts and return errors, with explicit `Close`
  cleanup instead of depending on `testing.TB`. Suite teardown preserves test
  failures and reports cleanup failures; partial setup also cleans up.
- Repository and transaction tests keep their own independent databases and may
  run in parallel.
- Keep the pool-closing failure test isolated. It opens a second pool on the
  shared DSN and closes it, without creating another migrated database.
- The `services/tadoku-api/infra/postgres/` helper tests have their own setup in
  `services/tadoku-api/infra/postgres/README.md`.
