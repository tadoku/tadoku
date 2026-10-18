---
title: Operate jobs and the worker
description: Worker health and leaderboard freshness, running jobs on a development branch, successful-job retention, and inspecting and replaying retained jobs.
sidebar_position: 5.1
---

# Operate jobs and the worker

Read this when checking worker health or leaderboard freshness, running jobs
on your development branch, or inspecting, replaying or expiring stored job
rows. [Jobs and worker](./jobs.md) covers defining, publishing and handling
jobs.

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
Watch the [worker metrics](./worker-configuration.md#worker-metrics): pending jobs and
oldest due age show delayed invalidations, failed jobs show terminal failures,
and unsupported-job gauges show work this worker cannot claim. The backlog
gauges refresh at startup and every fifteen seconds. Investigate any failed
leaderboard invalidation in `jobs`. Inspect its `last_error` and repair the
cause before using the
[replay command](#inspect-and-replay-retained-work). Verify the replayed job
completes and a subsequent leaderboard read reflects PostgreSQL. A later write
can also publish another invalidation, but do not depend on users making one to
recover a stale board.

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
