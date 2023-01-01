// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: languages.sql

package postgres

import (
	"context"
)

const listLanguages = `-- name: ListLanguages :many
select
  code,
  name
from languages
order by name asc
`

func (q *Queries) ListLanguages(ctx context.Context) ([]Language, error) {
	rows, err := q.db.QueryContext(ctx, listLanguages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Language
	for rows.Next() {
		var i Language
		if err := rows.Scan(&i.Code, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}