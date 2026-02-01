-- name: FindPageBySlug :one
select
  pages.id,
  "namespace",
  slug,
  pages_content.title,
  pages_content.html,
  published_at,
  pages.created_at,
  pages.updated_at
from pages
inner join pages_content
  on pages_content.id = pages.current_content_id
where
  deleted_at is null
  and "namespace" = sqlc.arg('namespace')
  and slug = sqlc.arg('slug');

-- name: FindPageByID :one
select
  pages.id,
  "namespace",
  slug,
  pages_content.title,
  pages_content.html,
  published_at,
  pages.created_at,
  pages.updated_at
from pages
inner join pages_content
  on pages_content.id = pages.current_content_id
where
  deleted_at is null
  and pages.id = sqlc.arg('id');

-- name: ListPages :many
select
  pages.id,
  "namespace",
  slug,
  pages_content.title,
  pages_content.html,
  published_at,
  pages.created_at,
  pages.updated_at
from pages
inner join pages_content
  on pages_content.id = pages.current_content_id
where
  deleted_at is null
  and "namespace" = sqlc.arg('namespace')
  and (sqlc.arg('include_drafts')::boolean or published_at is not null)
order by pages.created_at desc
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');

-- name: CreatePage :one
insert into pages (
  id,
  "namespace",
  slug,
  current_content_id,
  published_at
) values (
  sqlc.arg('id'),
  sqlc.arg('namespace'),
  sqlc.arg('slug'),
  sqlc.arg('current_content_id'),
  sqlc.arg('published_at')
) returning id;

-- name: CreatePageContent :one
insert into pages_content (
  id,
  page_id,
  title,
  html
) values (
  sqlc.arg('id'),
  sqlc.arg('page_id'),
  sqlc.arg('title'),
  sqlc.arg('html')
) returning id;

-- name: UpdatePage :one
update pages
set
  slug = sqlc.arg('slug'),
  current_content_id = sqlc.arg('current_content_id'),
  published_at = sqlc.arg('published_at'),
  updated_at = now()
where
  id = sqlc.arg('id') and
  deleted_at is null
returning id;

-- name: PagesMetadata :one
select
  count(pages.id) as total_size,
  sqlc.arg('include_drafts')::boolean as drafts_included
from pages
where
  deleted_at is null
  and (sqlc.arg('include_drafts')::boolean or published_at is not null)
  and "namespace" = sqlc.arg('namespace');