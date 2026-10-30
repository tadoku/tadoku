-- name: ContestLeaderboard :many
with scores as (
  select logs.user_id, sum(coalesce(contest_logs.computed_score, contest_logs.score)) as score
  from contest_logs
  inner join logs on logs.id = contest_logs.log_id
  where contest_logs.contest_id = sqlc.arg('contest_id')
    and logs.deleted_at is null
    and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
    and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
  group by logs.user_id
)
select contest_registrations.user_id, coalesce(scores.score, 0)::real as score
from contest_registrations
inner join users on users.id = contest_registrations.user_id
left join scores on scores.user_id = contest_registrations.user_id
where contest_registrations.contest_id = sqlc.arg('contest_id')
  and contest_registrations.deleted_at is null
  and (sqlc.narg('language_code')::varchar = any(contest_registrations.language_codes) or sqlc.narg('language_code') is null);

-- name: ContestExists :one
select exists(
  select 1
  from contests
  inner join users on users.id = contests.owner_user_id
  where contests.id = sqlc.arg('id') and contests.deleted_at is null
);

-- name: OfficialLeaderboard :many
select logs.user_id, sum(coalesce(logs.computed_score, logs.score))::real as score
from logs
inner join users on users.id = logs.user_id and users.deleted_at is null
where logs.eligible_official_leaderboard = true
  and logs.deleted_at is null
  and (logs.year = sqlc.narg('year')::smallint or sqlc.narg('year') is null)
  and (logs.language_code = sqlc.narg('language_code') or sqlc.narg('language_code') is null)
  and (logs.log_activity_id = sqlc.narg('activity_id')::integer or sqlc.narg('activity_id') is null)
group by logs.user_id
having sum(coalesce(logs.computed_score, logs.score)) > 0;
