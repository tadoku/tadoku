-- name: ListActivities :many
select
  id,
  name,
  "default"
from log_activities
order by name asc;