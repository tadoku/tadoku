-- name: ListUserContestLogRemovals :many
select
  contest_logs.contest_id,
  contest_logs.log_id,
  (contests.contest_end >= sqlc.arg(now))::boolean as withdrawn
from contest_logs
inner join logs on logs.id = contest_logs.log_id
inner join contests on contests.id = contest_logs.contest_id
where logs.user_id = sqlc.arg(user_id)
  and logs.deleted_at is null
  and (
    contests.contest_end >= sqlc.arg(now)
    or logs.description is not null
    or exists (select 1 from log_tags where log_tags.log_id = logs.id)
  )
order by contest_logs.log_id, contest_logs.contest_id;

-- name: ListOfficialYearsForUser :many
select distinct year
from logs
where user_id = sqlc.arg(user_id)
  and deleted_at is null
  and eligible_official_leaderboard
order by year;

-- name: DeleteUserContestLogsOutsideFinishedContests :exec
delete from contest_logs
using logs, contests
where logs.id = contest_logs.log_id
  and contests.id = contest_logs.contest_id
  and logs.user_id = sqlc.arg(user_id)
  and (logs.deleted_at is not null or contests.contest_end >= sqlc.arg(now));

-- name: FreezeUserContestHistory :execrows
update logs
set description = null, frozen_at = coalesce(frozen_at, sqlc.arg(frozen_at)::timestamp)
where user_id = sqlc.arg(user_id)
  and exists (select 1 from contest_logs where contest_logs.log_id = logs.id);

-- name: DeleteUserLogTags :exec
delete from log_tags where user_id = sqlc.arg(user_id);

-- name: DeleteUserLogsOutsideContests :execrows
delete from logs
where user_id = sqlc.arg(user_id)
  and not exists (select 1 from contest_logs where contest_logs.log_id = logs.id);
