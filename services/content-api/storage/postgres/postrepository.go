package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/tadoku/tadoku/services/content-api/domain/postcommand"
)

type PostRepository struct {
	psql *sql.DB
	q    *Queries
}

func NewPostRepository(psql *sql.DB) *PostRepository {
	return &PostRepository{
		psql: psql,
		q:    &Queries{psql},
	}
}

func (r *PostRepository) CreatePost(ctx context.Context, req *postcommand.PostCreateRequest) (*postcommand.PostCreateResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create post: %w", err)
	}

	pageContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	if _, err := qtx.CreatePost(ctx, CreatePostParams{
		req.ID,
		req.Namespace,
		req.Slug,
		pageContentID,
		NewNullTime(req.PublishedAt),
	}); err != nil {
		_ = tx.Rollback()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, postcommand.ErrPostAlreadyExists
		}

		return nil, fmt.Errorf("could not create post: %w", err)
	}

	if _, err := qtx.CreatePostContent(ctx, CreatePostContentParams{
		ID:      pageContentID,
		PostID:  req.ID,
		Title:   req.Title,
		Content: req.Content,
	}); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create post: %w", err)
	}

	page, err := qtx.FindPostBySlug(ctx, FindPostBySlugParams{
		Namespace: req.Namespace,
		Slug:      req.Slug,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create post: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not create post: %w", err)
	}

	return &postcommand.PostCreateResponse{
		ID:          page.ID,
		Namespace:   page.Namespace,
		Slug:        page.Slug,
		Title:       page.Title,
		Content:     page.Content,
		PublishedAt: NewTimeFromNullTime(page.PublishedAt),
	}, nil
}
