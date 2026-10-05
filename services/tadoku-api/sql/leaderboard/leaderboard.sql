-- name: LeaderboardForContest :many
with leaderboard as (
  select logs.user_id, sum(coalesce(contest_logs.computed_score, contest_logs.score)) as score
  from contest_logs
  inner join logs on logs.id = contest_logs.log_id
  where contest_logs.contest_id = sqlc.arg('contest_id')
    and logs.deleted_at is null
    and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
    and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
  group by logs.user_id
), ranked_leaderboard as (
  select user_id, score, rank() over(order by score desc) as "rank"
  from leaderboard
), registrations as (
  select contest_registrations.user_id, users.display_name as user_display_name, contest_registrations.created_at
  from contest_registrations
  inner join users on users.id = contest_registrations.user_id
  where contest_id = sqlc.arg('contest_id')
    and contest_registrations.deleted_at is null
    and (sqlc.narg('language_code') = any(language_codes) or sqlc.narg('language_code') is null)
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    registrations.user_id,
    registrations.user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score,
    registrations.created_at as registered_at
  from registrations
  left join ranked_leaderboard using(user_id)
), tied_leaderboard as (
  select
    enriched_leaderboard."rank",
    enriched_leaderboard.user_id,
    enriched_leaderboard.user_display_name,
    enriched_leaderboard.score,
    enriched_leaderboard.registered_at,
    coalesce((
      "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
      or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
    ), false)::boolean as is_tie
  from enriched_leaderboard
), page as (
  select
    tied_leaderboard."rank",
    tied_leaderboard.user_id,
    tied_leaderboard.user_display_name,
    tied_leaderboard.score,
    tied_leaderboard.registered_at,
    tied_leaderboard.is_tie
  from tied_leaderboard
  order by tied_leaderboard.score desc, tied_leaderboard.registered_at asc, tied_leaderboard.user_id desc
  limit sqlc.arg('page_size')
  offset sqlc.arg('start_from')
), total as (
  select count(*) as total_size
  from tied_leaderboard
)
select
  page."rank",
  page.user_id,
  page.user_display_name,
  page.score,
  page.is_tie,
  total.total_size
from total
left join page on true
order by page.score desc, page.registered_at asc, page.user_id desc;

-- name: ContestExists :one
select exists(
  select 1
  from contests
  inner join users on users.id = contests.owner_user_id
  where contests.id = sqlc.arg('id') and contests.deleted_at is null
);

-- name: YearlyLeaderboard :many
with leaderboard as (
  select logs.user_id, sum(coalesce(computed_score, score)) as score
  from logs
  inner join users on users.id = logs.user_id and users.deleted_at is null
  where logs.year = sqlc.arg('year')
    and eligible_official_leaderboard = true
    and logs.deleted_at is null
    and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
    and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
  group by logs.user_id
), ranked_leaderboard as (
  select user_id, score, rank() over(order by score desc) as "rank" from leaderboard where score > 0
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    ranked_leaderboard.user_id,
    users.display_name::varchar as user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score
  from ranked_leaderboard
  inner join users on users.id = ranked_leaderboard.user_id
), tied_leaderboard as (
  select
    enriched_leaderboard."rank",
    enriched_leaderboard.user_id,
    enriched_leaderboard.user_display_name,
    enriched_leaderboard.score,
    coalesce((
      "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
      or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
    ), false)::boolean as is_tie
  from enriched_leaderboard
), page as (
  select
    tied_leaderboard."rank",
    tied_leaderboard.user_id,
    tied_leaderboard.user_display_name,
    tied_leaderboard.score,
    tied_leaderboard.is_tie
  from tied_leaderboard
  order by tied_leaderboard.score desc, tied_leaderboard.user_display_name asc
  limit sqlc.arg('page_size')
  offset sqlc.arg('start_from')
), total as (
  select count(*) as total_size
  from tied_leaderboard
)
select
  page."rank",
  page.user_id,
  page.user_display_name,
  page.score,
  page.is_tie,
  total.total_size
from total
left join page on true
order by page.score desc, page.user_display_name asc;

-- name: GlobalLeaderboard :many
with leaderboard as (
  select logs.user_id, sum(coalesce(computed_score, score)) as score
  from logs
  inner join users on users.id = logs.user_id and users.deleted_at is null
  where eligible_official_leaderboard = true
    and logs.deleted_at is null
    and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
    and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
  group by logs.user_id
), ranked_leaderboard as (
  select user_id, score, rank() over(order by score desc) as "rank" from leaderboard where score > 0
), enriched_leaderboard as (
  select
    rank() over(order by coalesce(ranked_leaderboard.score, 0) desc) as "rank",
    ranked_leaderboard.user_id,
    users.display_name::varchar as user_display_name,
    coalesce(ranked_leaderboard.score, 0)::real as score
  from ranked_leaderboard
  inner join users on users.id = ranked_leaderboard.user_id
), tied_leaderboard as (
  select
    enriched_leaderboard."rank",
    enriched_leaderboard.user_id,
    enriched_leaderboard.user_display_name,
    enriched_leaderboard.score,
    coalesce((
      "rank" = lag("rank", 1, -1::bigint) over (order by "rank")
      or "rank" = lead("rank", 1, -1::bigint) over (order by "rank")
    ), false)::boolean as is_tie
  from enriched_leaderboard
), page as (
  select
    tied_leaderboard."rank",
    tied_leaderboard.user_id,
    tied_leaderboard.user_display_name,
    tied_leaderboard.score,
    tied_leaderboard.is_tie
  from tied_leaderboard
  order by tied_leaderboard.score desc, tied_leaderboard.user_display_name asc
  limit sqlc.arg('page_size')
  offset sqlc.arg('start_from')
), total as (
  select count(*) as total_size
  from tied_leaderboard
)
select
  page."rank",
  page.user_id,
  page.user_display_name,
  page.score,
  page.is_tie,
  total.total_size
from total
left join page on true
order by page.score desc, page.user_display_name asc;

-- name: ContestLeaderboardAllScores :many
select cr.user_id, coalesce(scores.score, 0)::real as score
from contest_registrations cr
left join (
  select logs.user_id, sum(coalesce(contest_logs.computed_score, contest_logs.score)) as score
  from contest_logs
  inner join logs on logs.id = contest_logs.log_id
  where contest_logs.contest_id = sqlc.arg('contest_id') and logs.deleted_at is null
  group by logs.user_id
) scores on scores.user_id = cr.user_id
where cr.contest_id = sqlc.arg('contest_id') and cr.deleted_at is null;

-- name: YearlyLeaderboardAllScores :many
select logs.user_id, sum(coalesce(computed_score, score))::real as score
from logs
inner join users on users.id = logs.user_id and users.deleted_at is null
where year = sqlc.arg('year') and eligible_official_leaderboard = true and logs.deleted_at is null
group by logs.user_id
having sum(coalesce(computed_score, score)) > 0;

-- name: GlobalLeaderboardAllScores :many
select logs.user_id, sum(coalesce(computed_score, score))::real as score
from logs
inner join users on users.id = logs.user_id and users.deleted_at is null
where eligible_official_leaderboard = true and logs.deleted_at is null
group by logs.user_id
having sum(coalesce(computed_score, score)) > 0;
