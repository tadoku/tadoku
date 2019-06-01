package infra

import (
	"github.com/jmoiron/sqlx"
	// Postgres driver that's used to connect to the db
	_ "github.com/lib/pq"

	"github.com/tadoku/api/domain"
)

// RDB is a relational database connection pool
type RDB = sqlx.DB

// NewRDB creates a new relational database connection pool
func NewRDB(URL string, maxIdleConns, maxOpenConns int) (*RDB, error) {
	db, err := sqlx.Open("postgres", URL)
	if err != nil {
		return nil, domain.WrapError(err)
	}

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)

	return db, nil
}
