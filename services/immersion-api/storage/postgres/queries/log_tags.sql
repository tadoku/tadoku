-- name: InsertLogTag :exec
insert into log_tags (log_id, user_id, tag)
values (sqlc.arg('log_id'), sqlc.arg('user_id'), sqlc.arg('tag'))
on conflict do nothing;

-- name: DeleteLogTagsForLog :exec
delete from log_tags where log_id = sqlc.arg('log_id');

-- name: ListTagSuggestionsForUser :many
select tag, count(*) as usage_count
from log_tags
where user_id = sqlc.arg('user_id')
  and tag ilike '%' || sqlc.arg('query') || '%'
group by tag
order by usage_count desc, tag
limit 30;

-- name: ListDefaultTagsMatching :many
select name as tag
from log_default_tags
where name ilike '%' || sqlc.arg('query') || '%'
order by name
limit 30;

