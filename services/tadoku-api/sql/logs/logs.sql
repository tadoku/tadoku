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

-- name: FetchScoresForContestProfile :many
select
  logs.language_code,
  sum(coalesce(contest_logs.computed_score, contest_logs.score))::real as score
from contest_logs
inner join logs
  on logs.id = contest_logs.log_id
where
  contest_logs.contest_id = sqlc.arg('contest_id')
  and logs.user_id = sqlc.arg('user_id')
  and logs.deleted_at is null
group by logs.language_code
order by 2 desc;

-- name: ActivityPerLanguageForContestProfile :many
with eligible_logs as (
  select
    logs.created_at::date as "date",
    logs.language_code,
    coalesce(contest_logs.computed_score, contest_logs.score) as score
  from contest_logs
  inner join logs
    on logs.id = contest_logs.log_id
  where
    contest_logs.contest_id = sqlc.arg('contest_id')
    and logs.user_id = sqlc.arg('user_id')
    and logs.deleted_at is null
)
select
  "date",
  language_code,
  sum(eligible_logs.score)::real as score
from eligible_logs
group by language_code, "date"
order by "date" asc;

-- name: ListLogsForContest :many
with eligible_logs as (
  select
    logs.id,
    logs.user_id,
    logs.language_code,
    languages.name as language_name,
    logs.log_activity_id as activity_id,
    logs.unit_id,
    coalesce(contest_logs.unit_key, '') as unit_key,
    coalesce(log_units.name, '') as unit_name,
    logs.description,
    contest_logs.amount,
    contest_logs.modifier,
    contest_logs.duration_seconds,
    coalesce(contest_logs.computed_score, contest_logs.score) as score,
    contest_logs.score_rule_set_id,
    contest_logs.score_rule_ids,
    contest_logs.score_rates,
    contest_logs.score_source,
    logs.created_at,
    logs.updated_at,
    logs.deleted_at,
    users.display_name as user_display_name,
    coalesce(
      (select array_agg(tag order by tag) from log_tags where log_id = logs.id),
      array[]::text[]
    )::text as tags
  from contest_logs
  inner join logs on (logs.id = contest_logs.log_id)
  inner join languages on (languages.code = logs.language_code)
  left join log_units on (log_units.id = logs.unit_id)
  inner join users on (users.id = logs.user_id)
  where
    (sqlc.arg('include_deleted')::boolean or logs.deleted_at is null)
    and (logs.user_id = sqlc.narg('user_id') or sqlc.narg('user_id') is null)
    and contest_logs.contest_id = sqlc.arg('contest_id')
)
select
  *,
  (select count(eligible_logs.id) from eligible_logs) as total_size
from eligible_logs
order by created_at desc
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');

-- name: ListLogsForUser :many
with eligible_logs as (
  select
    logs.id,
    logs.user_id,
    logs.language_code,
    languages.name as language_name,
    logs.log_activity_id as activity_id,
    logs.unit_id,
    coalesce(logs.unit_key, '') as unit_key,
    coalesce(log_units.name, '') as unit_name,
    logs.description,
    logs.amount,
    logs.modifier,
    logs.duration_seconds,
    coalesce(logs.computed_score, logs.score) as score,
    logs.score_rule_set_id,
    logs.score_rule_ids,
    logs.score_rates,
    logs.score_source,
    logs.created_at,
    logs.updated_at,
    logs.deleted_at,
    coalesce(
      (select array_agg(tag order by tag) from log_tags where log_id = logs.id),
      array[]::text[]
    )::text as tags
  from logs
  inner join languages on (languages.code = logs.language_code)
  left join log_units on (log_units.id = logs.unit_id)
  where
    (sqlc.arg('include_deleted')::boolean or logs.deleted_at is null)
    and logs.user_id = sqlc.arg('user_id')
)
select
  *,
  (select count(eligible_logs.id) from eligible_logs) as total_size
from eligible_logs
order by created_at desc
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');

-- name: FindLogByID :one
select
  logs.id,
  logs.user_id,
  users.display_name as user_display_name,
  logs.language_code,
  languages.name as language_name,
  logs.log_activity_id as activity_id,
  logs.unit_id,
  coalesce(logs.unit_key, '') as unit_key,
  coalesce(log_units.name, '') as unit_name,
  logs.description,
  logs.amount,
  logs.modifier,
  logs.duration_seconds,
  coalesce(logs.computed_score, logs.score) as score,
  logs.score_rule_set_id,
  logs.score_rule_ids,
  logs.score_rates,
  logs.score_source,
  logs.eligible_official_leaderboard,
  logs.created_at,
  logs.updated_at,
  logs.deleted_at,
  coalesce(
    (select array_agg(tag order by tag) from log_tags where log_id = logs.id),
    array[]::text[]
  )::text as tags
from logs
inner join languages on (languages.code = logs.language_code)
left join log_units on (log_units.id = logs.unit_id)
inner join users on (users.id = logs.user_id)
where
  (sqlc.arg('include_deleted')::boolean or logs.deleted_at is null)
  and logs.id = sqlc.arg('id');

-- name: FindAttachedContestRegistrationsForLog :many
select
  contest_logs.contest_id,
  contests.title,
  contest_registrations.id,
  contests.contest_end,
  case when owner_users.deleted_at is not null then 'Deleted organizer' else owner_users.display_name end::varchar as owner_user_display_name,
  contests.official,
  coalesce(contest_logs.computed_score, contest_logs.score) as score
from contest_logs
inner join contests on (contests.id = contest_logs.contest_id)
inner join logs on (logs.id = contest_logs.log_id)
inner join contest_registrations on (
  contest_registrations.contest_id = contest_logs.contest_id
  and contest_registrations.user_id = logs.user_id
)
inner join users as owner_users on (owner_users.id = contests.owner_user_id)
where log_id = sqlc.arg('id');
