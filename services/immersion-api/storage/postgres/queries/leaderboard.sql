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
    and (logs.log_activity_id = sqlc.narg('activity_id') or sqlc.narg('activity_id') is null)
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
    and (sqlc.narg('language_code') = any(language_codes) or sqlc.narg('language_code') is null)
)
select
  rank() over(order by score desc) as rank,
  registrations.user_id,
  registrations.user_display_name,
  coalesce(leaderboard.score, 0) as score
from registrations
left join leaderboard using(user_id)
order by
  score desc,
  registrations.user_id asc
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');