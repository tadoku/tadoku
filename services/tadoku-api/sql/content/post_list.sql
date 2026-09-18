-- name: ListPosts :many
with matches as materialized (
  select posts.id, posts.namespace, posts.slug,
         posts_content.title, posts_content.content,
         posts.published_at, posts.created_at, posts.updated_at
  from posts
  inner join posts_content on posts_content.id = posts.current_content_id
  where posts.deleted_at is null
    and posts.namespace = sqlc.arg(namespace)
    and (sqlc.arg(include_drafts)::boolean or posts.published_at <= sqlc.arg(cutoff)::timestamp)
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
select page.id, page.namespace, page.slug, page.title, page.content,
       page.published_at, page.created_at, page.updated_at,
       total.total_size
from total
left join page on true
order by page.created_at desc, page.id desc;
