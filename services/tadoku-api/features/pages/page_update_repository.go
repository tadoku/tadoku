package pages

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PagesRepository) UpdatePage(ctx context.Context, item *Page, contentChanged bool) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var publishedAt pgtype.Timestamp
	if item.PublishedAt != nil {
		publishedAt = postgres.Timestamp(*item.PublishedAt)
	}
	var contentID pgtype.UUID
	if contentChanged {
		contentID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	}

	query := queries.New(executor)
	_, err = query.UpdatePage(ctx, queries.UpdatePageParams{
		ID:               pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace:        item.Namespace,
		Slug:             item.Slug,
		CurrentContentID: contentID,
		PublishedAt:      publishedAt,
		UpdatedAt:        postgres.Timestamp(*item.UpdatedAt),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPageNotFound
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrPageAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("update page: %w", err)
	}

	if contentChanged {
		err = query.CreatePageContent(ctx, queries.CreatePageContentParams{
			ID:        contentID,
			PageID:    pgtype.UUID{Bytes: item.ID, Valid: true},
			Title:     item.Title,
			Html:      item.HTML,
			CreatedAt: postgres.Timestamp(*item.UpdatedAt),
		})
		if err != nil {
			return fmt.Errorf("create page revision: %w", err)
		}
	}
	return nil
}
