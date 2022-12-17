-- name: FindPageBySlug :one
select
  pages.id,
  slug,
  pages_content.title,
  pages_content.html,
  published_at
from pages
inner join pages_content
  on pages_content.id = pages.current_content_id
where
  deleted_at is null
  and slug = sqlc.arg('slug');

-- name: CreatePage :one
insert into pages (
  id,
  slug,
  current_content_id,
  published_at
) values (
  sqlc.arg('id'),
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