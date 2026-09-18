package content

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PostsRepository) ListPostVersions(ctx context.Context, namespace string, id uuid.UUID) ([]PostVersion, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	rows, err := queries.New(executor).ListPostVersions(ctx, queries.ListPostVersionsParams{
		Namespace: namespace,
		PostID:    pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("list post versions: %w", err)
	}

	versions := make([]PostVersion, 0, len(rows))
	for i, row := range rows {
		versions = append(versions, PostVersion{
			ID:        uuid.UUID(row.ID.Bytes),
			Version:   i + 1,
			Title:     row.Title,
			CreatedAt: row.CreatedAt.Time,
		})
	}

	return versions, nil
}
