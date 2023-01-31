-- name: LeaderboardForContest :many
with leaderboard as (
  select
    user_id,
    sum(score) as score
  from logs
  inner join contest_logs
    on contest_logs.log_id = logs.id
  where
    contest_logs.contest_id = sqlc.arg('contest_id')
    and logs.deleted_at is null
    and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
    and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
  group by user_id
), ranked_leaderboard as (
  select
    user_id,
    score,
    rank() over(order by score desc) as "rank"
  from leaderboard
), registrations as (
  select
    id,
    user_id,
    user_display_name,
    created_at
  from contest_registrations
  where
    contest_id = sqlc.arg('contest_id')
    and deleted_at is null
    and (sqlc.narg('language_code') = any(language_codes) or sqlc.narg('language_code') is null)
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    registrations.user_id,
    registrations.user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score,
    (select count(registrations.user_id) from registrations) as total_size
  from registrations
  left join ranked_leaderboard using(user_id)
  order by
    score desc,
    registrations.created_at asc
)
select
  *,
  coalesce((
    "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
    or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
  ), false)::boolean as is_tie
from enriched_leaderboard
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');

-- name: YearlyLeaderboard :many
with leaderboard as (
  select
    user_id,
    sum(score) as score
  from logs
  inner join contest_logs
    on contest_logs.log_id = logs.id
  where
    logs.year = sqlc.arg('year')
    and eligible_official_leaderboard = true
    and logs.deleted_at is null
    and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
    and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
  group by user_id
), ranked_leaderboard as (
  select
    user_id,
    score,
    rank() over(order by score desc) as "rank"
  from leaderboard
), registrations as (
  select
    contest_registrations.id,
    contest_registrations.user_id,
    contest_registrations.user_display_name,
    contest_registrations.created_at
  from contest_registrations
  inner join contests
    on contests.id = contest_registrations.contest_id
  where
    extract(year from contests.contest_start) = sqlc.arg('year')::integer
    and contest_registrations.deleted_at is null
    and (sqlc.narg('language_code') = any(language_codes) or sqlc.narg('language_code') is null)
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    registrations.user_id,
    registrations.user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score,
    (select count(registrations.user_id) from registrations) as total_size
  from registrations
  left join ranked_leaderboard using(user_id)
  order by
    score desc,
    registrations.created_at asc
)
select
  *,
  coalesce((
    "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
    or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
  ), false)::boolean as is_tie
from enriched_leaderboard
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');


-- name: GlobalLeaderboard :many
with leaderboard as (
  select
    user_id,
    sum(score) as score
  from logs
  inner join contest_logs
    on contest_logs.log_id = logs.id
  where
    eligible_official_leaderboard = true
    and logs.deleted_at is null
    and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
    and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
  group by user_id
), ranked_leaderboard as (
  select
    user_id,
    score,
    rank() over(order by score desc) as "rank"
  from leaderboard
), registrations as (
  select
    contest_registrations.user_id,
    max(contest_registrations.user_display_name)::varchar as user_display_name
  from contest_registrations
  where
    contest_registrations.deleted_at is null
    and (sqlc.narg('language_code') = any(language_codes) or sqlc.narg('language_code') is null)
  group by contest_registrations.user_id
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    registrations.user_id,
    registrations.user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score,
    (select count(registrations.user_id) from registrations) as total_size
  from registrations
  left join ranked_leaderboard using(user_id)
  order by
    score desc,
    registrations.user_display_name asc
)
select
  *,
  coalesce((
    "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
    or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
  ), false)::boolean as is_tie
from enriched_leaderboard
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');