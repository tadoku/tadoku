package content

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PostsRepository) CreatePost(ctx context.Context, item *Post) error {
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
	err = q.CreatePost(ctx, queries.CreatePostParams{
		ID:               pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace:        item.Namespace,
		Slug:             item.Slug,
		CurrentContentID: pgtype.UUID{Bytes: contentID, Valid: true},
		PublishedAt:      publishedAt,
		CreatedAt:        postgres.Timestamp(item.CreatedAt),
		UpdatedAt:        postgres.Timestamp(item.UpdatedAt),
	})
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrPostAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}

	err = q.CreatePostContent(ctx, queries.CreatePostContentParams{
		ID:        pgtype.UUID{Bytes: contentID, Valid: true},
		PostID:    pgtype.UUID{Bytes: item.ID, Valid: true},
		Title:     item.Title,
		Content:   item.Content,
		CreatedAt: postgres.Timestamp(item.CreatedAt),
	})
	if err != nil {
		return fmt.Errorf("create post content: %w", err)
	}
	return nil
}
