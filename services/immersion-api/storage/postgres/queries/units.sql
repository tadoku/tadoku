-- name: ListUnits :many
select
  id,
  log_activity_id,
  name,
  modifier,
  language_code
from log_units
order by log_activity_id asc;