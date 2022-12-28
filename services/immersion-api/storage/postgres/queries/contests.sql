-- name: CreateContest :one
insert into contests (
  owner_user_id,
  owner_user_display_name,
  official,
  "private",
  contest_start,
  contest_end,
  registration_start,
  registration_end,
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
  sqlc.arg('registration_start'),
  sqlc.arg('registration_end'),
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
  registration_start = sqlc.arg('registration_start'),
  registration_end = sqlc.arg('registration_end'),
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
  registration_start,
  registration_end,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  created_at,
  updated_at,
  deleted_at
from contests
order by created_at desc;

-- name: ListPublicContests :many
select
  id,
  owner_user_id,
  owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_start,
  registration_end,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  created_at,
  updated_at,
  deleted_at
from contests
where
  "private" = false
  and deleted_at is null
order by created_at desc;

-- name: ListOfficialContests :many
select
  id,
  "private",
  contest_start,
  contest_end,
  registration_start,
  registration_end,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  created_at,
  updated_at,
  deleted_at
from contests
where
  owner_user_id is null
  and deleted_at is null
order by created_at desc;

-- name: FindContestById :one
select
  id,
  owner_user_id,
  owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_start,
  registration_end,
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
  and deleted_at is null
order by created_at desc;