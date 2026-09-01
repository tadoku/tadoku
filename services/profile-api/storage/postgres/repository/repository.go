package repository

import (
	"database/sql"

	"github.com/tadoku/tadoku/services/profile-api/storage/postgres"
)

type Repository struct {
	q *postgres.Queries
}

func NewRepository(psql *sql.DB) *Repository {
	return &Repository{q: postgres.New(psql)}
}
