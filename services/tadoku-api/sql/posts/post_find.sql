-- name: FindPostBySlug :one
select
  posts.id,
  namespace,
  slug,
  posts_content.title,
  posts_content.content,
  published_at,
  posts.created_at,
  posts.updated_at
from posts
inner join posts_content
  on posts_content.id = posts.current_content_id
where
  deleted_at is null
  and namespace = sqlc.arg('namespace')
  and slug = sqlc.arg('slug');

-- name: FindPostByID :one
select
  posts.id,
  namespace,
  slug,
  posts_content.title,
  posts_content.content,
  published_at,
  posts.created_at,
  posts.updated_at
from posts
inner join posts_content
  on posts_content.id = posts.current_content_id
where
  deleted_at is null
  and posts.id = sqlc.arg('id')
  and namespace = sqlc.arg('namespace');
