-- name: FindAccount :one
select a.user_id, a.email, u.display_name, a.created_at
from accounts a
join users u on u.tenant = a.tenant and u.id = a.user_id
where a.user_id = sqlc.arg(user_id)
  and a.state = 'active'
  and u.deleted_at is null;

-- name: ListAccounts :many
select a.user_id, a.email, u.display_name, a.created_at
from accounts a
join users u on u.tenant = a.tenant and u.id = a.user_id
where a.state = 'active'
  and u.deleted_at is null
  and a.user_id > sqlc.arg(after_id)::uuid
  and (
    strpos(lower(a.email), lower(sqlc.arg(query)::text)) > 0
    or strpos(lower(u.display_name), lower(sqlc.arg(query)::text)) > 0
  )
order by a.user_id
limit sqlc.arg(page_limit);
