-- name: GetPostVersion :one
select
  posts_content.id,
  posts_content.title,
  posts_content.content,
  posts_content.created_at,
  (
    select count(*)
    from posts_content earlier
    where earlier.post_id = posts_content.post_id
      and (earlier.created_at, earlier.id) <= (posts_content.created_at, posts_content.id)
  ) as position
from posts_content
inner join posts on posts.id = posts_content.post_id
where posts_content.id = sqlc.arg('content_id')
  and posts_content.post_id = sqlc.arg('post_id')
  and posts.namespace = sqlc.arg('namespace')
  and posts.deleted_at is null;
