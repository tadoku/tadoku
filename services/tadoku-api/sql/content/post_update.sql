-- name: UpdatePost :one
update posts
set
  slug = sqlc.arg('slug'),
  current_content_id = coalesce(sqlc.narg('current_content_id')::uuid, current_content_id),
  published_at = sqlc.narg('published_at'),
  updated_at = sqlc.arg('updated_at')
where id = sqlc.arg('id')
  and namespace = sqlc.arg('namespace')
  and deleted_at is null
returning id;
