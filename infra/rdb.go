package infra

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type RDB = sqlx.DB

func NewRDB(URL string, maxIdleConns, maxOpenConns int) (*RDB, error) {
	db, err := sqlx.Open("postgres", URL)
	if err != nil {
		// @TODO: we should wrap errors
		return nil, err
	}

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)

	return db, nil
}
