-- name: findPageBySlug :one
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