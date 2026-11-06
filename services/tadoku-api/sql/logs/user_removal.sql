-- name: ReadUserContestLogViews :many
select
  contest_logs.contest_id,
  logs.id,
  logs.user_id,
  users.display_name as user_display_name,
  logs.language_code,
  languages.name as language_name,
  logs.log_activity_id as activity_id,
  coalesce(logs.unit_key, '')::text as unit_key,
  logs.description,
  coalesce(contest_logs.amount, 0)::real as amount,
  coalesce(contest_logs.modifier, 0)::real as modifier,
  contest_logs.duration_seconds,
  coalesce(contest_logs.computed_score, contest_logs.score, 0)::real as score,
  logs.created_at,
  coalesce(
    (select array_agg(tag order by tag) from log_tags where log_id = logs.id),
    array[]::text[]
  )::text[] as tags
from contest_logs
inner join logs on logs.id = contest_logs.log_id
inner join users on users.id = logs.user_id
inner join languages on languages.code = logs.language_code
where logs.user_id = sqlc.arg(user_id)
  and logs.deleted_at is null
order by logs.id, contest_logs.contest_id;

-- name: DeleteUserContestLogsOutsideFinishedContests :exec
delete from contest_logs
using logs, contests
where logs.id = contest_logs.log_id
  and contests.id = contest_logs.contest_id
  and logs.user_id = sqlc.arg(user_id)
  and (logs.deleted_at is not null or contests.contest_end >= sqlc.arg(earliest_unfinished_contest_end));

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
