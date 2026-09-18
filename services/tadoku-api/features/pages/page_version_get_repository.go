package pages

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PagesRepository) GetPageVersion(ctx context.Context, namespace string, pageID, contentID uuid.UUID) (*PageVersion, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).GetPageVersion(ctx, queries.GetPageVersionParams{
		Namespace: namespace,
		PageID:    pgtype.UUID{Bytes: pageID, Valid: true},
		ContentID: pgtype.UUID{Bytes: contentID, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPageNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get page version: %w", err)
	}

	return &PageVersion{
		ID:        uuid.UUID(row.ID.Bytes),
		Version:   int(row.Version),
		Title:     row.Title,
		HTML:      row.Html,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}
