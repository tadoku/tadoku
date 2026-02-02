-- name: FindProfileByUserID :one
select
  id,
  user_id,
  display_name,
  bio,
  created_at,
  updated_at
from profiles
where
  deleted_at is null
  and user_id = sqlc.arg('user_id');

-- name: FindProfileByID :one
select
  id,
  user_id,
  display_name,
  bio,
  created_at,
  updated_at
from profiles
where
  deleted_at is null
  and id = sqlc.arg('id');

-- name: CreateProfile :one
insert into profiles (
  id,
  user_id,
  display_name,
  bio
) values (
  sqlc.arg('id'),
  sqlc.arg('user_id'),
  sqlc.arg('display_name'),
  sqlc.arg('bio')
) returning id;

-- name: UpdateProfile :one
update profiles
set
  display_name = sqlc.arg('display_name'),
  bio = sqlc.arg('bio'),
  updated_at = now()
where
  id = sqlc.arg('id') and
  deleted_at is null
returning id;

-- name: DeleteProfile :exec
update profiles
set
  deleted_at = now()
where
  id = sqlc.arg('id') and
  deleted_at is null;
