package pages

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PagesRepository) ListPageVersions(ctx context.Context, namespace string, id uuid.UUID) ([]PageVersion, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListPageVersions(ctx, queries.ListPageVersionsParams{
		Namespace: namespace,
		PageID:    pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list page versions: %w", err)
	}

	versions := make([]PageVersion, 0, len(rows))
	for i, row := range rows {
		versions = append(versions, PageVersion{
			ID:        uuid.UUID(row.ID.Bytes),
			Version:   i + 1,
			Title:     row.Title,
			CreatedAt: row.CreatedAt.Time,
		})
	}

	return versions, nil
}
