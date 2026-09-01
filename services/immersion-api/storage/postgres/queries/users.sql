-- name: UpsertUser :one
insert into users (
  id,
  display_name
) values (
  sqlc.arg('id'),
  sqlc.arg('display_name')
) on conflict (id) do
update set
  display_name = case
    when users.updated_at < sqlc.arg('session_created_at') then sqlc.arg('display_name')
    else users.display_name
  end,
  updated_at = case
    when users.updated_at < sqlc.arg('session_created_at') then now()
    else users.updated_at
  end
where
  users.deletion_locked_at is null
  and users.deleted_at is null
returning id;

-- name: LockUserForMutation :one
select *
from users
where id = sqlc.arg('id')
for update;

-- name: EnsureAccountDeletionTarget :exec
insert into users (
  id,
  display_name
) values (
  sqlc.arg('id'),
  ''
) on conflict (id) do nothing;

-- name: SetAccountDeletionLock :exec
update users
set deletion_locked_at = coalesce(deletion_locked_at, sqlc.arg('locked_at'))
where id = sqlc.arg('id');

-- name: FindUserDisplayNames :many
select id, display_name from users where id = any(sqlc.arg('ids')::uuid[]);
