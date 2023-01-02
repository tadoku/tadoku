-- name: ListLanguages :many
select
  code,
  name
from languages
order by name asc;

-- name: ListLanguagesForContest :many
select
  code,
  name
from languages
where
  code = any((
    select language_code_allow_list
    from contests
    where id = sqlc.arg('contest_id')
  )::varchar[])
order by name asc;