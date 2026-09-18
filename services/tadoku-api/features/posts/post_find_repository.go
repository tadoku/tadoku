package posts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (r *PostsRepository) FindPostBySlug(ctx context.Context, namespace, slug string) (*Post, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindPostBySlug(ctx, queries.FindPostBySlugParams{
		Namespace: namespace,
		Slug:      slug,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find post by slug: %w", err)
	}
	return postFromFindRow(row), nil
}

func (r *PostsRepository) FindPostByID(ctx context.Context, namespace string, id uuid.UUID) (*Post, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, err
	}

	row, err := queries.New(executor).FindPostByID(ctx, queries.FindPostByIDParams{
		Namespace: namespace,
		ID:        pgtype.UUID{Bytes: id, Valid: true},
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPostNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find post by ID: %w", err)
	}
	return postFromFindRow(queries.FindPostBySlugRow(row)), nil
}

func postFromFindRow(row queries.FindPostBySlugRow) *Post {
	var publishedAt *time.Time
	if row.PublishedAt.Valid {
		publishedAt = &row.PublishedAt.Time
	}
	return &Post{
		ID:          uuid.UUID(row.ID.Bytes),
		Namespace:   row.Namespace,
		Slug:        row.Slug,
		Title:       row.Title,
		Content:     row.Content,
		PublishedAt: publishedAt,
		CreatedAt:   &row.CreatedAt.Time,
		UpdatedAt:   &row.UpdatedAt.Time,
	}
}
