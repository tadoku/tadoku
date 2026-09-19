-- name: ListContests :many
select
  contests.id,
  owner_user_id,
  case when users.deleted_at is not null then 'Deleted organizer' else users.display_name end::varchar as owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  contests.created_at,
  contests.updated_at,
  contests.deleted_at
from contests
inner join users on users.id = contests.owner_user_id
where
  (sqlc.arg(include_deleted)::boolean or contests.deleted_at is null)
  and (owner_user_id = sqlc.narg(user_id)::uuid or sqlc.narg(user_id)::uuid is null)
  and official = sqlc.arg(official)
  and ("private" = false or (sqlc.arg(include_private)::boolean or owner_user_id = sqlc.narg(user_id)::uuid))
order by contests.created_at desc
limit sqlc.arg(page_size)
offset sqlc.arg(start_from);

-- name: ContestsMetadata :one
select count(contests.id) as total_size
from contests
where
  (sqlc.arg(include_deleted)::boolean or contests.deleted_at is null)
  and (owner_user_id = sqlc.narg(user_id)::uuid or sqlc.narg(user_id)::uuid is null)
  and official = sqlc.arg(official);

-- name: FindContestByID :one
select
  contests.id,
  owner_user_id,
  case when users.deleted_at is not null then 'Deleted organizer' else users.display_name end::varchar as owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  contests.created_at,
  contests.updated_at,
  contests.deleted_at
from contests
inner join users on users.id = contests.owner_user_id
where
  contests.id = sqlc.arg(id)
  and (sqlc.arg(include_deleted)::boolean or contests.deleted_at is null)
order by contests.created_at desc;

-- name: FindLatestOfficialContest :one
select
  contests.id,
  owner_user_id,
  case when users.deleted_at is not null then 'Deleted organizer' else users.display_name end::varchar as owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  contests.created_at,
  contests.updated_at,
  contests.deleted_at
from contests
inner join users on users.id = contests.owner_user_id
where official = true
order by contest_start desc
limit 1;

-- name: ListLanguagesForContest :many
select code, name
from languages
left join contests
  on languages.code = any(language_code_allow_list) or language_code_allow_list is null
where id = sqlc.arg(contest_id)
order by name asc;

-- name: ListLanguages :many
select code, name
from languages
order by name asc;
