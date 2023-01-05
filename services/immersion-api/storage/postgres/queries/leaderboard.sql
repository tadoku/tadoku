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
  group by user_id
), registrations as (
  select
    id,
    user_id,
    user_display_name
  from contest_registrations
  where
    contest_id = sqlc.arg('contest_id')
    and deleted_at is null
)
select
  rank() over(order by score) as rank,
  registrations.user_id,
  registrations.user_display_name,
  coalesce(leaderboard.score, 0) as score
from registrations
left join leaderboard using(user_id)
order by
  score desc,
  registrations.user_id asc;