-- name: FindPostBySlug :one
select
  posts.id,
  "namespace",
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
  and "namespace" = sqlc.arg('namespace')
  and slug = sqlc.arg('slug');

-- name: FindPostByID :one
select
  posts.id,
  "namespace",
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
  and posts.id = sqlc.arg('id');

-- name: ListPosts :many
select
  posts.id,
  "namespace",
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
  and (sqlc.arg('include_drafts')::boolean or published_at is not null)
  and "namespace" = sqlc.arg('namespace')
order by posts.created_at desc
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');

-- name: CreatePost :one
insert into posts (
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

-- name: CreatePostContent :one
insert into posts_content (
  id,
  post_id,
  title,
  content
) values (
  sqlc.arg('id'),
  sqlc.arg('post_id'),
  sqlc.arg('title'),
  sqlc.arg('content')
) returning id;

-- name: UpdatePost :one
update posts
set
  slug = sqlc.arg('slug'),
  current_content_id = sqlc.arg('current_content_id'),
  published_at = sqlc.arg('published_at'),
  updated_at = now()
where
  id = sqlc.arg('id') and
  deleted_at is null
returning id;

-- name: PostsMetadata :one
select
  count(posts.id) as total_size,
  sqlc.arg('include_drafts')::boolean as drafts_included
from posts
where
  deleted_at is null
  and (sqlc.arg('include_drafts')::boolean or published_at is not null)
  and "namespace" = sqlc.arg('namespace');

-- name: DeletePost :exec
update posts
set deleted_at = now()
where id = sqlc.arg('id')
  and deleted_at is null;