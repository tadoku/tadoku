-- name: CountAnnouncements :one
select count(id)
from announcements
where deleted_at is null
  and namespace = sqlc.arg(namespace);

-- name: ListAnnouncements :many
select id, namespace, title, content, style, href,
       starts_at, ends_at, created_at, updated_at
from announcements
where deleted_at is null
  and namespace = sqlc.arg(namespace)
order by created_at desc
limit sqlc.arg(result_limit)
offset sqlc.arg(start_from);

-- name: ListActiveAnnouncements :many
select id, namespace, title, content, style, href,
       starts_at, ends_at, created_at, updated_at
from announcements
where deleted_at is null
  and namespace = sqlc.arg(namespace)
  and starts_at <= sqlc.arg(cutoff)::timestamp
  and ends_at > sqlc.arg(cutoff)::timestamp
order by starts_at desc
limit sqlc.arg(result_limit);
