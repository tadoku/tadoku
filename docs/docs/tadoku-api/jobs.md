---
title: Jobs and worker
description: Typed background jobs, transactional publication, worker handler registration, retries and leases, retention, replay, branch development and contract-version migration.
sidebar_position: 5
---

# Jobs and worker

Read this when adding background work, changing a stored job contract, or
operating the worker. Job definitions live in `domain/jobs`, durable queue
operations in `features/jobqueue`, and the worker application with its handler
registrations in `app/worker`. All paths on this page are relative to
`services/tadoku-api/` unless stated.

## Ownership

| Owner | Responsibility |
| --- | --- |
| `domain/jobs` | Concrete versioned payloads, stable names and pure validation shared by producers and consumers. |
| Producing feature | Business writes and deciding which typed jobs must follow them. Returns jobs with its result. |
| API application | Caller authorization and the transaction that persists both the feature write and every returned job. |
| `features/jobqueue` | Enqueue, claim, lease renewal, fenced outcomes, retry scheduling, failed records, replay and retention. |
| `app/worker` | Typed handlers, registration, payload decoding, execution policies, dispatch and composition of features. |
| `cmd/tadoku-worker` | Dependency construction, configuration, application lifecycle, the private health/metrics listeners and the replay command. |

Features never import or call sibling features. In particular, a producing
feature has no queue dependency: the application coordinates it with
`jobqueue`. Queue persistence treats the versioned name as data and contains no
business payload catalogue or handler contract. The worker application's
registry is the source of executable versions, including replay eligibility.

The queue persists in the `jobs` table.

## Tenants and tenant overrides

Each job row stores its tenant in `jobs.tenant`, supplied by the database column
default when it is enqueued. Claim parses that persisted value into a validated
tenant key before dispatch. A handler runs with that job's tenant, and lease
renewals, completion, retry and failure transitions use the same key. Terminal
transitions keep their tenant during bounded shutdown cleanup.

A base deployment uses `jobqueue.AllTenantsExcept(component)`. It claims across
all tenants except those with a matching `(tenant, component)` row in
`tenant_overrides`. The same filter applies to supported and unsupported backlog
metrics, and prevents the base from failing exhausted jobs for a listed tenant.
An override for another component does not affect this worker.

A branch deployment uses `jobqueue.OnlyTenant(key)` with a parsed, non-production
tenant. It claims and reports only that tenant, regardless of overrides. Both
constructors validate their inputs, and `worker.NewApplication` rejects a zero
`Config.Scope`. `Scope.Context` applies the queue scope through the existing
transaction-local PostgreSQL executor; handlers receive the claimed job's tenant
separately. `WORKER_BRANCH` and `WORKER_COMPONENT` select the scope; see
[Separate job worker](./configuration.md#separate-job-worker).

Listing a tenant without running a branch worker pauses its work. Jobs already
claimed before the override appears finish on the base worker. Provisioning must
insert the override before starting the branch worker, stop that worker before
deleting the override, and let the base resume remaining work afterward.

## Define a message

`jobs.Job` is sealed to the predefined domain catalogue. Each concrete value
type supplies `Type() jobs.Type`, `Validate() error` and the package-private
marker. A name is constant for all values, including the zero value. Use value
types for registered handlers; a name must never depend on payload fields.

```go
// domain/jobs: a concrete payload carries its versioned name.
const LeaderboardInvalidateContestV1 Type = "leaderboard.invalidate_contest.v1"

type InvalidateContestLeaderboardV1 struct {
    ContestID uuid.UUID `json:"contest_id"`
}

func (InvalidateContestLeaderboardV1) Type() Type {
    return LeaderboardInvalidateContestV1
}
```

Name exported constants after the full persisted name in PascalCase, including
the version: `LeaderboardInvalidateContestV1` and
`LeaderboardInvalidateOfficialV1`. Keep the version in payload type names and
persisted strings too, so multiple contract versions can coexist.

The message contract includes its fields, validation, meaning and side effects.
Keep each published version stable. A different interpretation of an existing
field can require a new version even when its JSON shape does not change.

Payloads must not import feature types. Choose the smallest sufficient facts:

- Use an ID when the handler should read current state.
- Define an explicit snapshot when it needs facts as they were at publication.
  The producing feature copies its data into those fields.
- Use a pure shared domain value when a business concept really belongs to
  multiple features. Move the small value and its rules, not an entire feature
  model or service, into the shared domain.

Never embed a feature model in a stored payload. The native domain import
policy excludes application, feature, generated, infrastructure, storage and
`internal` dependencies. Keep provider objects and secrets out of jobs.

## Publish atomically

A producing feature returns its ordinary result and `Jobs []jobs.Job`. It
constructs the required business jobs; the application passes them through
without reconstructing payloads or deciding the invalidations itself.

```go
func (s *Service) Enqueue(ctx context.Context, batch ...jobs.Job) error
```

The application invokes that queue method in the business transaction:

```go
err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
    result, err := a.feature.Write(ctx, input)
    if err != nil {
        return err
    }

    return a.jobqueue.Enqueue(ctx, result.Jobs...)
})
```

This is a generic composition example; use the actual feature operation and
result in each application workflow. Every caller of a job-producing operation
must persist all returned jobs before committing. This applies to worker
handlers as well as HTTP operations if they compose a feature that returns
follow-up work.

`Enqueue` accepts both standalone calls and calls inside a transaction. An empty
batch succeeds without inserting rows. It validates and serializes the whole
batch before inserting any job. Nil jobs, invalid payloads and database errors
fail the operation. Without a transaction, earlier inserts remain persisted if
a later insert fails. Use a transaction when the whole batch must be atomic.

For jobs accompanying a business write, always enqueue inside that write's
transaction and propagate errors from the callback so both business rows and
any inserted jobs roll back. Do not enqueue after commit. Report success only
after commit succeeds; workers can see transactional jobs only after commit.
An enqueue error does not independently roll back a caller-owned transaction.
See [Transactions](./database.md#transactions) for wrong-pool and ended-context
handling and uncertain commit outcomes.

## Register typed handlers

The worker application in `app/worker/application.go` owns the registry.
`registrations` lists every executable job with its execution policy:

```go
func registrations(
    invalidateContest func(context.Context, jobs.InvalidateContestLeaderboardV1) error,
    invalidateOfficial func(context.Context, jobs.InvalidateOfficialLeaderboardV1) error,
) (*registry, error) {
    leaderboardPolicy := policy{
        Concurrency: 2,
        Timeout:     20 * time.Second,
        MaxAttempts: 5,
    }
    return newRegistry(
        handle(invalidateContest, leaderboardPolicy),
        handle(invalidateOfficial, leaderboardPolicy),
    )
}
```

It has two callers. `NewApplication` passes the application's handler methods,
which compose feature services:

```go
func (a *Application) InvalidateContestLeaderboard(ctx context.Context, job jobs.InvalidateContestLeaderboardV1) error {
    return a.leaderboard.InvalidateContest(ctx, job.ContestID)
}
```

`Replay` passes `ignoreJob[J]` for each parameter, so the replay command accepts
exactly the registered names without constructing feature services. A new
parameter on `registrations` fails to compile until both callers supply it.

`handle[J]` derives the persisted name from `J`, so a registration cannot pair a
name with the wrong payload type. For each claimed job it decodes exactly one
JSON object, rejects unknown fields and calls `Validate` before the handler.
Construction fails for duplicate names, nil functions, pointer payload types and
invalid policies: `Concurrency` must be 1–100, `Timeout` positive and
`MaxAttempts` 1–2,147,483,647. The registry is immutable after construction. It
drives claiming, dispatch, backlog metrics and replay eligibility, so the queue
needs no job-specific change when a handler is added.

`cmd/tadoku-worker` constructs and runs the application:

```go
scope, err := jobqueue.AllTenantsExcept(jobqueue.WorkerComponent)
if err != nil {
    return err
}

application, err := worker.NewApplication(queue, leaderboardService, worker.Config{
    Scope:           scope,
    Concurrency:     cfg.Concurrency,
    ShutdownTimeout: cfg.ShutdownTimeout,
    Logger:          logger,
    Metrics:         metrics,
})
if err != nil {
    return err
}
return application.Run(ctx)
```

The command loads and validates its settings before construction, and
`NewApplication` rejects a non-positive concurrency or shutdown timeout. See
[Separate job worker](./configuration.md#separate-job-worker) for the settings
and their defaults.

## Execution and failure guarantees

Delivery is at least once. A process can complete an external effect and crash
before recording success. Make effects idempotent, use a stable effect key when
needed, and tolerate reordering: there is no global FIFO guarantee.

### Claims and concurrency

The run loop claims work whenever a running job finishes, and otherwise every
500 milliseconds. Each pass visits the registered names in rotating order and
claims at most the free capacity: the smaller of the free process-wide slots
(`WORKER_CONCURRENCY`) and the name's free `policy.Concurrency` slots. A claim
takes the earliest due rows and skips rows locked by another claimer; there is
no local prefetch queue. The limits apply per process. Adding replicas
multiplies them; there is no distributed global concurrency limit.

### Leases and slow handlers

Each claim holds a 10-second lease with a random claim token, timed by
PostgreSQL's clock. While the handler runs, the worker renews the lease every
third of the lease until the handler's `policy.Timeout`, then cancels the
handler's context. The remaining lease lets the worker record the timeout when
the handler returns promptly. Cancellation cannot stop a Go function, so the
slot stays occupied until the handler returns.

A handler that ignores cancellation keeps running after renewal stops. Once its
lease expires, a later claim can run the same job again in parallel, and the
stale attempt's outcome is discarded because its token no longer matches.
Provider calls must have bounds and honor cancellation.

If a renewal fails or the token no longer matches, the worker cancels the
handler and records no outcome. The job is claimed again after its lease
expires, which `tadoku_worker_expired_leases_total` counts. Fencing prevents a
stale owner from changing queue state; it cannot reverse an external side
effect. Business scheduling and audit time use the normal `timex` conventions.

### Retries and failure codes

Every claim counts as an attempt. The outcome sets the job's next state and its
`last_error` code; the worker log line `job failed` for the same `job_id` holds
the error text.

| Outcome | Next state | `last_error` |
| --- | --- | --- |
| Handler returns `nil` | `completed` | None |
| Payload fails decoding or `Validate` | `failed`, without retry or handler call | `invalid_payload` |
| Handler exceeds `policy.Timeout` | `pending` for retry, or `failed` after the last attempt | `deadline_exceeded` |
| Handler returns another error, including cancellation at shutdown | `pending` for retry, or `failed` after the last attempt | `handler_error` |
| Lease expires during the last attempt | `failed` at the next claim | `lease_expired` |
| A `pending` job already at a lowered attempt limit | `failed` at the next claim | `attempts_exhausted` |

After attempt _n_ fails, the retry waits a random delay between half and all of
2<sup>_n_−1</sup> seconds, capped at one minute: 0.5–1 seconds after the first
attempt, 1–2 seconds after the second and 4–8 seconds after the fourth. With
the leaderboard policy of five attempts, a job that keeps failing becomes
`failed` within about 15 seconds plus handler time.

Names without a registered handler stay `pending` and unclaimed. They appear in
the unsupported backlog gauges, and the worker logs
`unsupported job types remain outstanding` at startup and whenever their count
changes to a non-zero value. Failed records are kept for inspection and replay;
see [Inspect and replay retained work](#inspect-and-replay-retained-work).

### Shutdown

On `SIGTERM` the worker stops claiming and reports not ready. It waits up to
`WORKER_SHUTDOWN_TIMEOUT` for running handlers, then cancels the rest and waits
for them to return before closing its database and Valkey clients. A handler
that does not honor cancellation delays exit.

## Leaderboard freshness and worker health

PostgreSQL is the source of truth, but unfiltered global, yearly and contest
leaderboards use warm Valkey cache keys derived from the request or persisted
job tenant, in the form `tenant:<name>/<id>:leaderboard:…`. Invalidation changes
only that tenant's keys. Private development databases sharing `tadoku/prod`
retain their route prefix as described in
[Valkey configuration](./configuration.md#valkey-and-leaderboard-caches).
The API does not check worker readiness
or the queue before serving a warm key. It can therefore serve an outdated
leaderboard while the worker is down, while that key's invalidation job is
pending or running, and after the job has failed terminally. A failed
invalidation leaves the key stale until another write affecting the same
contest or year publishes a new invalidation, or an operator replays the failed
job. There is no automatic age limit on a warm key and no startup
reconciliation of all leaderboard keys.

On a cache miss, the API rebuilds from PostgreSQL behind a generation check so
an older rebuild cannot overwrite a completed invalidation. Filtered reads use
PostgreSQL, and Valkey errors fall back to PostgreSQL. Those paths do not make
an already warm key fresh while its invalidation remains unprocessed.

The worker's private `/livez` answers while the process serves. `/readyz`
succeeds only when the last claim pass succeeded for every registered name and
PostgreSQL answers a ping. It does not check Valkey, assert that any
leaderboard is current, or mean that every job succeeded.
Watch the [worker metrics](./configuration.md#worker-metrics): pending jobs and
oldest due age show delayed invalidations, failed jobs show terminal failures,
and unsupported-job gauges show work this worker cannot claim. The backlog
gauges refresh at startup and every fifteen seconds. Investigate any failed
leaderboard invalidation in `jobs`. Inspect its `last_error` and repair the
cause before using the
[replay command](#inspect-and-replay-retained-work). Verify the replayed job
completes and a subsequent leaderboard read reflects PostgreSQL. A later write
can also publish another invalidation, but do not depend on users making one to
recover a stale board.

## Exercise job effects in end-to-end tests

Prove a job's effect through the real publication and execution path. In an
HTTP [user journey](./user-journeys.md):

1. Send the authenticated producer request through the production router. Its
   transaction commits the business rows and the jobs together.
2. Add a `run_worker` step, or `run_branch_worker` with `forTenant` for a test
   tenant. The step runs `worker.NewApplication` with the real queue and feature
   services against the journey's PostgreSQL and Valkey fixtures. It waits up to
   ten seconds for every eligible job to complete, fails immediately on a
   terminal failure, and stops the worker before the next step.
3. Send the follow-up request and compare its golden response to prove the
   user-visible effect. Verify effects without a read endpoint through the real
   provider or a verify step. Queue completion alone does not prove the effect.

`Run` is the only execution entry point; there is no synchronous drain API.
Calling a handler directly or mocking it skips the publication, claiming and
dispatch path this test must cover.

## Run jobs on your development branch

dev-cli runs your branch's worker beside its API on the shared development
cluster; see [Development environment](../develop/environment.md) for setup.

```sh
dev up --owner alice --service tadoku-api
dev status --owner alice
dev logs --owner alice tadoku-worker
```

- Selecting the API or the worker starts both, scoped to tenant
  `tadoku/<route>`; `dev status` prints the route. A `tadoku-worker` override
  keeps the base worker away from that tenant while your worker runs. A
  frontend-only branch has no worker overlay, and the base worker processes its
  jobs.
- `dev logs` follows new log lines. A failed attempt logs `job failed` with the
  job ID, name, attempt, failure code, error and tenant.
- The worker has no route. Forward its private ports to check readiness and
  metrics; a branch worker's backlog gauges count only its tenant.

```sh
pod=$(kubectl --context homelab-talos-dev -n tdk-dev-tadoku-api get pod \
  -l dev-cli.io/owner=alice,dev-cli.io/service=tadoku-worker -o name)
kubectl --context homelab-talos-dev -n tdk-dev-tadoku-api port-forward "$pod" 8000:8000 9090:9090
curl -s -o /dev/null -w '%{http_code}\n' localhost:8000/readyz
curl -s localhost:9090/metrics | grep '^tadoku_worker_'
```

To inspect rows, connect psql to database `tadoku` as the non-owner runtime
role. The worker reads its credentials from Secret
`tadoku.tadoku-dev-db.credentials.postgresql.acid.zalan.do` in
`tdk-dev-tadoku-api`; forward a local port to the primary database pod, selected
by `cluster-name=tadoku-dev-db,spilo-role=master` in `tdk-dev-data`. A
migration branch uses database `tadoku-<route>`. Run the
[inspection queries](#inspect-and-replay-retained-work) with your own tenant.

To watch failure handling, insert jobs in your own tenant. The first fails on
its first attempt with `invalid_payload` because of the unknown field. The
second stays `pending` and raises `tadoku_worker_unsupported_pending_jobs`
because no handler registers it.

```sql
begin;
select set_config('tadoku.tenant', 'tadoku/<route>', true);
insert into jobs (task_type, payload) values
  ('leaderboard.invalidate_official.v1', '{"year": 2026, "source": "test"}'),
  ('leaderboard.invalidate_official.v2', '{"year": 2026}');
commit;
```

The replay command cannot replay branch jobs: it runs as `tadoku/prod`. Test
replay with `bazel test //services/tadoku-api/cmd/tadoku-worker:tadoku-worker_test`
and a disposable database, as described in [Testing](./testing.md).

`dev down` stops your worker and clears its override, so the base worker can
finish your remaining jobs before tenant teardown deletes the tenant's rows,
including its jobs.

## Successful-job retention

The worker cleans up successful history at startup and then hourly. Each run
deletes batches of 100 rows for at most five seconds; a larger backlog continues
at the next run. A base worker cleans every tenant, including tenants with an
override. A branch worker cleans only its own tenant.
`jobqueue.Service.CleanupCompleted(ctx, limit)` owns the retention rule; callers
do not supply an arbitrary cutoff. The service passes the business clock to a
single repository statement, which subtracts three calendar months in UTC.

A row is eligible only when `state = 'completed'` and `completed_at` is strictly
earlier than that cutoff. Exact-boundary completions remain. UTC time of day is
preserved; the day clamps to the last valid day of the target month. For
example, May 31 at 10:30 UTC has a February 28 cutoff at 10:30 UTC, or February
29 in a leap year. This is not a fixed number of days and does not depend on the
database session's timezone.

Successful replay rows expire under the same rule when no retained child
references them. The failed original stays indefinitely. The self-referencing
foreign key and the no-child cleanup condition preserve replay lineage; a row
still needed by a retained descendant is not removed. Automatic cleanup never
deletes pending, running or failed records, regardless of age. A successful
replay does not make its failed original eligible for cleanup.

Record important completion evidence before successful history expires. A
missing completed row or aged-out successful replay is not proof that an effect
never ran. Failure inspection and version-retirement decisions still account
for the retained failed originals.

## Add a job

1. Define the value type, constant versioned name and validation in
   `domain/jobs`. Decide whether it represents current-state IDs or a snapshot,
   and document idempotency, cancellation and replay behavior.
2. Enumerate meaningful failure modes before implementation. Add or adapt a
   real workflow or repository check for the behavior that could regress.
3. Implement the typed application method and add its `handle` entry, with
   concurrency, timeout and attempt limits, to `registrations`. Pass the method
   from `NewApplication` and `ignoreJob` from `Replay`.
4. Return the typed job from each producing feature. Audit every application
   caller and enqueue all returned jobs in the same transaction.
5. Release the worker that registers the new name before the API that publishes
   it. Jobs published earlier stay `pending` as unsupported work until a worker
   registers them.

Any new schema or data migration is a standalone PR, deployed independently
before dependent runtime changes. Retain repeatable verification evidence:
fixtures, command, tested revision, output and any simulated provider boundary.
See [Testing](./testing.md).

## Migrate v1 to v2

A versioned name and its payload form a protocol between independently deployed
API and worker images, and stored rows outlive both. Use a new version when the
payload shape, validation or meaning changes. Run both versions while v1 drains;
never rewrite stored rows. These facts drive the sequence:

- A worker claims only names it registers. Rows with other names stay `pending`
  and appear in the unsupported gauges.
- Decoding is strict. A v1 handler fails a payload with an extra field as
  `invalid_payload`.
- Replay copies the failed row's name and payload unchanged, and only for a
  name the replaying binary registers.
- Failed rows stay in `jobs` indefinitely, in every tenant.

### Define and register both versions

Add a distinct v2 value, name and validation, and register its handler beside
v1 in `registrations`. Keep the v1 payload and handler intact. Define whether
v1 and v2 effects can overlap safely and how idempotency works across them.
Verify old persisted v1 payloads and new v2 payloads against their handlers.

**Gate:** both versions can be processed, replay behavior is understood, and
supported producer and worker rollback versions have been identified.

### Deploy consumers, then switch producers

Release the worker that registers v1 and v2 before any producer emits v2, and
verify the running revision processes v2 through a controlled workflow. Then
release the producers that emit v2. Producer replicas still being replaced may
write v1 meanwhile, which is safe while v1 is registered. Do not enqueue both
versions for one business action.

**Gate:** all intended producers emit v2, the worker processes both versions,
and the unsupported pending and running gauges are zero.

### Drain v1 and decide on its failures

Count v1 rows in every tenant with the
[all-tenants query](#inspect-and-replay-retained-work); branch tenants share
the queue. Pending rows include delayed retries, and running rows hold leases;
let the worker drain them. For each failed v1 row, either repair the cause and
replay it while v1 is registered, or record why it needs no replay. Replays of
v1 are v1 rows too, and a successful replay leaves its failed original in place.

**Gate:** no pending or running v1 rows remain in any tenant, and every failed
v1 row has a recorded replay or no-replay decision.

### Retire v1

Remove the v1 registration only when all of these conditions hold:

- No current producer can emit v1.
- No supported producer rollback can emit v1.
- Every supported worker rollback can execute every version that may be queued.
- The previous gates passed.

Delete the v1 `handle` entry and its `registrations` parameter; remove the
payload type once nothing references it. Failed v1 rows then remain as history:

- They can no longer be replayed; the replay command answers
  `job is not failed or supported`. Replaying one later requires releasing a
  worker that registers v1 again. Never redirect it to v2.
- They count in `tadoku_worker_unsupported_failed_jobs`, and the worker logs
  `unsupported job types remain outstanding` at startup while any remain. Use
  the unsupported pending and running gauges to detect stranded new work.
- The queue has no retirement flag, archive or conversion command. Any new
  archive or conversion workflow requires its own review, preserves history and
  replay lineage, and ships any SQL as a standalone migration.

### Rollback and optional fields

After producers start writing v2, reverting every consumer to v1 strands v2
rows even if producers have already reverted. Keep a compatible consumer for
every version that remains queued or may still be published. A producer rollback
to v1 is safe only while v1 support remains. These are compatibility conditions,
not permission to perform a rollback.

Strict decoding rejects unknown fields. Adding an apparently optional JSON
field can therefore break an older worker. Prefer a new version. A same-version
extension requires readers that accept both forms to be deployed everywhere
first, with old/new fixtures and the supported rollback direction verified.
Do not silently relax decoding for the entire catalogue.

If draining both versions is insufficient, design a separate reviewed data
conversion with deterministic mapping, running-claim handling, idempotency,
audit lineage and reconciliation. Never bulk relabel stored v1 rows as v2.
A SQL data conversion remains a standalone migration PR and deployment.

## Inspect and replay retained work

`jobs` has row-level security. Connect as the runtime role and set the scope
inside the transaction; without it, the queries return no rows. Set
`tadoku.tenant` to inspect one tenant:

```sql
begin;
select set_config('tadoku.tenant', 'tadoku/prod', true);

select task_type, state, count(*)
from jobs
group by task_type, state
order by task_type, state;

select id, state, attempts, next_attempt_at, lease_expires_at,
       failed_at, last_error, replay_of_id
from jobs
where task_type = 'leaderboard.invalidate_contest.v1'
  and state <> 'completed'
order by id;
rollback;
```

A version retirement must count every tenant. The jobs policy admits the
all-tenants scope:

```sql
begin;
select set_config('tadoku.all_tenants', 'on', true);

select tenant, task_type, state, count(*)
from jobs
where task_type like 'leaderboard.invalidate_contest.%'
group by tenant, task_type, state
order by tenant, task_type, state;
rollback;
```

Delayed retries are `pending` with a future `next_attempt_at`. Running rows
have a `lease_expires_at`.

After repairing the cause, replay a failed job with the worker binary's `replay`
subcommand. It needs only the worker's `WORKER_POSTGRES_*` settings, so it can
run inside the worker container:

```sh
tadoku-worker replay --id 123 --actor operator-name --reason 'cause repaired'
```

- It replays only `tadoku/prod` jobs. For a job in another tenant, a job that is
  not `failed`, or a name the binary does not register, it exits with
  `job is not failed or supported`.
- It inserts a new `pending` job with the same name and payload, linked through
  `replay_of_id`, and records the trimmed actor and reason. It logs
  `job replayed` with `source_job_id` and the new `job_id`. The failed original
  stays unchanged.
- Verify the new job reaches `completed`. An invalid payload remains invalid on
  replay; repair requires its own reviewed procedure, not repeated replay or a
  name change.

These inspection and replay steps do not alter the schema. Schema and SQL data
migrations follow [Database and migrations](./database.md#migrations), with a
separate PR and deployment gate before dependent code.
