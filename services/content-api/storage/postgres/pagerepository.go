package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/tadoku/tadoku/services/content-api/domain/pagecreate"
	"github.com/tadoku/tadoku/services/content-api/domain/pagefind"
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
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	pageContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	_, err = qtx.CreatePage(ctx, CreatePageParams{
		req.ID,
		req.Slug,
		pageContentID,
		NewNullTime(req.PublishedAt),
	})
	if err != nil {
		_ = tx.Rollback()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, pagecreate.ErrPageAlreadyExists
		}

		return nil, fmt.Errorf("could not create page: %w", err)
	}

	_, err = qtx.CreatePageContent(ctx, CreatePageContentParams{
		ID:     pageContentID,
		PageID: req.ID,
		Title:  req.Title,
		Html:   req.Html,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	page, err := qtx.FindPageBySlug(ctx, req.Slug)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create page: %w", err)
	}

	return &pagecreate.PageCreateResponse{
		ID:    page.ID,
		Slug:  page.Slug,
		Title: page.Title,
		Html:  page.Html,
	}, nil
}

func (r *PageRepository) FindBySlug(ctx context.Context, slug string) (*pagefind.PageFindResponse, error) {
	page, err := r.q.FindPageBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pagefind.ErrPageNotFound
		}

		return nil, fmt.Errorf("could not find page: %w", err)
	}

	return &pagefind.PageFindResponse{
		ID:          page.ID,
		Slug:        page.Slug,
		Title:       page.Title,
		Html:        page.Html,
		PublishedAt: NewTimeFromNullTime(page.PublishedAt),
	}, nil
}