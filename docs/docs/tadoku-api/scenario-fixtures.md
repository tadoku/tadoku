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
own tenants instead of a whole-database reset. A scenario never calls `reset`.
Golden cases and journeys use `reset` unless they run as scenarios; see
[Golden cases and journeys as scenarios](#golden-cases-and-journeys-as-scenarios).

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

## Golden cases and journeys as scenarios

`services/tadoku-api/e2e/fixture_scenario_test.go` runs a checked-in golden
case or journey in its own scenario. The fixture layout, seed resolution and
step table stay as [HTTP end-to-end tests](./http-e2e.md) and
[User journeys](./user-journeys.md) describe them. A case switches by calling
`runScenarioCase` or `runScenarioCaseAt` instead of `runCase` or `runCaseAt`,
and a journey by calling `runScenarioJourney` instead of `runJourney`,
`runJourneyWithHandler` or `runJourneyWithSetup`. Only the rules for
[parallel scenarios](#parallel-scenarios) decide whether it may switch.

- **Tenants.** The `primary` role stands in for `tadoku/prod`, which fixture
  SQL defaults to. Every test tenant that a case's tokens, relationships or
  cookies name needs a `fixtureTenant` mapping it to its own role, such as
  `alphaAsOther` for `e2e/alpha-0000000a`. An unmapped fixture tenant fails the
  case.
- **Seeds.** The resolved `setup.sql`, the account cast and the resolved
  `finalize.sql` run in `primary` with the seed rules above; every other role
  gets the cast alone. Each relationship row is seeded into the role standing
  in for its `tenant`. A seed that names its tenant fails under row-level
  security, so let the column defaults supply it.
- **Requests.** Each fixture token's claims must equal the cast member's claims
  for the token's tenant exactly; the runner signs the same claims for the
  stand-in tenant. Session cookies named for a fixture tenant (`tadoku_session`
  for `tadoku/prod`, otherwise `tadoku_session_` plus the key with `/` replaced
  by `_`) are renamed for the stand-in. A request that names a fixture tenant
  anywhere else, in its path, another header or its body, fails.
- **Comparison.** Before comparing, the runner replaces each scenario tenant's
  generated route, such as `primary-1a2b3c4d` in `e2e/primary-1a2b3c4d`, with
  `{role}`. That covers the tenant key, its session cookie name and its
  recovery host, so goldens name `e2e/{primary}` and `e2e/{other}`. Every
  other value, `tadoku/prod` and IDs included, is compared unchanged, and goldens
  are never rewritten except by `-update-goldens`.
- **Steps.** Verify queries run in a read-only transaction bound to `primary`
  under the restricted application role. Valkey snapshots scan only
  `tenant:<key>:*` for the scenario's tenants. `run_worker` runs a
  `jobqueue.OnlyTenant` worker for each scenario tenant until its jobs,
  including jobs they enqueue, complete; `run_branch_worker` runs only the
  stand-in for `forTenant`.
- **Time and IDs.** A case runs at its business time with its implementation's
  UUID stream, and a journey with its `ids` reader or the fixed seed, so goldens
  match the serial runners'. Identical IDs in different tenants are safe only in
  tables keyed by tenant.
- **Parallelism.** `runScenarioCase` calls `t.Parallel` on the case's subtest
  and `runScenarioJourney` on the journey's test. A table whose cases all run as
  scenarios also calls `t.Parallel` on its top-level test. A table that keeps
  serial cases stays a serial top-level test; its scenario cases start after
  its serial cases have returned. Journeys with an `afterReset` fault and cases
  with process-wide fixtures stay on the serial runners. So does a case that
  tests its own credential: the runner signs a fresh token for any token whose
  claims equal the cast's, which would replace a bad signature or unknown key
  under test, and refuses every other credential. A journey whose setup
  touches only its own tenants, such as seeding its tenant's leaderboard cache,
  passes it to `runScenarioJourneyWithSetup`, which runs it after seeding.

`TestFixtureScenarioCasesStayInTheirTenants` in
`services/tadoku-api/e2e/fixture_scenario_canary_test.go` runs the same golden
cases with identical seeds, clocks and UUID streams in two overlapping
scenarios. It proves that each tenant's request writes, cache keys and verify
reads stay its own, and that one scenario's worker, including the job its job
enqueues, leaves the other's job pending.

## Parallel scenarios

A top-level E2E test may call `t.Parallel` only when all of these hold:

- It is a scenario test: it calls `newScenario` after `t.Parallel`, and sends
  every request and runs every worker through that scenario. The scenario
  golden and journey runners do this for their cases.
- It reaches no process-wide or shared mutation: no `reset`, `resetIdentity`
  or `db.Reset`, no Flipt `Reset` or `SetUnavailable`, no `timex.TheWorld`,
  `timex.Advance`, `uuid.SetRand` or `jwt.TimeFunc` write, no `t.Setenv`, no
  schema, trigger or sequence fault, and no write to shared reference data or
  the canonical tenant.
- It does not depend on process-wide fixture state that a serial test may
  leave behind, such as the Flipt fixture's unavailable mode, which lasts
  until the next reset.

Every other test stays a serial top-level test. Go starts parallel top-level
tests only after every serial top-level test has returned, so serial resets,
faults and canonical-only tests never overlap a parallel scenario. That barrier
is the only exclusion: a serial subtest under a parallel parent, or a mutex
that only fault tests take, does not keep ordinary scenarios out. Scenarios do
not reset, so a serial test that writes shared reference data, such as
`languages`, restores it before it returns: it calls `resetAfter`, or removes
the rows it added in cleanup.

Keep every step of a scenario sequential. Parallel subtests are only for
scenarios that deliberately overlap and wait for each other with
`meetScenario`; that needs `-test.parallel` of at least 2.

The E2E target's `args` set `-test.parallel=2`, for normal and race builds
alike. With three normal shards, up to six scenarios run at once. Raise the
bound only with measured memory, CPU and connection headroom.
