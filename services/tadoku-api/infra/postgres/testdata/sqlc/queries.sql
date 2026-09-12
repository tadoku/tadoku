-- name: InsertItem :exec
insert into query_compat_items (id, label) values ($1, $2);

-- name: RenameItem :exec
update query_compat_items set label = $2 where id = $1;

-- name: GetItem :one
select id, label from query_compat_items where id = $1;

-- name: ListItems :many
select id, label from query_compat_items order by id;
