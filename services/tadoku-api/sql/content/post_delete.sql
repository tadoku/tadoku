-- name: DeletePost :exec
update posts
set deleted_at = sqlc.arg(deleted_at)::timestamp
where deleted_at is null
  and namespace = sqlc.arg(namespace)
  and id = sqlc.arg(id);
