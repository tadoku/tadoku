-- name: Insert :exec
insert into emails (id, template, recipient, subject, body, params, created_at)
values ($1, $2, $3, $4, $5, $6, $7);

-- name: Find :one
select id, tenant, template, recipient, subject, body, params, created_at,
  status, sent_at, provider_message_id, error
from emails
where id = $1;

-- name: List :many
select id, tenant, template, recipient, subject, body, params, created_at,
  status, sent_at, provider_message_id, error
from emails
where (sqlc.arg('recipient')::text = '' or recipient = sqlc.arg('recipient')::text)
  and (sqlc.arg('template')::text = '' or template = sqlc.arg('template')::text)
  and (sqlc.arg('status')::text = '' or status = sqlc.arg('status')::text)
  and (sqlc.narg('since')::timestamptz is null or created_at >= sqlc.narg('since')::timestamptz)
  and (sqlc.narg('before_time')::timestamptz is null
    or (created_at, id) < (sqlc.narg('before_time')::timestamptz, sqlc.arg('before_id')::uuid))
order by created_at desc, id desc
limit sqlc.arg('page_size')::integer;

-- name: Finish :exec
update emails
set status = $2, sent_at = $3, provider_message_id = $4, error = $5
where id = $1 and status = 'pending';

-- name: Retain :execrows
delete from emails
where created_at < sqlc.arg('cutoff')::timestamptz;

-- name: ReconcileFailures :exec
update emails as e
set status = case
  when sqlc.arg('hold')::boolean or e.tenant <> 'tadoku/prod' then 'held'
  else 'failed'
end,
error = case
  when sqlc.arg('hold')::boolean or e.tenant <> 'tadoku/prod' then null
  else (
  select coalesce(j.last_error, 'attempts_exhausted')
  from jobs as j
  where j.tenant = e.tenant and j.task_type = 'email.send.v1'
    and j.payload->>'email_id' = e.id::text and j.state = 'failed'
  order by j.id desc
  limit 1
  )
end
where e.status = 'pending'
  and exists (
    select 1 from jobs as j
    where j.tenant = e.tenant and j.task_type = 'email.send.v1'
      and j.payload->>'email_id' = e.id::text and j.state = 'failed'
  )
  and not exists (
    select 1 from jobs as j
    where j.tenant = e.tenant and j.task_type = 'email.send.v1'
      and j.payload->>'email_id' = e.id::text and j.state <> 'failed'
  );
