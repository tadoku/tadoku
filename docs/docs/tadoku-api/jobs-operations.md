---
title: Operate jobs and the worker
description: Worker health and leaderboard freshness, running jobs on a development branch, successful-job retention, expired credential removal, and inspecting and replaying retained jobs.
sidebar_position: 5.1
---

# Operate jobs and the worker

Read this when checking worker health or leaderboard freshness, running jobs
on your development branch, or inspecting, replaying or expiring stored job
rows. [Jobs and worker](./jobs.md) covers defining, publishing and handling
jobs.

## Leaderboard freshness and worker health

PostgreSQL is the source of truth, but unfiltered global, yearly and contest
leaderboards use warm Valkey cache keys derived from the request tenant, in the
form `tenant:<name>/<id>:leaderboard:…`. Private development databases sharing
`tadoku/prod` retain their route prefix as described in
[Valkey configuration](./configuration.md#valkey-and-leaderboard-caches).

PostgreSQL keeps a version per board in `leaderboard_versions`, keyed by
tenant and `global`, `yearly:<year>` or `contest:<id>`. Deferred row triggers
on `logs`, `contest_logs`, `contest_registrations`, `users` and `contests`
raise the version of every unfiltered board a write changes, whatever issued
the write. They run when the writing transaction commits, after its last
statement, and a transaction may raise a board more than once; only a change
in version is meaningful. A write that changes nothing an unfiltered board
reads, such as a description edit, freezing a log, changing registered
languages or purging a left registration, raises nothing. At commit a raise
takes a transaction-scoped advisory lock for the tenant before any version
row, so committing writers of one tenant raise in turn and cannot deadlock on
version rows or with their own row locks. A board with no row is at version 0.
A write that commits while its test tenant's deletion is in progress fails
with `lock_not_available` instead of waiting on the deletion, so its data and
its versions roll back together.
`TestLeaderboardVersionsCoverEveryLeaderboardInput` in
`services/tadoku-api/features/leaderboard/` fails when a relation in
`services/tadoku-api/sql/leaderboard/leaderboard.sql` lacks its triggers;
add the triggers with the query change.

Every unfiltered read first reads its board's version from PostgreSQL, one
primary-key lookup. A board has two Valkey keys: the sorted set and its
`:last_updated` marker, which names the version the set was built from and the
set's size. The page script serves a page only when the marker names the
version the read saw and the set's size; otherwise the read loads the board
from PostgreSQL, serves it and publishes it under that version, unless the
marker already names a later one. The version is read before the board, so a
published board is never older than the version its marker names. A served
board therefore reflects every write committed before the read began, without
the worker, and neither key expires. A lost key, a restart without persistence
or a rebuild that stops partway is only a miss. The development Valkey evicts
with `allkeys-lru` and keeps nothing across restarts. Production's Valkey is
Upstash, configured outside this repository, with eviction off.
`services/tadoku-api/spec/tla/LeaderboardCache.tla` models the read path.

Filtered reads use PostgreSQL only. Valkey errors fall back to PostgreSQL. A
PostgreSQL error on the version read fails the request: no board is served
without PostgreSQL.

Writes also publish `leaderboard.invalidate_contest.v1` and
`leaderboard.invalidate_official.v1` jobs for the previous release's reads,
which do not check versions. Their handlers delete the board's keys and
advance its `:generation`. They do not decide freshness for version-checking
reads. While pods of the previous release still serve, a pending or failed
invalidation can leave the boards those pods read stale, so replay a failed
one until they are gone.

The worker's private `/livez` answers while the process serves. `/readyz`
succeeds only when the last claim pass succeeded for every registered name and
PostgreSQL answers a ping. It does not check Valkey or mean that every job
succeeded.
Watch the [worker metrics](./worker-configuration.md#worker-metrics): pending
jobs and oldest due age show delayed work, failed jobs show terminal failures,
and unsupported-job gauges show work this worker cannot claim. The backlog
gauges refresh at startup and every fifteen seconds.

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

To replay a failed job in your tenant, run the
[replay command](#inspect-and-replay-retained-work) in your branch's worker pod
with `--tenant tadoku/<route>`; that pod already connects to the database your
branch uses. A frontend-only branch has no worker; run it in the base worker
pod instead.

`dev down` stops your worker and clears its override, so the base worker can
finish your remaining jobs before tenant teardown deletes the tenant's rows,
including its jobs.

## Successful-job retention

The worker cleans up successful history at startup and then hourly. Each run
deletes batches of 100 rows for at most five seconds; a larger backlog continues
at the next run. A base worker cleans every tenant, including tenants with an
override. A branch worker cleans only its own tenant.
`jobqueue.Service.CleanupCompleted(ctx, limit)` owns the retention rule; callers
do not supply an arbitrary cutoff. The service subtracts 90 days from the
business clock and passes that cutoff to a single repository statement.

A row is eligible only when `state = 'completed'`, it is not a replay
(`replay_of_id` is null), no row replays it, and `completed_at` is strictly
earlier than that cutoff. Exact-boundary completions remain. The cutoff does not
depend on the database session's timezone.

Completed replays stay indefinitely, like the failed jobs they replay. A
completed replay is what refuses further
[replays](#inspect-and-replay-retained-work) in its lineage, so removing it
would make the lineage replayable again. Automatic cleanup never deletes
pending, running or failed records, regardless of age.

Record important completion evidence before successful history expires. A
missing completed row is not proof that an effect never ran. Failure
inspection and version-retirement decisions still account for the retained
failed jobs.

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
tadoku-worker replay --tenant tadoku/prod --id 123 --actor operator-name --reason 'cause repaired'
```

- `--tenant` is required and names the failed job's tenant, the key you set in
  `tadoku.tenant` to inspect it. There is no default; production is
  `--tenant tadoku/prod`. A missing or malformed key exits before connecting.
  The `WORKER_POSTGRES_*` settings choose the database; `--tenant` only chooses
  the tenant within it, and the command ignores `WORKER_BRANCH`. The
  development base database also has a `tadoku/prod` tenant.
- It replays only jobs in the selected tenant, and must connect as the runtime
  role so row-level security applies. For a job in another tenant, including an
  unregistered key, a job that is not `failed`, or a name the binary does not
  register, it exits with `job is not failed or supported`, naming the selected
  tenant, and inserts nothing.
- It inserts a new `pending` job in the same tenant with the same name and
  payload, linked through `replay_of_id`, and records the trimmed actor and
  reason. It logs `job replayed` with the tenant, `source_job_id` and the new
  `job_id`. The failed original stays unchanged. A worker serving that tenant
  processes the new job.
- A job's lineage is the original job and every job replayed from it,
  directly or through other replays. A replay is refused while any job in the
  lineage is `pending`, `running` or `completed`, so replays never add a second
  unfinished job or run again after a success. A failed job can be replayed
  only if it has no replay yet; to retry
  after a replay failed, replay that failed replay. The command refuses, names
  the job that blocks the replay and inserts nothing:
  - `job lineage cannot be replayed: job <id> is completed`, or `is pending`
    or `is running`, when a job in the lineage completed or is still
    unfinished. A completed job is named first.
  - `job was already replayed as job <id>` when the job's replay failed.
    Replay that job instead.

  A replay locks its lineage's original job, so two replays in one lineage
  that run at the same time insert at most one job. The rule covers the whole
  lineage, including one where a job has several replays.
- Replaying a failed `email.send.v1` job resends its email unless the email
  is sent or held, and can deliver a message the provider already accepted.
  In a tenant that holds its mail when the replay runs, or for an email that
  retention deleted, the replay completes without sending and a failed email
  stays `failed`; that completed replay refuses further replays. Check the
  email's status as well as the job's. See
  [Retained email](./email.md#delivery-and-failures).
- Verify the new job reaches `completed`. An invalid payload remains invalid on
  replay; repair requires its own reviewed procedure, not repeated replay or a
  name change.

These inspection and replay steps do not alter the schema. Schema and SQL data
migrations follow [Database and migrations](./database.md#migrations), with a
separate PR and deployment gate before dependent code.

## Retained email maintenance

The worker performs tenant-scoped email retention at startup and daily, with
the periods in [Retained email](./email.md#retention-and-operations). A branch
worker cleans only its branch. Every fifteen seconds it also
reconciles pending email rows whose send jobs exhausted their attempts. This
is independent of the 90-day successful-job cleanup. See
[Retained email](./email.md) for inspection, production audit records and SMTP
failure semantics; never inspect production content with unaudited operator
SQL as a substitute for its read API.

## Expired credential removal

The worker removes expired native identity sessions and recovery tokens at
startup and then hourly, after successful-job cleanup. A row is eligible when
its `expires_at` is strictly earlier than the business clock, read once when a
tenant's pass starts. Exact-boundary rows remain, although they no longer
authenticate. Every unexpired session remains, and so does every unexpired
recovery token, whether it is unused, consumed or invalidated by a sibling's
consumption. Removal never deletes users, accounts, access tokens, emails or
jobs, and it does not change session or recovery-token lifetimes or revocation.

`identity.Expiry.RemoveExpired` owns the rule; callers do not supply a cutoff.
Each batch is one statement in its own implicit transaction that deletes at
most 100 sessions and 100 recovery tokens and skips rows that a concurrent
request holds locked. Batches repeat while either table fills its batch. One
pass across all tenants has a five-second budget. Worker shutdown or the end of
the budget stops it between batches, and the remaining rows wait for the next
run. A database error is logged for its tenant, the pass continues with the
next tenant, and that tenant's rows wait for the next run. Repeating a pass
removes nothing new.

A base worker enumerates every registered tenant, including tenants with an
override, and runs each tenant's batches within that tenant's ordinary
row-level security scope. A branch worker removes only its own tenant's rows.
No statement spans tenants, and removal uses the worker's ordinary database
role.

The tables have no expiry index; each batch reads the tenant's rows through the
tenant-prefixed primary key. Add an index, as a standalone migration, only when
measured query plans show the need.
