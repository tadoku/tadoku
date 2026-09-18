-- name: CreatePage :exec
insert into pages (
  id, namespace, slug, current_content_id, published_at, created_at, updated_at
) values (
  sqlc.arg(id), sqlc.arg(namespace), sqlc.arg(slug), sqlc.arg(current_content_id),
  sqlc.narg(published_at), sqlc.arg(created_at), sqlc.arg(updated_at)
);

-- name: CreatePageContent :exec
insert into pages_content (id, page_id, title, html, created_at)
values (
  sqlc.arg(id), sqlc.arg(page_id), sqlc.arg(title), sqlc.arg(html), sqlc.arg(created_at)
);
