-- name: ListPages :many
with matches as materialized (
  select pages.id, pages.namespace, pages.slug,
         pages_content.title, pages_content.html,
         pages.published_at, pages.created_at, pages.updated_at
  from pages
  inner join pages_content on pages_content.id = pages.current_content_id
  where pages.deleted_at is null
    and pages.namespace = sqlc.arg(namespace)
    and (sqlc.arg(include_drafts)::boolean or pages.published_at is not null)
), page as (
  select *
  from matches
  order by created_at desc, id desc
  limit sqlc.arg(result_limit)
  offset sqlc.arg(start_from)::bigint
), total as (
  select count(*) as total_size
  from matches
)
select page.id, page.namespace, page.slug, page.title, page.html,
       page.published_at, page.created_at, page.updated_at,
       total.total_size
from total
left join page on true
order by page.created_at desc, page.id desc;
