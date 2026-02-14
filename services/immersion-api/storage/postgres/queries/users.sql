-- name: UpsertUser :exec
insert into users (
  id,
  display_name
) values (
  sqlc.arg('id'),
  sqlc.arg('display_name')
) on conflict (id) do
update set
  display_name = sqlc.arg('display_name'),
  updated_at = now()
where
    users.updated_at < sqlc.arg('session_created_at');

-- name: FindUserDisplayNames :many
select id, display_name from users where id = any(sqlc.arg('ids')::uuid[]);