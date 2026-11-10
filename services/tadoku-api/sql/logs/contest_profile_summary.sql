-- name: ContestProfileSummaryDays :many
with filters as not materialized (
  select
    sqlc.narg('language_codes')::text[] as language_codes,
    sqlc.narg('activity_ids')::smallint[] as activity_ids,
    sqlc.narg('tags')::text[] as tag_filter
)
select
  logs.created_at::date as "date",
  logs.log_activity_id as activity_id,
  sum(coalesce(contest_logs.computed_score, contest_logs.score))::real as score,
  count(*) as log_count
from contest_logs
inner join logs
  on logs.id = contest_logs.log_id
cross join filters
where
  contest_logs.contest_id = sqlc.arg('contest_id')
  and logs.user_id = sqlc.arg('user_id')
  and logs.deleted_at is null
  and (filters.language_codes is null or logs.language_code = any(filters.language_codes))
  and (filters.activity_ids is null or logs.log_activity_id = any(filters.activity_ids))
  and (
    filters.tag_filter is null
    or exists (
      select 1
      from log_tags
      where
        log_tags.log_id = logs.id
        and log_tags.tag = any(filters.tag_filter)
    )
  )
group by logs.created_at::date, logs.log_activity_id
order by "date" asc, activity_id asc;

-- name: ContestProfileSummaryLanguages :many
with filters as not materialized (
  select
    sqlc.narg('activity_ids')::smallint[] as activity_ids,
    sqlc.narg('tags')::text[] as tag_filter
), scores as (
  select
    logs.language_code,
    sum(coalesce(contest_logs.computed_score, contest_logs.score))::real as total_score
  from contest_logs
  inner join logs
    on logs.id = contest_logs.log_id
  cross join filters
  where
    contest_logs.contest_id = sqlc.arg('contest_id')
    and logs.user_id = sqlc.arg('user_id')
    and logs.deleted_at is null
    and (filters.activity_ids is null or logs.log_activity_id = any(filters.activity_ids))
    and (
      filters.tag_filter is null
      or exists (
        select 1
        from log_tags
        where
          log_tags.log_id = logs.id
          and log_tags.tag = any(filters.tag_filter)
      )
    )
  group by logs.language_code
  order by total_score desc, logs.language_code asc
  limit 100
)
select
  scores.language_code,
  languages.name as language_name,
  scores.total_score as score
from scores
inner join lateral (
  select name
  from languages
  where languages.code = scores.language_code
  limit 1
) languages on true
order by scores.total_score desc, scores.language_code asc;

-- name: ContestProfileSummaryActivities :many
with filters as not materialized (
  select
    sqlc.narg('language_codes')::text[] as language_codes,
    sqlc.narg('tags')::text[] as tag_filter
)
select
  logs.log_activity_id as activity_id,
  sum(coalesce(contest_logs.computed_score, contest_logs.score))::real as total_score
from contest_logs
inner join logs
  on logs.id = contest_logs.log_id
cross join filters
where
  contest_logs.contest_id = sqlc.arg('contest_id')
  and logs.user_id = sqlc.arg('user_id')
  and logs.deleted_at is null
  and (filters.language_codes is null or logs.language_code = any(filters.language_codes))
  and (
    filters.tag_filter is null
    or exists (
      select 1
      from log_tags
      where
        log_tags.log_id = logs.id
        and log_tags.tag = any(filters.tag_filter)
    )
  )
group by logs.log_activity_id
order by total_score desc, activity_id asc
limit 100;

-- name: ContestProfileSummaryTags :many
with filters as not materialized (
  select
    sqlc.narg('language_codes')::text[] as language_codes,
    sqlc.narg('activity_ids')::smallint[] as activity_ids
)
select
  log_tags.tag,
  sum(coalesce(contest_logs.computed_score, contest_logs.score))::real as total_score,
  count(*) as log_count
from contest_logs
inner join logs
  on logs.id = contest_logs.log_id
inner join log_tags
  on log_tags.log_id = logs.id
cross join filters
where
  contest_logs.contest_id = sqlc.arg('contest_id')
  and logs.user_id = sqlc.arg('user_id')
  and logs.deleted_at is null
  and (filters.language_codes is null or logs.language_code = any(filters.language_codes))
  and (filters.activity_ids is null or logs.log_activity_id = any(filters.activity_ids))
group by log_tags.tag
order by total_score desc, log_tags.tag asc
limit 200;
