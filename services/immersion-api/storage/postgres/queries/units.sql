-- name: ListUnits :many
select
  id,
  log_activity_id,
  name,
  modifier,
  language_code
from log_units
order by log_activity_id asc;

-- name: FindUnitForTracking :one
select
  id,
  log_activity_id,
  name,
  modifier,
  language_code
from log_units
where
  id = sqlc.arg('id')
  and log_activity_id = sqlc.arg('log_activity_id')
  and (language_code is null or language_code = sqlc.arg('language_code'));