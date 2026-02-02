-- name: GetUserRoleByUserID :one
select role from user_roles where user_id = sqlc.arg('user_id');

-- name: UpsertUserRole :exec
insert into user_roles (
  user_id,
  role,
  updated_at
) values (
  sqlc.arg('user_id'),
  sqlc.arg('role'),
  now()
) on conflict (user_id) do update set
  role = sqlc.arg('role'),
  updated_at = now();

-- name: DeleteUserRole :exec
delete from user_roles where user_id = sqlc.arg('user_id');
