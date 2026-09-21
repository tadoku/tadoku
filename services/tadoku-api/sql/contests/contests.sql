-- name: ListContests :many
with matches as materialized (
  select
    contests.id,
    owner_user_id,
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
  where
    (sqlc.arg(include_deleted)::boolean or contests.deleted_at is null)
    and (owner_user_id = sqlc.narg(user_id)::uuid or sqlc.narg(user_id)::uuid is null)
    and official = sqlc.arg(official)
), page as (
  select
    matches.id,
    matches.owner_user_id,
    case when users.deleted_at is not null then 'Deleted organizer' else users.display_name end::varchar as owner_user_display_name,
    matches."private",
    matches.contest_start,
    matches.contest_end,
    matches.registration_end,
    matches.title,
    matches."description",
    matches.language_code_allow_list,
    matches.activity_type_id_allow_list,
    matches.official,
    matches.created_at,
    matches.updated_at,
    matches.deleted_at
  from matches
  inner join users on users.id = matches.owner_user_id
  where matches."private" = false
    or sqlc.arg(include_private)::boolean
    or matches.owner_user_id = sqlc.narg(user_id)::uuid
  order by matches.created_at desc
  limit sqlc.arg(page_size)
  offset sqlc.arg(start_from)
), total as (
  select count(*) as total_size
  from matches
)
select
  page.id,
  page.owner_user_id,
  page.owner_user_display_name,
  page."private",
  page.contest_start,
  page.contest_end,
  page.registration_end,
  page.title,
  page."description",
  page.language_code_allow_list,
  page.activity_type_id_allow_list,
  page.official,
  page.created_at,
  page.updated_at,
  page.deleted_at,
  total.total_size
from total
left join page on true
order by page.created_at desc;

-- name: CountContestsCreatedByUserForYear :one
select count(id)
from contests
where owner_user_id = sqlc.arg(owner_user_id)
  and extract(year from created_at) = sqlc.arg(year)::integer;

-- name: LanguagesExist :one
select count(distinct languages.code) = count(distinct requested.code)
from unnest(sqlc.arg(codes)::varchar[]) as requested(code)
left join languages using (code);

-- name: CreateContest :exec
insert into contests (
  id,
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
  activity_type_id_allow_list,
  created_at,
  updated_at
) values (
  sqlc.arg(id),
  sqlc.arg(owner_user_id),
  sqlc.arg(owner_user_display_name),
  sqlc.arg(official),
  sqlc.arg(private),
  sqlc.arg(contest_start),
  sqlc.arg(contest_end),
  sqlc.arg(registration_end),
  sqlc.arg(title),
  sqlc.narg(description),
  sqlc.arg(language_code_allow_list),
  sqlc.arg(activity_type_id_allow_list),
  sqlc.arg(created_at),
  sqlc.arg(updated_at)
);

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

-- name: FindContestRegistrationForUser :one
select
  contest_registrations.id,
  contest_registrations.contest_id,
  contest_registrations.user_id,
  contest_registrations.language_codes,
  contest_registrations.created_at,
  contest_registrations.updated_at,
  users.display_name as user_display_name
from contest_registrations
inner join contests on contests.id = contest_registrations.contest_id
inner join users on users.id = contest_registrations.user_id
where contest_registrations.user_id = sqlc.arg(user_id)
  and contest_registrations.contest_id = sqlc.arg(contest_id)
  and contest_registrations.deleted_at is null;

-- name: ListRegistrationLanguages :many
select code, name
from languages
where code = any(sqlc.arg(codes)::varchar[])
order by name asc;

-- name: ListOngoingContestRegistrations :many
select
  contest_registrations.id,
  contest_registrations.contest_id,
  contest_registrations.user_id,
  contest_registrations.language_codes,
  users.display_name as user_display_name,
  contests.activity_type_id_allow_list,
  contests.registration_end,
  contests.contest_start,
  contests.contest_end,
  contests.private,
  contests.official,
  contests.title,
  contests.description,
  contests.owner_user_id,
  owner_users.display_name as owner_user_display_name
from contest_registrations
inner join contests on contests.id = contest_registrations.contest_id
inner join users on users.id = contest_registrations.user_id
inner join users as owner_users on owner_users.id = contests.owner_user_id
where contest_registrations.user_id = sqlc.arg(user_id)
  and contests.contest_start <= sqlc.arg(now)::timestamp
  and (contests.contest_end + '1 day'::interval) > sqlc.arg(now)::timestamp
  and contest_registrations.deleted_at is null;

-- name: UpsertContestRegistration :exec
insert into contest_registrations (
  id,
  contest_id,
  user_id,
  language_codes,
  created_at,
  updated_at
) values (
  sqlc.arg(id),
  sqlc.arg(contest_id),
  sqlc.arg(user_id),
  sqlc.arg(language_codes),
  sqlc.arg(created_at),
  sqlc.arg(updated_at)
) on conflict (id) do update set
  language_codes = sqlc.arg(language_codes),
  updated_at = sqlc.arg(updated_at);

-- name: DetachContestLogsForLanguages :exec
delete from contest_logs
where contest_id = sqlc.arg(contest_id)
  and log_id in (
    select logs.id
    from logs
    where logs.user_id = sqlc.arg(user_id)
      and logs.language_code = any(sqlc.arg(language_codes)::varchar[])
      and logs.deleted_at is null
  );

-- name: InsertContestScoreRefresh :exec
insert into leaderboard_outbox (event_type, user_id, contest_id)
values ('refresh_contest_score', sqlc.arg(user_id), sqlc.arg(contest_id));

-- name: InsertOfficialScoresRefresh :exec
insert into leaderboard_outbox (event_type, user_id, year)
values ('refresh_official_scores', sqlc.arg(user_id), sqlc.arg(year));
