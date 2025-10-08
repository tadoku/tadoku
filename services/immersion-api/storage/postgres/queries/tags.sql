-- name: ListTags :many
select
  id,
  log_activity_id,
  name
from log_default_tags
order by log_activity_id asc;