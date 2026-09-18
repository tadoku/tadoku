package content

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PostsRepository) UpdatePost(ctx context.Context, post *Post, contentChanged bool) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var publishedAt pgtype.Timestamp
	if post.PublishedAt != nil {
		publishedAt = postgres.Timestamp(*post.PublishedAt)
	}
	var contentID pgtype.UUID
	if contentChanged {
		contentID = pgtype.UUID{Bytes: uuid.New(), Valid: true}
	}

	query := queries.New(executor)
	_, err = query.UpdatePost(ctx, queries.UpdatePostParams{
		ID:               pgtype.UUID{Bytes: post.ID, Valid: true},
		Namespace:        post.Namespace,
		Slug:             post.Slug,
		CurrentContentID: contentID,
		PublishedAt:      publishedAt,
		UpdatedAt:        postgres.Timestamp(post.UpdatedAt),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrPostNotFound
	}
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrPostAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}

	if contentChanged {
		err = query.CreatePostContent(ctx, queries.CreatePostContentParams{
			ID:        contentID,
			PostID:    pgtype.UUID{Bytes: post.ID, Valid: true},
			Title:     post.Title,
			Content:   post.Content,
			CreatedAt: postgres.Timestamp(post.UpdatedAt),
		})
		if err != nil {
			return fmt.Errorf("create post revision: %w", err)
		}
	}
	return nil
}
