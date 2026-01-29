-- name: ListUserTags :many
select tag, count(*) as usage_count
from log_tags
where user_id = $1
  and ($2::text = '' or tag like $2 || '%')
group by tag
order by usage_count desc, tag asc
limit $3 offset $4;

-- name: ListDefaultTags :many
select name as tag
from log_default_tags
order by name asc;
