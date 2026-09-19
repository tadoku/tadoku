package pages

import "github.com/jackc/pgx/v5/pgxpool"

type PagesRepository struct {
	db *pgxpool.Pool
}

func NewPagesRepository(db *pgxpool.Pool) *PagesRepository {
	return &PagesRepository{db: db}
}
