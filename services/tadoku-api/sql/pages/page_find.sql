-- name: FindPageBySlug :one
select
  pages.id,
  namespace,
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
  and namespace = sqlc.arg('namespace')
  and slug = sqlc.arg('slug');

-- name: FindPageByID :one
select
  pages.id,
  namespace,
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
  and pages.id = sqlc.arg('id')
  and namespace = sqlc.arg('namespace');
