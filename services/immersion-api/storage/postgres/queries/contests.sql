-- name: CreateContest :one
insert into contests (
  owner_user_id,
  owner_user_display_name,
  official,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list
) values (
  sqlc.arg('owner_user_id'),
  sqlc.arg('owner_user_display_name'),
  sqlc.arg('official'),
  sqlc.arg('private'),
  sqlc.arg('contest_start'),
  sqlc.arg('contest_end'),
  sqlc.arg('registration_end'),
  sqlc.arg('title'),
  sqlc.arg('description'),
  sqlc.arg('language_code_allow_list'),
  sqlc.arg('activity_type_id_allow_list')
) returning id;

-- name: UpdateContest :one
update contests
set
  "private" = sqlc.arg('private'),
  contest_start = sqlc.arg('contest_start'),
  contest_end = sqlc.arg('contest_end'),
  registration_end = sqlc.arg('registration_end'),
  title = sqlc.arg('title'),
  "description" = sqlc.arg('description'),
  updated_at = now()
where
  id = sqlc.arg('id')
  and deleted_at is null
returning id;

-- name: CancelContest :one
update contests
set deleted_at = now()
where
  id = sqlc.arg('id')
  and deleted_at is null
returning id;

-- name: ListContests :many
select
  id,
  owner_user_id,
  owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  created_at,
  updated_at,
  deleted_at
from contests
where
  (sqlc.arg('include_deleted')::boolean or deleted_at is null)
  and (owner_user_id = sqlc.narg('user_id') or sqlc.narg('user_id') is null)
  and official = sqlc.arg('official')
  and ("private" = false or (sqlc.arg('include_private')::boolean or owner_user_id = sqlc.narg('user_id')))
order by created_at desc
limit sqlc.arg('page_size')
offset sqlc.arg('start_from');

-- name: FindContestById :one
select
  id,
  owner_user_id,
  owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  created_at,
  updated_at,
  deleted_at
from contests
where
  id = sqlc.arg('id')
  and (sqlc.arg('include_deleted')::boolean or deleted_at is null)
order by created_at desc;

-- name: ContestsMetadata :one
select
  count(contests.id) as total_size,
  sqlc.arg('include_deleted')::boolean as include_deleted
from contests
where
  (sqlc.arg('include_deleted')::boolean or deleted_at is null)
  and (owner_user_id = sqlc.narg('user_id') or sqlc.narg('user_id') is null)
  and (official = sqlc.arg('official'));

-- name: FindLatestOfficialContest :one
select
  id,
  owner_user_id,
  owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  created_at,
  updated_at,
  deleted_at
from contests
where
  official = true
order by contest_start desc
limit 1;

-- name: ContestSummary :one
select
  sum(logs.score) as total_score,
  count(distinct logs.user_id) as participant_count,
  count(distinct logs.language_code) as language_count
from logs
inner join contest_logs
  on contest_logs.log_id = logs.id
where
  contest_logs.contest_id = sqlc.arg('contest_id')
  and logs.deleted_at is null;