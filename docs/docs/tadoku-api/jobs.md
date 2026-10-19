---
title: Jobs and worker
description: Typed background jobs, transactional publication, worker handler registration, execution guarantees, retries and leases, and end-to-end tests of job effects.
sidebar_position: 5
---

# Jobs and worker

Read this when adding background work or changing how a job is defined,
published or handled. [Operate jobs and the worker](./jobs-operations.md)
covers worker health, branch runs, retention and replay, and
[Migrate a job contract from v1 to v2](./jobs-versioning.md) covers changing a
stored job contract. Job definitions live in `domain/jobs`, durable queue
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
[Job worker configuration](./worker-configuration.md).

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

### Email message contract

`jobs.SendEmailV1` has the persisted name `email.send.v1` and carries exactly
`{"email_id":"<uuid>"}`. A zero UUID is invalid. The tenant comes from the queue
row and handler context, never from the payload. Recipients, rendered content,
parameters and credentials stay outside this job. The handler reads current
email state and must tolerate retries without resending terminal deliveries.
The payload is declared but has no executable registration yet; no producer is
wired. Sending and retention remain unimplemented.

`domain/mail.Message` carries template, recipient and string parameters between
features through application composition. `email.Writer.Write` requires the
business transaction and returns `WriteResult.Jobs`; the application enqueues
all jobs before committing, together with the identity write.

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
[Job worker configuration](./worker-configuration.md) for the settings
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
see [Inspect and replay retained work](./jobs-operations.md#inspect-and-replay-retained-work).

### Shutdown

On `SIGTERM` the worker stops claiming and reports not ready. It waits up to
`WORKER_SHUTDOWN_TIMEOUT` for running handlers, then cancels the rest and waits
for them to return before closing its database and Valkey clients. A handler
that does not honor cancellation delays exit.

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
