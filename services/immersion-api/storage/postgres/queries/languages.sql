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
left join contests
  on languages.code = any(language_code_allow_list) or language_code_allow_list is null
where
  id = sqlc.arg('contest_id')
order by name asc;

-- name: GetLanguagesByCode :many
select
  code,
  name
from languages
where
  code = any(sqlc.arg('language_codes')::varchar[])
order by name asc;