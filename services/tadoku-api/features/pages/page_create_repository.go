package pages

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PagesRepository) CreatePage(ctx context.Context, item *Page) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var publishedAt pgtype.Timestamp
	if item.PublishedAt != nil {
		publishedAt = postgres.Timestamp(*item.PublishedAt)
	}
	contentID := uuid.New()
	q := queries.New(executor)
	err = q.CreatePage(ctx, queries.CreatePageParams{
		ID:               pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace:        item.Namespace,
		Slug:             item.Slug,
		CurrentContentID: pgtype.UUID{Bytes: contentID, Valid: true},
		PublishedAt:      publishedAt,
		CreatedAt:        postgres.Timestamp(*item.CreatedAt),
		UpdatedAt:        postgres.Timestamp(*item.UpdatedAt),
	})
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrPageAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create page: %w", err)
	}

	err = q.CreatePageContent(ctx, queries.CreatePageContentParams{
		ID:        pgtype.UUID{Bytes: contentID, Valid: true},
		PageID:    pgtype.UUID{Bytes: item.ID, Valid: true},
		Title:     item.Title,
		Html:      item.HTML,
		CreatedAt: postgres.Timestamp(*item.CreatedAt),
	})
	if err != nil {
		return fmt.Errorf("create page content: %w", err)
	}
	return nil
}
