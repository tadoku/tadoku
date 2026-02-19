-- name: CreateLog :one
insert into logs (
  id,
  user_id,
  language_code,
  log_activity_id,
  unit_id,
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
  sqlc.arg('amount'),
  sqlc.arg('modifier'),
  sqlc.arg('eligible_official_leaderboard'),
  sqlc.arg('description')
) returning id;

-- name: CreateContestLogRelation :exec
insert into contest_logs (
  contest_id,
  log_id,
  amount,
  modifier
) values (
  (select contest_id from contest_registrations where id = sqlc.arg('registration_id')),
  sqlc.arg('log_id'),
  sqlc.arg('amount'),
  sqlc.arg('modifier')
);

-- name: ListLogsForContest :many
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
    contest_logs.amount,
    contest_logs.modifier,
    contest_logs.score,
    logs.created_at,
    logs.updated_at,
    logs.deleted_at,
    users.display_name as user_display_name,
    coalesce(
      (select array_agg(tag order by tag) from log_tags where log_id = logs.id),
      array[]::text[]
    ) as tags
  from contest_logs
  inner join logs on (logs.id = contest_logs.log_id)
  inner join languages on (languages.code = logs.language_code)
  inner join log_activities on (log_activities.id = logs.log_activity_id)
  inner join log_units on (log_units.id = logs.unit_id)
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
    log_activities.name as activity_name,
    log_units.name as unit_name,
    logs.description,
    logs.amount,
    logs.modifier,
    logs.score,
    logs.created_at,
    logs.updated_at,
    logs.deleted_at,
    coalesce(
      (select array_agg(tag order by tag) from log_tags where log_id = logs.id),
      array[]::text[]
    ) as tags
  from logs
  inner join languages on (languages.code = logs.language_code)
  inner join log_activities on (log_activities.id = logs.log_activity_id)
  inner join log_units on (log_units.id = logs.unit_id)
  where
    (sqlc.arg('include_deleted')::boolean or deleted_at is null)
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
  log_activities.name as activity_name,
  logs.unit_id,
  log_units.name as unit_name,
  logs.description,
  logs.amount,
  logs.modifier,
  logs.score,
  logs.eligible_official_leaderboard,
  logs.created_at,
  logs.updated_at,
  logs.deleted_at,
  coalesce(
    (select array_agg(tag order by tag) from log_tags where log_id = logs.id),
    array[]::text[]
  ) as tags
from logs
inner join languages on (languages.code = logs.language_code)
inner join log_activities on (log_activities.id = logs.log_activity_id)
inner join log_units on (log_units.id = logs.unit_id)
inner join users on (users.id = logs.user_id)
where
  (sqlc.arg('include_deleted')::boolean or deleted_at is null)
  and logs.id = sqlc.arg('id');

-- name: FindAttachedContestRegistrationsForLog :many
select
  contest_logs.contest_id,
  contests.title,
  contest_registrations.id,
  contests.contest_end,
  owner_users.display_name as owner_user_display_name,
  contests.official,
  contest_logs.score
from contest_logs
inner join contests on (contests.id = contest_logs.contest_id)
inner join logs on (logs.id = contest_logs.log_id)
inner join contest_registrations on (
  contest_registrations.contest_id = contest_logs.contest_id
  and contest_registrations.user_id = logs.user_id
)
inner join users as owner_users on (owner_users.id = contests.owner_user_id)
where log_id = sqlc.arg('id');

-- name: YearlyActivityForUser :many
select
  sum(score)::real as score,
  count(id) as update_count,
  created_at::date as "date"
from logs
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and deleted_at is null
group by "date"
order by date asc;

-- name: FetchScoresForProfile :many
select
  language_code,
  sum(score)::real as score,
  languages.name as language_name
from logs
inner join languages on (languages.code = logs.language_code)
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and deleted_at is null
group by language_code, languages.name
order by score desc;

-- name: YearlyActivitySplitForUser :many
select
  sum(logs.score)::real as score,
  logs.log_activity_id,
  log_activities.name as log_activity_name
from logs
inner join log_activities on (log_activities.id = logs.log_activity_id)
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and deleted_at is null
group by logs.log_activity_id, log_activities.name
order by score desc;

-- name: DeleteLog :exec
update logs
set deleted_at = now()
where
  id = sqlc.arg('log_id')
  and deleted_at is null;

-- name: CheckIfLogCanBeDeleted :one
select (not(true = any(
  select
    (contests.contest_end < sqlc.arg('now'))
  from contest_logs
  inner join contests on (contests.id = contest_logs.contest_id)
  where
    contest_logs.log_id = sqlc.arg('log_id')
)))::boolean as can_be_deleted;

-- name: FetchContestIDsForLog :many
select contest_id
from contest_logs
where log_id = sqlc.arg('log_id');

-- name: FetchLogOutboxContext :one
select user_id, year, eligible_official_leaderboard
from logs
where id = sqlc.arg('log_id');

-- name: FetchContestIDForRegistration :one
select contest_id
from contest_registrations
where id = sqlc.arg('registration_id');

-- name: DetachLogFromContest :exec
delete from contest_logs
where contest_id = sqlc.arg('contest_id')
  and log_id = sqlc.arg('log_id');

-- name: UpdateLogEligibleOfficialLeaderboard :exec
update logs
set eligible_official_leaderboard = (
  select coalesce(bool_or(contests.official), false)
  from contest_logs
  inner join contests on contests.id = contest_logs.contest_id
  where contest_logs.log_id = sqlc.arg('log_id')
),
updated_at = now()
where id = sqlc.arg('log_id');

-- name: DetachContestLogsForLanguages :exec
delete from contest_logs
where contest_id = sqlc.arg('contest_id')
  and log_id in (
    select logs.id
    from logs
    where logs.user_id = sqlc.arg('user_id')
      and logs.language_code = any(sqlc.arg('language_codes')::varchar[])
      and logs.deleted_at is null
  );