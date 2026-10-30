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
), registrations as (
  select contest_registrations.user_id
  from contest_registrations
  inner join users on users.id = contest_registrations.user_id
  where contest_id = sqlc.arg('contest_id')
    and contest_registrations.deleted_at is null
    and (sqlc.narg('language_code') = any(language_codes) or sqlc.narg('language_code') is null)
), board as (
  select registrations.user_id, coalesce(leaderboard.score, 0)::real as score
  from registrations
  left join leaderboard using(user_id)
), page as (
  select board.user_id, board.score
  from board
  order by board.score desc, board.user_id desc
  limit sqlc.arg('page_size')
  offset sqlc.arg('start_from')
), counts as (
  select
    count(*) as total_size,
    count(*) filter (where board.score > (select max(page.score) from page)) as above_first,
    count(*) filter (where board.score >= (select min(page.score) from page)) as through_last
  from board
)
select
  page.user_id,
  page.score,
  counts.total_size,
  counts.above_first,
  counts.through_last
from counts
left join page on true
order by page.score desc, page.user_id desc;

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
), board as (
  select leaderboard.user_id, leaderboard.score::real as score
  from leaderboard
  where leaderboard.score > 0
), page as (
  select board.user_id, board.score
  from board
  order by board.score desc, board.user_id desc
  limit sqlc.arg('page_size')
  offset sqlc.arg('start_from')
), counts as (
  select
    count(*) as total_size,
    count(*) filter (where board.score > (select max(page.score) from page)) as above_first,
    count(*) filter (where board.score >= (select min(page.score) from page)) as through_last
  from board
)
select
  page.user_id,
  page.score,
  counts.total_size,
  counts.above_first,
  counts.through_last
from counts
left join page on true
order by page.score desc, page.user_id desc;

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
), board as (
  select leaderboard.user_id, leaderboard.score::real as score
  from leaderboard
  where leaderboard.score > 0
), page as (
  select board.user_id, board.score
  from board
  order by board.score desc, board.user_id desc
  limit sqlc.arg('page_size')
  offset sqlc.arg('start_from')
), counts as (
  select
    count(*) as total_size,
    count(*) filter (where board.score > (select max(page.score) from page)) as above_first,
    count(*) filter (where board.score >= (select min(page.score) from page)) as through_last
  from board
)
select
  page.user_id,
  page.score,
  counts.total_size,
  counts.above_first,
  counts.through_last
from counts
left join page on true
order by page.score desc, page.user_id desc;

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
