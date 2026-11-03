---
title: Scenario fixtures
description: How a Tadoku API E2E scenario owns private test tenants, seeds them under row-level security, signs its cast, scopes its workers and tears everything down, and which scenarios may overlap.
sidebar_position: 6.6
---

# Scenario fixtures

Read this when you write or convert an E2E scenario that owns its own tenants
instead of resetting the whole database. Golden-case layout is in
[HTTP end-to-end tests](./http-e2e.md) and journeys in
[User journeys](./user-journeys.md).

`services/tadoku-api/e2e/scenario_test.go` gives one independent scenario its
own tenants instead of a whole-database reset. Existing golden and journey
cases still use `reset`; a scenario never calls it.

- `newScenario(t, roles...)` allocates a fresh `e2e/<role>-<8 hex>` test
  tenant per semantic role, such as `identity-a` and `identity-b` for a paired
  scenario, and provisions each through `tenantlifecycle.Application` on the
  owner pool. Call it after any `t.Parallel`.
- Provisioning is independent of the canonical tenant's current state: each
  tenant starts with the migration-time platform scoring snapshot, no pages and
  a Flipt namespace from `k8s/dev/base/flipt/features.yaml`.
- `seed(t, role, tenantSeed{...})` runs the role's `setup` SQL,
  `testdata/tenant_accounts.sql` and its `finalize` SQL, then inserts its
  `relationships` rows, in one transaction through
  `testpostgres.Database.SeedTenant`. Its login holds DML only on tables whose
  every row-level security policy binds `tenant` to `tadoku.tenant`, and on
  their sequences. Row-level security confines every statement, including an
  unscoped `update` or `delete`, to the role's tenant. Writes to the tenant
  registry, tenant overrides, shared reference tables such as `languages` and
  migration metadata fail with `permission denied`, so a seed cannot register
  tenants or read shared reference rows. A seed that leaves its transaction
  bound to another tenant or to all tenants fails. Scenario relationship rows
  omit `tenant`; the role supplies it.
- `serve(t, role, member, request)` runs the request through the production
  router with a freshly signed token for that cast member and tenant. Tokens
  carry the checked-in cast's subjects, traits and timing relative to
  `jwt.TimeFunc`, signed by a key generated per test process; the local JWKS
  serves its public half next to the checked-in keys.
- `startWorker` and `runWorker` run workers scoped with `jobqueue.OnlyTenant`
  to one role's tenant; `runWorker` completes that tenant's jobs, including
  jobs they enqueue.
- `executionContext`, when a runner sets it, derives every request and worker
  context, so a scenario can supply its own business clock and ID source.
- Cleanup is registered before provisioning. It stops and joins the
  scenario's workers, then tears down every allocated tenant, including
  partially provisioned ones, each with its own fresh bounded context, and
  reports failures with `t.Error`. Teardown removes the tenant's rows, Flipt namespace and every
  `tenant:<key>:` Valkey key, which covers leaderboard and identity throttle
  keys.

Only scenarios whose fixed IDs land in `(tenant, …)`-keyed tables can overlap;
several mutable tables still have an `id`-only primary key. The scenario
canaries in `services/tadoku-api/e2e/scenario_canary_test.go` prove teardown
isolation, zero residue, worker scoping, seed confinement, deterministic
provisioning and cleanup after fatal, cancelled and failed setup.
