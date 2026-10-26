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
  and u.deletion_locked_at is null
  and a.password_hash = sqlc.arg(password_hash)
for update of a;

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

-- name: LockRecoveryAccount :one
select a.user_id, a.email, u.display_name, a.created_at, a.password_hash
from accounts a
join users u on u.tenant = a.tenant and u.id = a.user_id
where a.email_normalized = sqlc.arg(email_normalized)
  and a.state = 'active'
  and u.deleted_at is null
  and u.deletion_locked_at is null
for update of a;

-- name: LockAccount :one
select a.user_id, a.email, u.display_name, a.created_at, a.password_hash
from accounts a
join users u on u.tenant = a.tenant and u.id = a.user_id
where a.user_id = sqlc.arg(user_id)
  and a.state = 'active'
  and u.deleted_at is null
  and u.deletion_locked_at is null
for update of a;

-- name: CreateRecoveryToken :exec
insert into recovery_tokens (token_hash, user_id, expires_at)
values (sqlc.arg(token_hash), sqlc.arg(user_id), sqlc.arg(expires_at));

-- name: RecoveryAccount :one
select user_id from recovery_tokens where token_hash = sqlc.arg(token_hash);

-- name: UseRecoveryToken :execrows
update recovery_tokens
set used_at = sqlc.arg(now)
where token_hash = sqlc.arg(token_hash)
  and used_at is null
  and expires_at > sqlc.arg(now);

-- name: InvalidateRecoveryTokens :exec
update recovery_tokens
set used_at = sqlc.arg(now)
where user_id = sqlc.arg(user_id)
  and used_at is null;

-- name: RevokeOtherSessions :exec
delete from sessions
where user_id = sqlc.arg(user_id)
  and token_hash <> sqlc.arg(keep_token_hash);

-- name: LockAccountForRequest :exec
select a.user_id
from accounts a
join users u on u.tenant = a.tenant and u.id = a.user_id
where a.user_id = sqlc.arg(user_id)
  and u.deleted_at is null
  and u.deletion_locked_at is null
for update of a;

-- name: LockAccountForDeletion :exec
select user_id from accounts where user_id = sqlc.arg(user_id) for update;

-- name: AccountHasEmail :one
select exists (
  select 1
  from accounts
  where user_id = sqlc.arg(user_id)
    and email_normalized = sqlc.arg(email_normalized)
)::boolean;

-- name: DeleteAccount :many
with deleted_recovery_tokens as (
  delete from recovery_tokens where recovery_tokens.user_id = sqlc.arg(user_id)::uuid
), deleted_sessions as (
  delete from sessions where sessions.user_id = sqlc.arg(user_id)::uuid
)
delete from accounts
where accounts.user_id = sqlc.arg(user_id)::uuid
returning accounts.email;

-- name: DeleteExpiredCredentials :one
with expired_sessions as (
  delete from sessions
  where token_hash in (
    select expired.token_hash
    from sessions as expired
    where expired.expires_at < sqlc.arg(now)
    limit sqlc.arg(batch_size)::integer
    for update skip locked
  )
  returning 1
), expired_recovery_tokens as (
  delete from recovery_tokens
  where token_hash in (
    select expired.token_hash
    from recovery_tokens as expired
    where expired.expires_at < sqlc.arg(now)
    limit sqlc.arg(batch_size)::integer
    for update skip locked
  )
  returning 1
)
select
  (select count(*) from expired_sessions)::integer as sessions,
  (select count(*) from expired_recovery_tokens)::integer as recovery_tokens;

-- name: UpdateAccountEmail :exec
update accounts
set email = sqlc.arg(email), email_normalized = sqlc.arg(email), updated_at = sqlc.arg(now)
where user_id = sqlc.arg(user_id);

-- name: UpdateDisplayName :exec
update users
set display_name = sqlc.arg(display_name), updated_at = sqlc.arg(now)
where id = sqlc.arg(user_id);

-- name: UpdateAccountPassword :exec
update accounts
set password_hash = sqlc.arg(password_hash), updated_at = sqlc.arg(now)
where user_id = sqlc.arg(user_id);

-- name: DevelopmentAccountExists :one
select exists (
  select 1
  from accounts
  where tenant = current_setting('tadoku.tenant')
    and user_id = sqlc.arg(user_id)
);

-- name: InsertDevelopmentUser :exec
insert into users (tenant, id, display_name, created_at, updated_at)
values (
  current_setting('tadoku.tenant'),
  sqlc.arg(id),
  sqlc.arg(display_name),
  coalesce((
    select created_at
    from users
    where tenant = 'tadoku/prod'
      and id = sqlc.arg(id)
  ), now() at time zone 'UTC'),
  now() at time zone 'UTC'
)
on conflict (tenant, id) do nothing;

-- name: InsertDevelopmentAccount :exec
insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at)
select tenant, id, sqlc.arg(email), sqlc.arg(email), sqlc.arg(password_hash), created_at at time zone 'UTC'
from users
where tenant = current_setting('tadoku.tenant')
  and id = sqlc.arg(user_id)
on conflict (tenant, user_id) do nothing;
