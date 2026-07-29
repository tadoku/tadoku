-- name: ListUnits :many
select
  id,
  unit_key,
  log_activity_id,
  name,
  modifier,
  language_code
from log_units
order by log_activity_id asc;

-- name: FindUnitForTracking :one
select
  id,
  unit_key,
  log_activity_id,
  name,
  modifier,
  language_code
from log_units
where
  id = sqlc.arg('id')
  and log_activity_id = sqlc.arg('log_activity_id')
  and (language_code is null or language_code = sqlc.arg('language_code'));

-- name: FindUnitForTrackingByKey :one
select
  id,
  unit_key,
  log_activity_id,
  name,
  modifier,
  language_code
from log_units
where
  unit_key = sqlc.arg('unit_key')
  and log_activity_id = sqlc.arg('log_activity_id')
  and (language_code is null or language_code = sqlc.arg('language_code'))
order by language_code is null asc
limit 1;
