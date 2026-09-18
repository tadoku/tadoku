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

-- name: ListAnnouncements :many
with matches as materialized (
  select id, namespace, title, content, style, href,
         starts_at, ends_at, created_at, updated_at
  from announcements
  where deleted_at is null
    and namespace = sqlc.arg(namespace)
), page as (
  select *
  from matches
  order by created_at desc, id desc
  limit sqlc.arg(result_limit)
  offset sqlc.arg(start_from)::bigint
), total as (
  select count(*) as total_size
  from matches
)
select page.id, page.namespace, page.title, page.content, page.style, page.href,
       page.starts_at, page.ends_at, page.created_at, page.updated_at,
       total.total_size
from total
left join page on true
order by page.created_at desc, page.id desc;

-- name: ListActiveAnnouncements :many
select id, namespace, title, content, style, href,
       starts_at, ends_at, created_at, updated_at
from announcements
where deleted_at is null
  and namespace = sqlc.arg(namespace)
  and starts_at <= sqlc.arg(cutoff)::timestamp
  and ends_at > sqlc.arg(cutoff)::timestamp
order by starts_at desc, id desc
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
