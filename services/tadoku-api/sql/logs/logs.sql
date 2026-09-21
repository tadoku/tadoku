-- name: ListUnits :many
select
  id,
  unit_key,
  log_activity_id,
  name,
  modifier,
  language_code
from log_units
order by log_activity_id asc;

-- name: ListDistinctLanguageCodesForUser :many
select distinct language_code
from logs
where logs.user_id = sqlc.arg('user_id') and logs.deleted_at is null;

-- name: ListTagSuggestionsForUser :many
select tag, count(*) as usage_count
from log_tags
where user_id = sqlc.arg('user_id')
  and tag ilike '%' || sqlc.arg('query') || '%'
group by tag
order by usage_count desc, tag
limit 30;
