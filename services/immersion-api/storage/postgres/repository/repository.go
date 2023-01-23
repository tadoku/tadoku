package repository

import (
	"database/sql"

	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

type Repository struct {
	psql *sql.DB
	q    *postgres.Queries
}

func NewRepository(psql *sql.DB) *Repository {
	return &Repository{
		psql: psql,
		q:    postgres.NewQueries(psql),
	}
}
