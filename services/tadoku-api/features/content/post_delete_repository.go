package content

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PostsRepository) DeletePost(ctx context.Context, namespace string, id uuid.UUID, deletedAt time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).DeletePost(ctx, queries.DeletePostParams{
		Namespace: namespace,
		ID:        pgtype.UUID{Bytes: id, Valid: true},
		DeletedAt: postgres.Timestamp(deletedAt),
	})
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	return nil
}
