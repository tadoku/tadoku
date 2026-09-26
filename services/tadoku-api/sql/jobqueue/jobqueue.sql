-- name: Enqueue :one
insert into async_outbox (task_type, payload, created_at, next_attempt_at)
values (sqlc.arg('task_type')::text, sqlc.arg('payload')::jsonb,
  sqlc.arg('now')::timestamptz, sqlc.arg('now')::timestamptz)
returning id;

-- name: Claim :many
with exhausted as (
  select id from async_outbox
  where task_type = sqlc.arg('task_type')::text
    and attempts >= sqlc.arg('max_attempts')::integer
    and ((state = 'pending' and next_attempt_at <= sqlc.arg('now')::timestamptz)
      or (state = 'running' and lease_expires_at <= clock_timestamp()))
  order by next_attempt_at, id
  for update skip locked
  limit 100
), failed as (
  update async_outbox as task
  set state = 'failed', claim_token = null, lease_expires_at = null,
    failed_at = sqlc.arg('now')::timestamptz,
    last_error = case when task.state = 'running' then 'lease_expired' else 'attempts_exhausted' end
  from exhausted
  where task.id = exhausted.id
  returning task.id
), picked as (
  select id, (state = 'running') as reclaimed from async_outbox
  where task_type = sqlc.arg('task_type')::text
    and attempts < sqlc.arg('max_attempts')::integer
    and ((state = 'pending' and next_attempt_at <= sqlc.arg('now')::timestamptz)
      or (state = 'running' and lease_expires_at <= clock_timestamp()))
  order by next_attempt_at, id
  for update skip locked
  limit sqlc.arg('batch_size')::integer
), claimed as (
  update async_outbox as task
  set state = 'running', attempts = task.attempts + 1,
    claim_token = gen_random_uuid(),
    lease_expires_at = clock_timestamp() + sqlc.arg('lease_micros')::bigint * interval '1 microsecond'
  from picked
  where task.id = picked.id
  returning task.id, task.task_type, task.payload, task.attempts,
    task.claim_token, task.lease_expires_at
)
select claimed.id, claimed.task_type, claimed.payload, claimed.attempts,
  claimed.claim_token, claimed.lease_expires_at, picked.reclaimed
from claimed inner join picked using (id) order by claimed.id;

-- name: Renew :one
with locked as materialized (
  select id, lease_expires_at from async_outbox
  where id = sqlc.arg('id')::bigint and claim_token = sqlc.arg('claim_token')::uuid
    and state = 'running'
  for update skip locked
)
update async_outbox as task
set lease_expires_at = clock_timestamp() + sqlc.arg('lease_micros')::bigint * interval '1 microsecond'
from locked where task.id = locked.id and locked.lease_expires_at > clock_timestamp()
returning task.lease_expires_at;

-- name: Complete :execrows
with locked as materialized (
  select id, lease_expires_at from async_outbox
  where id = sqlc.arg('id')::bigint and claim_token = sqlc.arg('claim_token')::uuid
    and state = 'running'
  for update skip locked
)
update async_outbox as task set state = 'completed', claim_token = null,
  lease_expires_at = null, completed_at = sqlc.arg('now')::timestamptz, last_error = null
from locked where task.id = locked.id and locked.lease_expires_at > clock_timestamp();

-- name: Retry :execrows
with locked as materialized (
  select id, lease_expires_at from async_outbox
  where id = sqlc.arg('id')::bigint and claim_token = sqlc.arg('claim_token')::uuid
    and state = 'running'
  for update skip locked
)
update async_outbox as task
set state = case when attempts >= sqlc.arg('max_attempts')::integer then 'failed' else 'pending' end,
  claim_token = null, lease_expires_at = null,
  failed_at = case when attempts >= sqlc.arg('max_attempts')::integer
    then sqlc.arg('now')::timestamptz else null end,
  next_attempt_at = sqlc.arg('next_attempt_at')::timestamptz,
  last_error = sqlc.arg('error_code')::text
from locked where task.id = locked.id and locked.lease_expires_at > clock_timestamp();

-- name: Fail :execrows
with locked as materialized (
  select id, lease_expires_at from async_outbox
  where id = sqlc.arg('id')::bigint and claim_token = sqlc.arg('claim_token')::uuid
    and state = 'running'
  for update skip locked
)
update async_outbox as task set state = 'failed', claim_token = null,
  lease_expires_at = null, failed_at = sqlc.arg('now')::timestamptz,
  last_error = sqlc.arg('error_code')::text
from locked where task.id = locked.id and locked.lease_expires_at > clock_timestamp();

-- name: Replay :one
insert into async_outbox (task_type, payload, created_at, next_attempt_at,
  replay_of_id, replay_actor, replay_reason)
select task_type, payload, sqlc.arg('now')::timestamptz, sqlc.arg('now')::timestamptz,
  id, sqlc.arg('actor')::text, sqlc.arg('reason')::text
from async_outbox where id = sqlc.arg('failed_id')::bigint and state = 'failed'
  and task_type = any(sqlc.arg('supported_types')::text[])
returning id;

-- name: CleanupCompleted :execrows
with removable as (
  select parent.id from async_outbox as parent
  where parent.state = 'completed'
    and parent.completed_at < ((sqlc.arg('now')::timestamptz at time zone 'UTC') - interval '3 months') at time zone 'UTC'
    and not exists (select 1 from async_outbox child where child.replay_of_id = parent.id)
  order by parent.completed_at, parent.id
  limit sqlc.arg('batch_size')::integer
)
delete from async_outbox as task using removable where task.id = removable.id;

-- name: Outstanding :one
select count(*) from async_outbox
where task_type = sqlc.arg('task_type')::text and state in ('pending', 'running');

-- name: Stats :one
select
  count(*) filter (where state = 'pending') as pending,
  count(*) filter (where state = 'failed') as failed,
  min(case when state = 'pending' then next_attempt_at
    when state = 'running' then lease_expires_at end)::timestamptz as oldest_due_at
from async_outbox where task_type = sqlc.arg('task_type')::text;

-- name: UnsupportedStats :one
select
  count(*) filter (where state = 'pending') as pending,
  count(*) filter (where state = 'running') as running,
  count(*) filter (where state = 'failed') as failed,
  min(case when state = 'pending' then next_attempt_at
    when state = 'running' then lease_expires_at end)::timestamptz as oldest_due_at
from async_outbox
where state <> 'completed' and not (task_type = any(sqlc.arg('known_types')::text[]));
