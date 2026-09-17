-- name: DeleteAnnouncement :exec
update announcements
set deleted_at = sqlc.arg(deleted_at)::timestamp
where deleted_at is null
  and namespace = sqlc.arg(namespace)
  and id = sqlc.arg(id);

-- name: CreateAnnouncement :exec
insert into announcements (
  id, namespace, title, content, style, href,
  starts_at, ends_at, created_at, updated_at
) values (
  sqlc.arg(id), sqlc.arg(namespace), sqlc.arg(title), sqlc.arg(content),
  sqlc.arg(style), sqlc.arg(href), sqlc.arg(starts_at), sqlc.arg(ends_at),
  sqlc.arg(created_at), sqlc.arg(updated_at)
);

-- name: FindAnnouncementByID :one
select id, namespace, title, content, style, href,
       starts_at, ends_at, created_at, updated_at
from announcements
where deleted_at is null
  and namespace = sqlc.arg(namespace)
  and id = sqlc.arg(id);

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
offset sqlc.arg(start_from)::bigint;

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

-- name: UpdateAnnouncement :one
update announcements
set title = sqlc.arg(title),
    content = sqlc.arg(content),
    style = sqlc.arg(style),
    href = sqlc.narg(href),
    starts_at = sqlc.arg(starts_at),
    ends_at = sqlc.arg(ends_at),
    updated_at = sqlc.arg(updated_at)
where id = sqlc.arg(id)
  and namespace = sqlc.arg(namespace)
  and deleted_at is null
returning id;
