-- name: CreatePost :exec
insert into posts (
  id, namespace, slug, current_content_id, published_at, created_at, updated_at
) values (
  sqlc.arg(id), sqlc.arg(namespace), sqlc.arg(slug), sqlc.arg(current_content_id),
  sqlc.narg(published_at), sqlc.arg(created_at), sqlc.arg(updated_at)
);

-- name: CreatePostContent :exec
insert into posts_content (id, post_id, title, content, created_at)
values (
  sqlc.arg(id), sqlc.arg(post_id), sqlc.arg(title), sqlc.arg(content), sqlc.arg(created_at)
);
