package content

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PostsRepository) GetPostVersion(ctx context.Context, namespace string, postID, contentID uuid.UUID) (*PostVersion, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).GetPostVersion(ctx, queries.GetPostVersionParams{
		Namespace: namespace,
		PostID:    pgtype.UUID{Bytes: postID, Valid: true},
		ContentID: pgtype.UUID{Bytes: contentID, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get post version: %w", err)
	}

	return &PostVersion{
		ID:        uuid.UUID(row.ID.Bytes),
		Version:   int(row.Version),
		Title:     row.Title,
		Content:   row.Content,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}
