---
title: Jobs and worker
description: Typed background jobs, atomic publication, worker registration, execution guarantees, replay and safe contract-version migration.
sidebar_position: 5
---

# Jobs and worker

Read this when adding background work, changing a stored job contract, or
operating the worker. Job definitions live in `domain/jobs`, durable queue
operations in `features/jobqueue`, and handlers and execution in `app/worker`.
All paths on this page are relative to `services/tadoku-api/` unless stated.

## Ownership

| Owner | Responsibility |
| --- | --- |
| `domain/jobs` | Concrete versioned payloads, stable names and pure validation shared by producers and consumers. |
| Producing feature | Business writes and deciding which typed jobs must follow them. Returns jobs with its result. |
| API application | Caller authorization and the transaction that persists both the feature write and every returned job. |
| `features/jobqueue` | Enqueue, claim, lease renewal, fenced outcomes, retry scheduling, failed records, replay and retention. |
| `app/worker` | Typed handlers, registration, payload decoding, execution policies, dispatch and composition of features. |
| `cmd/tadoku-worker` | Dependency construction, configuration, application lifecycle and private health/metrics listener. |

Features never import or call sibling features. In particular, a producing
feature has no queue dependency: the application coordinates it with
`jobqueue`. Queue persistence treats the versioned name as data and contains no
business payload catalogue or handler contract. The application registry is
the source of executable versions, including replay eligibility.

The queue persists in the `jobs` table. Deploy the standalone table-rename
migration before any runtime that queries this name. Persisted message names,
payloads and replay lineage remain unchanged; this schema change does not
introduce a new message version. See [Table migration](#table-migration).

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

`Enqueue` requires an active transaction. An empty batch inside that transaction
succeeds without inserting rows. It validates and serializes the concrete
values and persists them through the transaction-bound executor; nil jobs,
invalid payloads and database errors fail the operation. Return errors from the
transaction callback so both business rows and any inserted queue rows roll
back. Do not enqueue after commit. Report success only after commit succeeds;
workers can see jobs only after commit.

An enqueue error does not independently roll back a caller-owned transaction:
the caller must propagate it. See [Transactions](./database.md#transactions) for
wrong-pool and ended-context handling and uncertain commit outcomes.

## Register typed handlers

A handler is an application method with a concrete job parameter:

```go
func (a *Application) InvalidateContestLeaderboard(
    ctx context.Context,
    job jobs.InvalidateContestLeaderboardV1,
) error {
    return a.leaderboard.InvalidateContest(ctx, job.ContestID)
}
```

The private generic adapter belongs to `app/worker`:

```go
func handle[J jobs.Job](
    fn func(context.Context, J) error,
    policy Policy,
) registration
```

Register handlers during application construction. Go infers the message type
from the method, so the registration cannot pair a separate name string with
the wrong payload type.

```go
handle(a.InvalidateContestLeaderboard, Policy{
    Concurrency: 2,
    Timeout:     20 * time.Second,
    MaxAttempts: 5,
})
```

The registry derives the name from that type, decodes exactly one payload,
rejects unknown fields and validates before calling the method. Duplicate
names, nil functions, pointer job registrations and invalid policies fail
construction. Registration is immutable before `Run`. The same registry drives
claim eligibility and dispatch; an old worker does not claim versions it
cannot execute. Queue operations need no job-specific edits when a handler is
added.

Construct and run the application from the entry point:

```go
app, err := worker.NewApplication(queue, leaderboard, worker.Config{
    Concurrency: 4,
})
if err != nil {
    return err
}
return app.Run(ctx)
```

`Config` also accepts `ShutdownTimeout`, `Logger` and `Metrics`; startup can
provide the process logger and metric registry. Zero-value concurrency uses
four slots and zero-value shutdown timeout uses fifteen seconds. Handler
`Policy` values must be positive. Keep registration and handler signatures in
the worker application. Startup constructs the application and its external
resources.

## Execution and failure guarantees

Delivery is at least once. A process can complete an external effect and crash
before recording success. Make effects idempotent, use a stable effect key when
needed, and tolerate reordering: there is no global FIFO guarantee.

Concurrency limits apply per process, with both a global cap and a cap for each
registered type. Claims consume only free capacity and do not build a local
prefetch queue. A timeout cancels the context; it cannot stop a Go goroutine.
Keep the slot occupied until the handler actually returns. Provider calls must
have bounds and honor cancellation. Adding replicas multiplies the process
caps; it does not create a distributed global concurrency limit.

Claims and renewals use PostgreSQL wall time. Queue acknowledgments require a
matching, unexpired claim token. Losing a lease cancels the handler, and expired
claims can be recovered. Fencing prevents a stale owner from changing queue
state; it cannot reverse an external side effect. Business scheduling and audit
time use the normal `timex` conventions.

Invalid payloads for registered versions fail without invoking business code.
Ordinary errors retry with bounded backoff and an attempt limit; permanent and
exhausted failures remain inspectable. Unknown versions stay unclaimed and
visible in unsupported backlog reporting. Inspect due age, failures, attempts,
expired leases and in-flight work without using job IDs as metric labels.

Leaderboard cache readiness is permission for the API to serve cached results.
The separate worker maintains a short-lived `leaderboard:ready` key in Valkey,
scoped by the configured cache prefix. If the marker is absent, expired or
unreadable, API reads use PostgreSQL. This replaces the embedded worker's
process-local readiness flag as the separate worker is introduced; the staged
cutover temporarily requires both signals before retaining the shared marker
alone. It remains necessary because API and worker processes can restart or
roll out independently. Cache readiness is separate from Kubernetes Pod
readiness and failure reporting. Unsupported outstanding work blocks cache
readiness. A known failed leaderboard job remains visible for replay, but a
successful full cache reconciliation can restore readiness; retained failure
history alone does not permanently disable the cache.

Replay inserts a new record linked to the failed original and records the
operator and reason. It preserves the original version and payload. Registry
support is required before replay; neither replay nor package renaming upgrades
a message. Keep the failed original inspectable.

## Successful-job retention

The worker automatically cleans up successful history in bounded batches.
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

## Table migration

The standalone rename migration changes `async_outbox` to `jobs`, including
its sequence, constraints and indexes, without rewriting the already merged
migration that created the queue. It preserves rows, live claims, replay links
and sequence state. There is no old-name compatibility view.

Before deploying the rename, positively verify that no producer, worker or
manual tool still requires the old table name. Current supported mainline
application code does not use that queue, but an older unmerged worker release
would be incompatible. If inventory finds an old-name caller, stop the migration
deployment gate and resolve that compatibility requirement before proceeding.
Read-only inspection of declared GitOps workloads alone does not prove that no
external client exists.

Merge and deploy the migration independently, then base dependent runtime on the
updated `main` before merging or deploying it. Review stacks and isolated
migration fixtures are not evidence of live deployment. Operator SQL on this
page and in linked runbooks requires the migrated `jobs` schema. This schema
sequence is separate from the [v1-to-v2 message rollout](#migrate-v1-to-v2).

## Add a job

1. Define the value type, constant versioned name and validation in
   `domain/jobs`. Decide whether it represents current-state IDs or a snapshot,
   and document idempotency, cancellation and replay behavior.
2. Enumerate meaningful failure modes before implementation. Add or adapt a
   real workflow or repository check for the behavior that could regress.
3. Implement the typed application method and register it with concurrency,
   timeout and attempt limits. Compose feature services in the application.
4. Return the typed job from each producing feature. Audit every application
   caller and enqueue all returned jobs in the same transaction.
5. Deploy a compatible consumer before enabling publication. Verify processing
   and unsupported-backlog visibility, then switch producers.

Any new schema or data migration is a standalone PR, deployed independently
before dependent runtime changes. Retain repeatable verification evidence:
fixtures, command, tested revision, output and any simulated provider boundary.
See [Testing](./testing.md).

## Migrate v1 to v2

Use this sequence when payload shape, validation or meaning breaks the existing
contract. The normal approach runs both versions while v1 drains; it does not
rewrite the queue.

### Define and register both versions

Add a distinct v2 value and stable name. Keep the v1 payload and handler intact.
Define whether v1 and v2 effects can overlap safely and how idempotency works
across them. Register both versions in the worker application and verify old
persisted payloads and new payloads against the intended handlers.

**Gate:** both versions can be processed, replay behavior is understood, and
supported producer and worker rollback versions have been identified.

### Deploy consumers, then switch producers

First ensure every worker that can publish shared cache readiness uses the
fail-closed readiness protocol: unsupported outstanding work prevents it from
publishing readiness. Unsupported pending, running and failed records must be
visible. An older binary that ignores unknown work can otherwise mark stale
cache data ready while a compatible worker is still handling it.

Deploy v1+v2 support to every serving worker before any producer emits v2;
verify the actual revisions, including all shared-readiness publishers. Verify
v2 processing through a controlled workflow. Each worker claims only names in
its own registry, leaving unknown rows for compatible workers; that claim rule
alone is insufficient to protect shared cache readiness.

After the consumer gate passes, switch producers to v2. Old producer replicas
may continue writing v1 while the rollout completes. Do not enqueue both
versions for one business action.

**Gate:** all intended producers emit v2, compatible consumers process both,
and unsupported backlog is understood and monitored.

### Account for remaining v1 work

Inspect every non-completed v1 record, including delayed retries, running jobs
with leases and failed records that could be replayed. Linked replay records
also retain v1. Drain active work and repair/replay failures with the v1 handler,
or separately design an explicit archive/disposition workflow. The current
queue has no retirement flag or conversion command. Successful replay leaves
the original failed v1 row intact; after unregistering v1 that retained row
would be unsupported and block cache readiness. Keep the v1 handler while any
such row remains. A future archive/disposition must remove unsupported queue
rows while preserving required history and replay lineage, with separate review
and any required standalone data-migration deployment. An operator note alone
does not change queue eligibility. Completed history remains available for
inspection without requiring a handler.

**Gate:** no pending, running or retryable v1 work requires execution. Keep
registration while failed v1 records remain in the queue; handler removal must
wait for a separately implemented and verified archival/disposition process if
such rows exist.

### Retire v1 after compatibility gates pass

Remove v1 registration only when all of these conditions hold:

- No current producer can emit v1.
- No supported producer rollback can emit v1, unless a v1 consumer remains.
- No pending, running, delayed or failed v1 rows remain in the queue. Replayed
  failures still count because their original records are retained.
- Every supported worker rollback can execute every version that may be queued.
- The retention and inspection policy for historical v1 records is preserved.

Verify unsupported backlog before and after removing the registration. A
future replay request for a retired version needs an explicitly compatible
consumer and reviewed disposition; do not redirect it to v2 automatically.

### Rollback and optional fields

After producers start writing v2, reverting every consumer to v1 strands v2
rows even if producers have already reverted. Every worker rollback must retain v1+v2 support and the fail-closed readiness
protocol while v2 work exists. Merely leaving one compatible consumer running
does not protect a shared readiness marker from an incompatible publisher. A producer rollback to v1 is safe
only while v1 support remains. These are compatibility conditions, not
permission to perform a rollback.

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

Run read-only inspection against the exact database and namespace being
operated. For a version retirement, group all states rather than counting only
due rows; delayed retries use `pending` with a future `next_attempt_at`.

```sql
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
```

After repairing the cause and checking that the retained version is supported,
use the private worker command with the intended database configuration:

```sh
tadoku-worker replay --id 123 --actor operator-name --reason 'cause repaired'
```

The replay command requires only PostgreSQL configuration; it checks supported
versions through the same application registry and returns the new linked job
ID in its log. Verify the new record reaches its expected terminal state. An
invalid payload remains invalid on replay; repair requires its own reviewed
procedure, not repeated replay or a name change.

These inspection and replay steps do not alter the schema. Schema and SQL data
migrations follow [Database and migrations](./database.md#migrations), with a
separate PR and deployment gate before dependent code.
