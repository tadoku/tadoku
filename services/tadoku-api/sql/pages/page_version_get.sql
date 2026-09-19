-- name: GetPageVersion :one
with versions as (
  select
    pc.id,
    row_number() over (order by pc.created_at asc, pc.id asc) as version,
    pc.title,
    pc.html,
    pc.created_at
  from pages_content pc
  join pages p on p.id = pc.page_id
  where p.id = sqlc.arg('page_id')
    and p.namespace = sqlc.arg('namespace')
    and p.deleted_at is null
)
select versions.id, versions.version, versions.title, versions.html, versions.created_at
from versions
where versions.id = sqlc.arg('content_id')::uuid;
