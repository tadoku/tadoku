-- name: ListActivities :many
select
  id,
  name,
  "default",
  time_modifier,
  input_type
from log_activities
order by id asc;

-- name: ListActivitiesForContest :many
select
  id,
  name,
  time_modifier,
  input_type
from log_activities
where
  id = any((
    select activity_type_id_allow_list
    from contests
    where contests.id = sqlc.arg('contest_id')
  )::integer[])
order by name asc;

-- name: FindActivityByID :one
select
  id,
  name,
  "default",
  time_modifier,
  input_type
from log_activities
where id = sqlc.arg('id');