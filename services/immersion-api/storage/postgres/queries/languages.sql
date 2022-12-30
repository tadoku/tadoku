-- name: ListLanguages :many
select
  code,
  name
from languages
order by name asc;