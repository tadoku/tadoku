-- name: ListPageVersions :many
select
  pages_content.id,
  pages_content.title,
  pages_content.created_at
from pages_content
inner join pages on pages.id = pages_content.page_id
where pages_content.page_id = sqlc.arg('page_id')
  and pages.namespace = sqlc.arg('namespace')
  and pages.deleted_at is null
order by pages_content.created_at asc, pages_content.id asc;
