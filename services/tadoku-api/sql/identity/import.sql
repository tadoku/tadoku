-- name: ImportUsers :many
select id, display_name,
  (deleted_at is not null or deletion_locked_at is not null)::boolean as deleted
from users
where tenant = current_setting('tadoku.tenant');

-- name: ImportAccounts :many
select user_id, email, email_normalized, password_hash, state, created_at
from accounts
where tenant = current_setting('tadoku.tenant');

-- name: InsertImportedUser :exec
insert into users (tenant, id, display_name, created_at, updated_at)
values (
  current_setting('tadoku.tenant'),
  sqlc.arg(id),
  sqlc.arg(display_name),
  sqlc.arg(created_at)::timestamp,
  sqlc.arg(updated_at)::timestamp
);

-- name: UpdateImportedUserName :exec
update users
set display_name = sqlc.arg(display_name)
where tenant = current_setting('tadoku.tenant')
  and id = sqlc.arg(id);

-- name: InsertImportedAccount :exec
insert into accounts (tenant, user_id, email, email_normalized, password_hash, state, created_at, updated_at)
values (
  current_setting('tadoku.tenant'),
  sqlc.arg(user_id),
  sqlc.arg(email),
  sqlc.arg(email_normalized),
  sqlc.arg(password_hash),
  sqlc.arg(state),
  sqlc.arg(created_at),
  sqlc.arg(updated_at)
);

-- name: ImportAccountCount :one
select count(*)
from accounts
where tenant = current_setting('tadoku.tenant');
