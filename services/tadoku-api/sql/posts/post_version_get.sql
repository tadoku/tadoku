-- name: GetPostVersion :one
with versions as (
  select
    pc.id,
    row_number() over (order by pc.created_at asc, pc.id asc) as version,
    pc.title,
    pc.content,
    pc.created_at
  from posts_content pc
  join posts p on p.id = pc.post_id
  where p.id = sqlc.arg('post_id')
    and p.namespace = sqlc.arg('namespace')
    and p.deleted_at is null
)
select versions.id, versions.version, versions.title, versions.content, versions.created_at
from versions
where versions.id = sqlc.arg('content_id')::uuid;
