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

-- name: YearlyActivityForUser :many
select
  sum(coalesce(computed_score, score))::real as score,
  count(id) as update_count,
  created_at::date as "date"
from logs
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and logs.deleted_at is null
group by "date"
order by date asc;

-- name: FetchScoresForProfile :many
select
  language_code,
  sum(coalesce(logs.computed_score, logs.score))::real as score,
  languages.name as language_name
from logs
inner join languages on (languages.code = logs.language_code)
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and logs.deleted_at is null
group by language_code, languages.name
order by score desc;

-- name: YearlyActivitySplitForUser :many
select
  sum(coalesce(logs.computed_score, logs.score))::real as score,
  logs.log_activity_id
from logs
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and logs.deleted_at is null
group by logs.log_activity_id
order by score desc;
