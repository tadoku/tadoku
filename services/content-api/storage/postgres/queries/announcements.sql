-- name: CreateAnnouncement :one
insert into announcements (
  id,
  "namespace",
  title,
  content,
  style,
  href,
  starts_at,
  ends_at
) values (
  sqlc.arg('id'),
  sqlc.arg('namespace'),
  sqlc.arg('title'),
  sqlc.arg('content'),
  sqlc.arg('style'),
  sqlc.arg('href'),
  sqlc.arg('starts_at'),
  sqlc.arg('ends_at')
) returning id;

-- name: FindAnnouncementByID :one
select
  id,
  "namespace",
  title,
  content,
  style,
  href,
  starts_at,
  ends_at,
  created_at,
  updated_at
from announcements
where
  deleted_at is null
  and id = sqlc.arg('id');

-- name: UpdateAnnouncement :one
update announcements
set
  "namespace" = sqlc.arg('namespace'),
  title = sqlc.arg('title'),
  content = sqlc.arg('content'),
  style = sqlc.arg('style'),
  href = sqlc.arg('href'),
  starts_at = sqlc.arg('starts_at'),
  ends_at = sqlc.arg('ends_at'),
  updated_at = now()
where
  id = sqlc.arg('id') and
  deleted_at is null
returning id;

-- name: DeleteAnnouncement :exec
update announcements
set deleted_at = now()
where id = sqlc.arg('id')
  and deleted_at is null;

-- name: ListAnnouncements :many
select
  id,
  "namespace",
  title,
  content,
  style,
  href,
  starts_at,
  ends_at,
  created_at,
  updated_at
from announcements
where
  deleted_at is null
  and "namespace" = sqlc.arg('namespace')
order by created_at desc
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');

-- name: AnnouncementsMetadata :one
select
  count(id) as total_size
from announcements
where
  deleted_at is null
  and "namespace" = sqlc.arg('namespace');

-- name: ListActiveAnnouncements :many
select
  id,
  "namespace",
  title,
  content,
  style,
  href,
  starts_at,
  ends_at,
  created_at,
  updated_at
from announcements
where
  deleted_at is null
  and "namespace" = sqlc.arg('namespace')
  and starts_at <= now()
  and ends_at > now()
order by starts_at desc
limit 10;
