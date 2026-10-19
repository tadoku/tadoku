-- name: RegisterAccount :exec
with new_user as (
  insert into users (id, display_name, created_at, updated_at)
  values (sqlc.arg(user_id), sqlc.arg(display_name), sqlc.arg(now), sqlc.arg(now))
  returning tenant, id, created_at, updated_at
), new_account as (
  insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at, updated_at)
  select tenant, id, sqlc.arg(email), sqlc.arg(email_normalized), sqlc.arg(password_hash),
    created_at at time zone 'UTC', updated_at at time zone 'UTC'
  from new_user
  returning tenant, user_id
)
insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at, created_at)
select tenant, sqlc.arg(token_hash), user_id, sqlc.arg(authenticated_at), sqlc.arg(expires_at), sqlc.arg(authenticated_at)
from new_account;

-- name: FindCredentials :one
select a.user_id, a.email, u.display_name, a.created_at, a.password_hash,
  (a.state = 'active' and u.deleted_at is null and u.deletion_locked_at is null)::boolean as active
from accounts a
join users u on u.tenant = a.tenant and u.id = a.user_id
where a.email_normalized = sqlc.arg(email_normalized);

-- name: CreateSession :execrows
insert into sessions (token_hash, user_id, authenticated_at, expires_at, created_at)
select sqlc.arg(token_hash), a.user_id, sqlc.arg(authenticated_at), sqlc.arg(expires_at), sqlc.arg(authenticated_at)
from accounts a
join users u on u.tenant = a.tenant and u.id = a.user_id
where a.user_id = sqlc.arg(user_id)
  and a.state = 'active'
  and u.deleted_at is null
  and u.deletion_locked_at is null;

-- name: FindSession :one
select a.user_id, a.email, u.display_name, a.created_at, s.authenticated_at, s.expires_at
from sessions s
join accounts a on a.tenant = s.tenant and a.user_id = s.user_id
join users u on u.tenant = a.tenant and u.id = a.user_id
where s.token_hash = sqlc.arg(token_hash)
  and s.expires_at > sqlc.arg(now)
  and a.state = 'active'
  and u.deleted_at is null
  and u.deletion_locked_at is null;

-- name: DeleteSession :exec
delete from sessions where token_hash = sqlc.arg(token_hash);
