-- name: ListLanguages :many
select code, name
from languages
order by name asc;

-- name: CreateLanguage :exec
insert into languages (code, name)
values (sqlc.arg(code), sqlc.arg(name));

-- name: UpdateLanguage :execrows
update languages
set name = sqlc.arg(name)
where code = sqlc.arg(code);
