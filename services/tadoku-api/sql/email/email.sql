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
where created_at < sqlc.arg('now')::timestamptz - case
  when tenant = 'tadoku/prod' then interval '3 days'
  else interval '14 days'
end;
