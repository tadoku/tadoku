package postgres

import (
	"context"
	"database/sql"

	"github.com/tadoku/tadoku/services/blog-api/domain/pagecreate"
)

type PageRepository struct {
	psql *sql.DB
	q    *Queries
}

func NewPageRepository(psql *sql.DB) *PageRepository {
	return &PageRepository{
		psql: psql,
		q:    &Queries{psql},
	}
}

func (r *PageRepository) CreatePage(ctx context.Context, req *pagecreate.PageCreateRequest) (*pagecreate.PageCreateResponse, error) {
	return nil, nil
}
