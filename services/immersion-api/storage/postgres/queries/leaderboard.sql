-- name: LeaderboardForContest :many
with leaderboard as (
  select
    logs.user_id,
    sum(coalesce(contest_logs.computed_score, contest_logs.score)) as score
  from contest_logs
  inner join logs
    on logs.id = contest_logs.log_id
  where
    contest_logs.contest_id = sqlc.arg('contest_id')
    and logs.deleted_at is null
    and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
    and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
  group by logs.user_id
), ranked_leaderboard as (
  select
    user_id,
    score,
    rank() over(order by score desc) as "rank"
  from leaderboard
), registrations as (
  select
    contest_registrations.id,
    user_id,
    users.display_name as user_display_name,
    contest_registrations.created_at
  from contest_registrations
  inner join users on users.id = contest_registrations.user_id
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
    sum(coalesce(computed_score, score)) as score
  from logs
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
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    ranked_leaderboard.user_id,
    users.display_name::varchar as user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score
  from ranked_leaderboard
  inner join users on users.id = ranked_leaderboard.user_id
  order by
    score desc,
    user_display_name asc
)
select
  *,
  coalesce((
    "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
    or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
  ), false)::boolean as is_tie,
  (select count(user_id) from enriched_leaderboard) as total_size
from enriched_leaderboard
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');

-- name: ContestLeaderboardAllScores :many
-- Returns all user scores for a contest without pagination/ranking.
-- Used for rebuilding the Redis leaderboard sorted set.
select
  cr.user_id,
  coalesce(scores.score, 0)::real as score
from contest_registrations cr
left join (
  select
    logs.user_id,
    sum(coalesce(contest_logs.computed_score, contest_logs.score)) as score
  from contest_logs
  inner join logs on logs.id = contest_logs.log_id
  where
    contest_logs.contest_id = sqlc.arg('contest_id')
    and logs.deleted_at is null
  group by logs.user_id
) scores on scores.user_id = cr.user_id
where
  cr.contest_id = sqlc.arg('contest_id')
  and cr.deleted_at is null;

-- name: YearlyLeaderboardAllScores :many
-- Returns all user scores for a year without pagination/ranking.
-- Used for rebuilding the Redis leaderboard sorted set.
select
  user_id,
  sum(coalesce(computed_score, score))::real as score
from logs
where
  year = sqlc.arg('year')
  and eligible_official_leaderboard = true
  and deleted_at is null
group by user_id
having sum(coalesce(computed_score, score)) > 0;

-- name: GlobalLeaderboardAllScores :many
-- Returns all user scores globally without pagination/ranking.
-- Used for rebuilding the Redis leaderboard sorted set.
select
  user_id,
  sum(coalesce(computed_score, score))::real as score
from logs
where
  eligible_official_leaderboard = true
  and deleted_at is null
group by user_id
having sum(coalesce(computed_score, score)) > 0;

-- name: UserContestScore :one
-- Returns a single user's total score for a contest.
-- Used for updating a user's score in the Redis leaderboard.
select
  coalesce(sum(coalesce(contest_logs.computed_score, contest_logs.score)), 0)::real as score
from contest_logs
inner join logs on logs.id = contest_logs.log_id
where
  contest_logs.contest_id = sqlc.arg('contest_id')
  and logs.user_id = sqlc.arg('user_id')
  and logs.deleted_at is null;

-- name: UserYearlyScore :one
-- Returns a single user's total score for a year (official logs only).
-- Used for updating a user's score in the Redis leaderboard.
select
  coalesce(sum(coalesce(computed_score, score)), 0)::real as score
from logs
where
  year = sqlc.arg('year')
  and user_id = sqlc.arg('user_id')
  and eligible_official_leaderboard = true
  and deleted_at is null;

-- name: UserGlobalScore :one
-- Returns a single user's total global score (official logs only).
-- Used for updating a user's score in the Redis leaderboard.
select
  coalesce(sum(coalesce(computed_score, score)), 0)::real as score
from logs
where
  user_id = sqlc.arg('user_id')
  and eligible_official_leaderboard = true
  and deleted_at is null;

-- name: GlobalLeaderboard :many
with leaderboard as (
  select
    user_id,
    sum(coalesce(computed_score, score)) as score
  from logs
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
  where score > 0
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    ranked_leaderboard.user_id,
    users.display_name::varchar as user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score
  from ranked_leaderboard
  inner join users on users.id = ranked_leaderboard.user_id
  order by
    score desc,
    user_display_name asc
)
select
  *,
  coalesce((
    "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
    or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
  ), false)::boolean as is_tie,
  (select count(user_id) from enriched_leaderboard) as total_size
from enriched_leaderboard
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');