-- name: SynchronizeUser :one
insert into users (id, display_name, created_at, updated_at)
values (sqlc.arg(id), sqlc.arg(display_name), sqlc.arg(created_at), sqlc.arg(updated_at))
on conflict (id) do update set
  display_name = case
    when users.updated_at < sqlc.arg(session_created_at) then sqlc.arg(display_name)
    else users.display_name
  end,
  updated_at = case
    when users.updated_at < sqlc.arg(session_created_at) then sqlc.arg(updated_at)
    else users.updated_at
  end
where users.deletion_locked_at is null and users.deleted_at is null
returning id;

-- name: LockUser :one
select deletion_locked_at, deleted_at
from users
where id = sqlc.arg(id)
for update;

-- name: YearlyActivityForUser :many
select
  sum(coalesce(computed_score, score))::real as score,
  count(id) as update_count,
  created_at::date as "date"
from logs
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and logs.deleted_at is null
group by "date"
order by date asc;

-- name: FetchScoresForProfile :many
select
  language_code,
  sum(coalesce(logs.computed_score, logs.score))::real as score,
  languages.name as language_name
from logs
inner join languages on (languages.code = logs.language_code)
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and logs.deleted_at is null
group by language_code, languages.name
order by score desc;

-- name: YearlyActivitySplitForUser :many
select
  sum(coalesce(logs.computed_score, logs.score))::real as score,
  logs.log_activity_id
from logs
where
  user_id = sqlc.arg('user_id')
  and year = sqlc.arg('year')
  and logs.deleted_at is null
group by logs.log_activity_id
order by score desc;
