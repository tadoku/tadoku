-- name: LockTenant :one
select key
from tenants
where key = current_setting('tadoku.tenant')
for update;

-- name: Insert :one
insert into access_tokens (id, name, token_hash, scopes, created_at, expires_at)
select @id, @name, @token_hash, @scopes::text[], @created_at, @expires_at
where (
  select count(*)
  from access_tokens
  where tenant = current_setting('tadoku.tenant')
    and revoked_at is null
    and expires_at > @created_at
) < sqlc.arg(active_limit)::bigint
returning id;

-- name: Verify :one
update access_tokens
set last_used_at = sqlc.arg(now)
where tenant = current_setting('tadoku.tenant')
  and token_hash = @token_hash
  and revoked_at is null
  and expires_at > sqlc.arg(now)
  and scopes = sqlc.arg(accepted_scopes)::text[]
returning id, name, scopes, created_at, expires_at, last_used_at, revoked_at;

-- name: List :many
select id, name, scopes, created_at, expires_at, last_used_at, revoked_at
from access_tokens
where tenant = current_setting('tadoku.tenant')
order by created_at desc, id desc;

-- name: Revoke :one
update access_tokens
set revoked_at = coalesce(revoked_at, sqlc.arg(now))
where tenant = current_setting('tadoku.tenant') and id = @id
returning id;
