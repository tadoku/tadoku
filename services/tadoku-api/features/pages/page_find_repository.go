package pages

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PagesRepository) FindPageBySlug(ctx context.Context, namespace, slug string) (*Page, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindPageBySlug(ctx, queries.FindPageBySlugParams{
		Namespace: namespace,
		Slug:      slug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find page by slug: %w", err)
	}
	return pageFromFindRow(row), nil
}

func (r *PagesRepository) FindPageByID(ctx context.Context, namespace string, id uuid.UUID) (*Page, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindPageByID(ctx, queries.FindPageByIDParams{
		Namespace: namespace,
		ID:        pgtype.UUID{Bytes: id, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find page by ID: %w", err)
	}
	return pageFromFindRow(queries.FindPageBySlugRow(row)), nil
}

func pageFromFindRow(row queries.FindPageBySlugRow) *Page {
	var publishedAt *time.Time
	if row.PublishedAt.Valid {
		publishedAt = &row.PublishedAt.Time
	}
	return &Page{
		ID:          uuid.UUID(row.ID.Bytes),
		Namespace:   row.Namespace,
		Slug:        row.Slug,
		Title:       row.Title,
		HTML:        row.Html,
		PublishedAt: publishedAt,
		CreatedAt:   &row.CreatedAt.Time,
		UpdatedAt:   &row.UpdatedAt.Time,
	}
}
