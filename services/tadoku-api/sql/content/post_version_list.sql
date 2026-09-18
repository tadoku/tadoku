-- name: ListPostVersions :many
select
  posts_content.id,
  posts_content.title,
  posts_content.created_at
from posts_content
inner join posts on posts.id = posts_content.post_id
where posts_content.post_id = sqlc.arg('post_id')
  and posts.namespace = sqlc.arg('namespace')
  and posts.deleted_at is null
order by posts_content.created_at asc, posts_content.id asc;
