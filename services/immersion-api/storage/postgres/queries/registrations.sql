
-- name: FindContestRegistrationForUser :one
select
  contest_registrations.id,
  contest_registrations.contest_id,
  contest_registrations.user_id,
  contest_registrations.user_display_name,
  contest_registrations.language_codes,
  contests.activity_type_id_allow_list,
  contests.registration_end,
  contests.contest_start,
  contests.contest_end,
  contests.private,
  contests.official,
  contests.title,
  contests.description
from contest_registrations
inner join contests
  on contests.id = contest_registrations.contest_id
where
  user_id = sqlc.arg('user_id')
  and contest_id = sqlc.arg('contest_id')
  and contest_registrations.deleted_at is null;

-- name: UpsertContestRegistration :one
insert into contest_registrations (
  id,
  contest_id,
  user_id,
  user_display_name,
  language_codes
) values (
  sqlc.arg('id'),
  sqlc.arg('contest_id'),
  sqlc.arg('user_id'),
  sqlc.arg('user_display_name'),
  sqlc.arg('language_codes')
) on conflict (id) do
update set
  language_codes = sqlc.arg('language_codes'),
  updated_at = now()
returning id;

-- name: FindOngoingContestRegistrationForUser :many
select
  contest_registrations.id,
  contest_registrations.contest_id,
  contest_registrations.user_id,
  contest_registrations.user_display_name,
  contest_registrations.language_codes,
  contests.activity_type_id_allow_list,
  contests.registration_end,
  contests.contest_start,
  contests.contest_end,
  contests.private,
  contests.official,
  contests.title,
  contests.description
from contest_registrations
inner join contests
  on contests.id = contest_registrations.contest_id
where
  user_id = sqlc.arg('user_id')
  and contests.contest_start <= sqlc.arg('now')::timestamp
  and (contests.contest_end + '1 day'::interval) > sqlc.arg('now')::timestamp
  and contest_registrations.deleted_at is null;


-- name: FindYearlyContestRegistrationForUser :many
select
  contest_registrations.id,
  contest_registrations.contest_id,
  contest_registrations.user_id,
  contest_registrations.user_display_name,
  contest_registrations.language_codes,
  contests.activity_type_id_allow_list,
  contests.registration_end,
  contests.contest_start,
  contests.contest_end,
  contests.private,
  contests.official,
  contests.title,
  contests.description
from contest_registrations
inner join contests
  on contests.id = contest_registrations.contest_id
where
  user_id = sqlc.arg('user_id')
  and (contests.private != true or sqlc.arg('include_private')::boolean)
  and extract(year from contests.contest_start) = sqlc.arg('year')::integer
  and contest_registrations.deleted_at is null;