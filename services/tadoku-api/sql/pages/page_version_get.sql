-- name: GetPageVersion :one
select
  pages_content.id,
  pages_content.title,
  pages_content.html,
  pages_content.created_at,
  (
    select count(*)
    from pages_content earlier
    where earlier.page_id = pages_content.page_id
      and (earlier.created_at, earlier.id) <= (pages_content.created_at, pages_content.id)
  ) as position
from pages_content
inner join pages on pages.id = pages_content.page_id
where pages_content.id = sqlc.arg('content_id')
  and pages_content.page_id = sqlc.arg('page_id')
  and pages.namespace = sqlc.arg('namespace')
  and pages.deleted_at is null;
