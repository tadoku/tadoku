-- name: CreateLog :one
insert into logs (
  id,
  user_id,
  language_code,
  log_activity_id,
  unit_id,
  tags,
  amount,
  modifier,
  eligible_official_leaderboard,
  "description"
) values (
  sqlc.arg('id'),
  sqlc.arg('user_id'),
  sqlc.arg('language_code'),
  sqlc.arg('log_activity_id'),
  sqlc.arg('unit_id'),
  sqlc.arg('tags'),
  sqlc.arg('amount'),
  sqlc.arg('modifier'),
  sqlc.arg('eligible_official_leaderboard'),
  sqlc.arg('description')
) returning id;

-- name: CreateContestLogRelation :exec
insert into contest_logs (
  contest_id,
  log_id
) values (
  (select contest_id from contest_registrations where id = sqlc.arg('registration_id')),
  sqlc.arg('log_id')
);

-- name: ListLogsForContestUser :many
with eligible_logs as (
  select
    logs.id,
    logs.user_id,
    logs.language_code,
    languages.name as language_name,
    logs.log_activity_id as activity_id,
    log_activities.name as activity_name,
    log_units.name as unit_name,
    logs.description,
    logs.tags,
    logs.amount,
    logs.modifier,
    logs.score,
    logs.created_at,
    logs.updated_at,
    logs.deleted_at
  from contest_logs
  inner join logs on (logs.id = contest_logs.log_id)
  inner join languages on (languages.code = logs.language_code)
  inner join log_activities on (log_activities.id = logs.log_activity_id)
  inner join log_units on (log_units.id = logs.unit_id)
  where
    (sqlc.arg('include_deleted')::boolean or deleted_at is null)
    and logs.user_id = sqlc.arg('user_id')
    and contest_logs.contest_id = sqlc.arg('contest_id')
)
select
  *,
  (select count(eligible_logs.id) from eligible_logs) as total_size
from eligible_logs
order by created_at desc
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');