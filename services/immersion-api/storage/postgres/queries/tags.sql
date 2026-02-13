-- name: ListTags :many
select
  id,
  name
from log_default_tags
order by name asc;