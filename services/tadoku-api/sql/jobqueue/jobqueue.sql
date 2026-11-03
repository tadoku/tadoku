-- name: Enqueue :one
insert into jobs (task_type, payload, created_at, next_attempt_at)
values (sqlc.arg('task_type')::text, sqlc.arg('payload')::jsonb,
  sqlc.arg('now')::timestamptz, sqlc.arg('now')::timestamptz)
returning id;

-- name: FailExhausted :many
with exhausted as (
  select id from jobs
  where task_type = sqlc.arg('task_type')::text
    and (sqlc.narg('component')::text is null or not exists (
      select 1 from tenant_overrides as o
      where o.tenant = jobs.tenant
        and o.component = sqlc.narg('component')::text
    ))
    and attempts >= sqlc.arg('max_attempts')::integer
    and ((state = 'pending' and next_attempt_at <= sqlc.arg('now')::timestamptz)
      or (state = 'running' and lease_expires_at <= clock_timestamp()))
  order by next_attempt_at, id
  for update skip locked
  limit 100
)
update jobs as task
set state = 'failed', claim_token = null, lease_expires_at = null,
  failed_at = sqlc.arg('now')::timestamptz,
  last_error = case when task.state = 'running' then 'lease_expired' else sqlc.arg('exhausted_error')::text end
from exhausted
where task.id = exhausted.id
returning task.tenant, task.last_error;

-- name: Claim :many
with picked as (
  select id, (state = 'running') as reclaimed from jobs
  where task_type = sqlc.arg('task_type')::text
    and (sqlc.narg('component')::text is null or not exists (
      select 1 from tenant_overrides as o
      where o.tenant = jobs.tenant
        and o.component = sqlc.narg('component')::text
    ))
    and attempts < sqlc.arg('max_attempts')::integer
    and ((state = 'pending' and next_attempt_at <= sqlc.arg('now')::timestamptz)
      or (state = 'running' and lease_expires_at <= clock_timestamp()))
  order by next_attempt_at, id
  for update skip locked
  limit sqlc.arg('batch_size')::integer
), claimed as (
  update jobs as task
  set state = 'running', attempts = task.attempts + 1,
    claim_token = gen_random_uuid(),
    lease_expires_at = clock_timestamp() + sqlc.arg('lease_micros')::bigint * interval '1 microsecond'
  from picked
  where task.id = picked.id
  returning task.id, task.tenant, task.task_type, task.payload, task.attempts,
    task.claim_token, task.lease_expires_at
)
select claimed.id, claimed.tenant, claimed.task_type, claimed.payload, claimed.attempts,
  claimed.claim_token, claimed.lease_expires_at, picked.reclaimed
from claimed inner join picked using (id) order by claimed.id;

-- name: Renew :one
with locked as materialized (
  select id, lease_expires_at from jobs
  where id = sqlc.arg('id')::bigint and claim_token = sqlc.arg('claim_token')::uuid
    and state = 'running'
  for update skip locked
)
update jobs as task
set lease_expires_at = clock_timestamp() + sqlc.arg('lease_micros')::bigint * interval '1 microsecond'
from locked where task.id = locked.id and locked.lease_expires_at > clock_timestamp()
returning task.lease_expires_at;

-- name: Complete :execrows
with locked as materialized (
  select id, lease_expires_at from jobs
  where id = sqlc.arg('id')::bigint and claim_token = sqlc.arg('claim_token')::uuid
    and state = 'running'
  for update skip locked
)
update jobs as task set state = 'completed', claim_token = null,
  lease_expires_at = null, completed_at = sqlc.arg('now')::timestamptz, last_error = null
from locked where task.id = locked.id and locked.lease_expires_at > clock_timestamp();

-- name: Retry :execrows
with locked as materialized (
  select id, lease_expires_at from jobs
  where id = sqlc.arg('id')::bigint and claim_token = sqlc.arg('claim_token')::uuid
    and state = 'running'
  for update skip locked
)
update jobs as task
set state = case when attempts >= sqlc.arg('max_attempts')::integer then 'failed' else 'pending' end,
  claim_token = null, lease_expires_at = null,
  failed_at = case when attempts >= sqlc.arg('max_attempts')::integer
    then sqlc.arg('now')::timestamptz else null end,
  next_attempt_at = sqlc.arg('next_attempt_at')::timestamptz,
  last_error = sqlc.arg('error_code')::text
from locked where task.id = locked.id and locked.lease_expires_at > clock_timestamp();

-- name: Fail :execrows
with locked as materialized (
  select id, lease_expires_at from jobs
  where id = sqlc.arg('id')::bigint and claim_token = sqlc.arg('claim_token')::uuid
    and state = 'running'
  for update skip locked
)
update jobs as task set state = 'failed', claim_token = null,
  lease_expires_at = null, failed_at = sqlc.arg('now')::timestamptz,
  last_error = sqlc.arg('error_code')::text
from locked where task.id = locked.id and locked.lease_expires_at > clock_timestamp();

-- name: LockReplayLineage :one
with recursive ancestry as (
  select id, replay_of_id
  from jobs
  where id = sqlc.arg('failed_id')::bigint
    and state = 'failed'
    and task_type = any(sqlc.arg('supported_types')::text[])
  union
  select parent.id, parent.replay_of_id
  from jobs as parent
  join ancestry on parent.id = ancestry.replay_of_id
)
select root.id
from jobs as root
join ancestry on root.id = ancestry.id
where ancestry.replay_of_id is null
for update of root;

-- name: Replay :one
with recursive ancestry as (
  select id, replay_of_id
  from jobs
  where id = sqlc.arg('failed_id')::bigint
  union
  select parent.id, parent.replay_of_id
  from jobs as parent
  join ancestry on parent.id = ancestry.replay_of_id
), lineage as (
  select id, state, replay_of_id
  from jobs
  where id = (select id from ancestry where replay_of_id is null)
  union
  select child.id, child.state, child.replay_of_id
  from jobs as child
  join lineage on child.replay_of_id = lineage.id
), blocker as (
  select id, state
  from lineage
  where state in ('pending', 'running', 'completed')
    or replay_of_id = sqlc.arg('failed_id')::bigint
  order by
    case state
      when 'completed' then 0
      when 'failed' then 2
      else 1
    end,
    id
  limit 1
), inserted as (
  insert into jobs (task_type, payload, created_at, next_attempt_at,
    replay_of_id, replay_actor, replay_reason)
  select task_type, payload, sqlc.arg('now')::timestamptz, sqlc.arg('now')::timestamptz,
    id, sqlc.arg('actor')::text, sqlc.arg('reason')::text
  from jobs
  where id = sqlc.arg('failed_id')::bigint
    and state = 'failed'
    and not exists (select 1 from blocker)
  returning id
)
select coalesce((select id from inserted), 0)::bigint as replay_id,
  coalesce((select id from blocker), 0)::bigint as blocker_id,
  coalesce((select state from blocker), '')::text as blocker_state;

-- name: CleanupCompleted :execrows
with removable as (
  select parent.id from jobs as parent
  where parent.state = 'completed'
    and parent.replay_of_id is null
    and parent.completed_at < sqlc.arg('cutoff')::timestamptz
    and not exists (select 1 from jobs child where child.replay_of_id = parent.id)
  order by parent.completed_at, parent.id
  limit sqlc.arg('batch_size')::integer
)
delete from jobs as task using removable where task.id = removable.id;

-- name: Tenants :many
select key from tenants order by key;

-- name: Stats :one
select
  count(*) filter (where state = 'pending') as pending,
  count(*) filter (where state = 'failed') as failed,
  min(case when state = 'pending' then next_attempt_at
    when state = 'running' then lease_expires_at end)::timestamptz as oldest_due_at
from jobs
where task_type = sqlc.arg('task_type')::text
  and state <> 'completed'
  and (sqlc.narg('component')::text is null or not exists (
    select 1 from tenant_overrides as o
    where o.tenant = jobs.tenant
      and o.component = sqlc.narg('component')::text
  ));

-- name: UnsupportedStats :one
select
  count(*) filter (where state = 'pending') as pending,
  count(*) filter (where state = 'running') as running,
  count(*) filter (where state = 'failed') as failed,
  min(case when state = 'pending' then next_attempt_at
    when state = 'running' then lease_expires_at end)::timestamptz as oldest_due_at
from jobs
where state <> 'completed'
  and not (task_type = any(sqlc.arg('known_types')::text[]))
  and (sqlc.narg('component')::text is null or not exists (
    select 1 from tenant_overrides as o
    where o.tenant = jobs.tenant
      and o.component = sqlc.narg('component')::text
  ));
