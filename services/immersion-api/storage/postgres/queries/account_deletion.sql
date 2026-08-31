-- name: LockAccountDeletionTarget :one
select id, deletion_locked_at, deleted_at
from users
where id = sqlc.arg('user_id')
for update;

-- name: HasRunningOwnedContest :one
select exists (
  select 1
  from contests
  where owner_user_id = sqlc.arg('user_id')
    and deleted_at is null
    and contest_start <= sqlc.arg('deleted_at')::date
    and (contest_end + 1)::timestamp > sqlc.arg('deleted_at')
) as has_running_contest;

-- name: ListNonHistoricalContestIDsForAccount :many
select distinct associated_contests.contest_id
from (
  select contest_registrations.contest_id
  from contest_registrations
  where contest_registrations.user_id = sqlc.arg('user_id')

  union

  select contest_logs.contest_id
  from contest_logs
  inner join logs on logs.id = contest_logs.log_id
  where logs.user_id = sqlc.arg('user_id')

  union

  select contests.id
  from contests
  where contests.owner_user_id = sqlc.arg('user_id')
) associated_contests
inner join contests on contests.id = associated_contests.contest_id
where contests.deleted_at is not null
  or (contests.contest_end + 1)::timestamp > sqlc.arg('deleted_at');

-- name: ListOfficialLeaderboardYearsForAccount :many
select distinct year
from logs
where user_id = sqlc.arg('user_id')
  and eligible_official_leaderboard = true
  and deleted_at is null
order by year;

-- name: DeleteNonHistoricalRegistrationsForAccount :exec
delete from contest_registrations
using contests
where contest_registrations.contest_id = contests.id
  and contest_registrations.user_id = sqlc.arg('user_id')
  and (
    contests.deleted_at is not null
    or (contests.contest_end + 1)::timestamp > sqlc.arg('deleted_at')
  );

-- name: DetachNonHistoricalLogsForAccount :exec
delete from contest_logs
using logs, contests
where contest_logs.log_id = logs.id
  and contest_logs.contest_id = contests.id
  and logs.user_id = sqlc.arg('user_id')
  and (
    contests.deleted_at is not null
    or (contests.contest_end + 1)::timestamp > sqlc.arg('deleted_at')
  );

-- name: CancelFutureOwnedContestsForAccount :exec
update contests
set
  deleted_at = coalesce(deleted_at, sqlc.arg('canceled_at')),
  owner_user_display_name = 'Deleted organizer',
  updated_at = sqlc.arg('canceled_at')
where owner_user_id = sqlc.arg('user_id')
  and contest_start > sqlc.arg('today')::date;

-- name: AnonymizeOwnedContestsForAccount :exec
update contests
set owner_user_display_name = 'Deleted organizer'
where owner_user_id = sqlc.arg('user_id');

-- name: FreezeHistoricalLogsForAccount :exec
update logs
set
  frozen_at = coalesce(frozen_at, sqlc.arg('deleted_at')),
  description = null
where user_id = sqlc.arg('user_id')
  and deleted_at is null
  and exists (
    select 1
    from contest_logs
    inner join contests on contests.id = contest_logs.contest_id
    where contest_logs.log_id = logs.id
      and contests.deleted_at is null
      and (contests.contest_end + 1)::timestamp <= sqlc.arg('deleted_at')
  );

-- name: DeleteTagsForAccount :exec
delete from log_tags
where user_id = sqlc.arg('user_id');

-- name: DeleteNonHistoricalLogsForAccount :exec
delete from logs
where user_id = sqlc.arg('user_id')
  and not exists (
    select 1
    from contest_logs
    inner join contests on contests.id = contest_logs.contest_id
    where contest_logs.log_id = logs.id
      and contests.deleted_at is null
      and (contests.contest_end + 1)::timestamp <= sqlc.arg('deleted_at')
  );

-- name: MarkAccountDeleted :exec
update users
set
  display_name = 'Deleted participant',
  deleted_at = coalesce(deleted_at, sqlc.arg('deleted_at')),
  updated_at = sqlc.arg('deleted_at')
where id = sqlc.arg('user_id');
