-- name: findPageBySlug :one
select
  id,
  slug,
  title,
  html,
  published_at
from pages
where
  deleted_at is null
  and slug = sqlc.arg('slug');
