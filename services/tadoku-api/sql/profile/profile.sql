-- name: SynchronizeUser :one
insert into users (id, display_name, created_at, updated_at)
values (sqlc.arg(id), sqlc.arg(display_name), sqlc.arg(created_at), sqlc.arg(updated_at))
on conflict (id) do update set
  display_name = case
    when users.updated_at < sqlc.arg(session_created_at) then sqlc.arg(display_name)
    else users.display_name
  end,
  updated_at = case
    when users.updated_at < sqlc.arg(session_created_at) then sqlc.arg(updated_at)
    else users.updated_at
  end
where users.deletion_locked_at is null and users.deleted_at is null
returning id;

-- name: LockUser :one
select deletion_locked_at, deleted_at
from users
where id = sqlc.arg(id)
for update;
