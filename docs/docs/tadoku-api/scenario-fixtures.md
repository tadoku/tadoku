---
title: Scenario fixtures
description: How a Tadoku API E2E scenario owns private test tenants, business clock and IDs, seeds them under row-level security, signs its cast, scopes its workers and tears everything down, and which scenario tests may run in parallel.
sidebar_position: 6.6
---

# Scenario fixtures

Read this when you write or convert an E2E scenario that owns its own tenants
instead of resetting the whole database, or make an E2E test parallel. Golden-case layout is in
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
  every row-level security policy is exactly the `tadoku.tenant` binding, and
  on their sequences. Row-level security confines every statement, including
  an unscoped `update` or `delete`, to the role's tenant. Writes to the tenant
  registry, tenant overrides, `jobs` (whose policy also honours
  `tadoku.all_tenants`), shared reference tables such as `languages` and
  migration metadata fail with `permission denied`, so a seed cannot register
  tenants, seed jobs or read shared reference rows. Scenario relationship rows
  omit `tenant`; the role supplies it.
- A scenario seed file must not change session state: the fixture refuses a
  file containing `set_config(`, a `set` or `reset` statement, or a reference
  to `tadoku.tenant` or `tadoku.all_tenants`, outside comments, and names the
  file. Rows take their tenant from the column defaults. `SeedTenant` also
  fails if the transaction ends bound to another tenant or to all tenants.
- `serve(t, role, member, request)` runs the request through the production
  router with a freshly signed token for that cast member and tenant. Tokens
  carry the checked-in cast's subjects, traits and timing relative to
  `jwt.TimeFunc`, signed by a key generated per test process; the local JWKS
  serves its public half next to the checked-in keys.
- `startWorker` and `runWorker` run workers scoped with `jobqueue.OnlyTenant`
  to one role's tenant; `runWorker` completes that tenant's jobs, including
  jobs they enqueue.
- `serveWith` does the same through another handler, such as a private router
  built on the shared pools with `newTestRouterWithLogger`.
- Every request and worker context of a scenario carries its own business
  clock and UUID source. The clock starts at `fixtureInstant`; call
  `s.clock.Set` or `s.clock.Advance` to move that scenario's time alone. IDs
  come from a stream seeded by the scenario's first tenant, so overlapping
  scenarios draw different IDs. JWT verification stays at `fixtureInstant`,
  which `TestMain` fixes once.
- Cleanup is registered before provisioning. It stops and joins the
  scenario's workers, then tears down every allocated tenant, including
  partially provisioned ones, each with its own fresh bounded context, and
  reports failures with `t.Error`. Teardown removes the tenant's rows, Flipt namespace and every
  `tenant:<key>:` Valkey key, which covers leaderboard and identity throttle
  keys.

The scenario canaries in `services/tadoku-api/e2e/scenario_canary_test.go`
prove teardown isolation, zero residue, worker scoping, seed confinement,
deterministic provisioning, cleanup after fatal, cancelled and failed setup,
and that overlapping scenarios write their own clock's timestamps and IDs
through requests and a worker.

## Parallel scenarios

A top-level E2E test may call `t.Parallel` only when all of these hold:

- It is a scenario test: it calls `newScenario` after `t.Parallel`, and sends
  every request and runs every worker through that scenario.
- It reaches no process-wide or shared mutation: no `reset`, `resetIdentity`
  or `db.Reset`, no Flipt `Reset` or `SetUnavailable`, no `timex.TheWorld`,
  `timex.Advance`, `uuid.SetRand` or `jwt.TimeFunc` write, no `t.Setenv`, no
  schema, trigger or sequence fault, and no write to shared reference data or
  the canonical tenant.
- None of its fixed or generated IDs land in a table whose primary key is
  `id` alone: `contests`, `contest_registrations`, `logs`,
  `scoring_rule_sets`, `scoring_rules`, `posts`, `posts_content`, `pages`,
  `pages_content`, `announcements` and `moderation_audit_log`. Two tenants
  writing the same ID there collide.
- It does not depend on process-wide fixture state that a serial test may
  leave behind, such as the Flipt fixture's unavailable mode, which lasts
  until the next reset.

Every other test stays a serial top-level test. Go starts parallel top-level
tests only after every serial top-level test has returned, so serial resets,
faults and canonical-only tests never overlap a parallel scenario. That barrier
is the only exclusion: a serial subtest under a parallel parent, or a mutex
that only fault tests take, does not keep ordinary scenarios out.

Keep every step of a scenario sequential. Parallel subtests are only for
scenarios that deliberately overlap and wait for each other with
`meetScenario`; that needs `-test.parallel` of at least 2.

The E2E target's `args` set `-test.parallel=2`, for normal and race builds
alike. With three normal shards, up to six scenarios run at once. Raise the
bound only with measured memory, CPU and connection headroom.
