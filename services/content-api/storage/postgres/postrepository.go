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
	"github.com/tadoku/tadoku/services/content-api/domain/postquery"
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

	postContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	if _, err := qtx.CreatePost(ctx, CreatePostParams{
		req.ID,
		req.Namespace,
		req.Slug,
		postContentID,
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
		ID:      postContentID,
		PostID:  req.ID,
		Title:   req.Title,
		Content: req.Content,
	}); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create post: %w", err)
	}

	post, err := qtx.FindPostBySlug(ctx, FindPostBySlugParams{
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
		ID:          post.ID,
		Namespace:   post.Namespace,
		Slug:        post.Slug,
		Title:       post.Title,
		Content:     post.Content,
		PublishedAt: NewTimeFromNullTime(post.PublishedAt),
	}, nil
}

func (r *PostRepository) UpdatePost(ctx context.Context, id uuid.UUID, req *postcommand.PostUpdateRequest) (*postcommand.PostUpdateResponse, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create post: %w", err)
	}

	postContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	_, err = qtx.UpdatePost(ctx, UpdatePostParams{
		ID:               id,
		Slug:             req.Slug,
		CurrentContentID: postContentID,
		PublishedAt:      NewNullTime(req.PublishedAt),
	})
	if err != nil {
		_ = tx.Rollback()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, postcommand.ErrPostAlreadyExists
		}

		return nil, fmt.Errorf("could not create post: %w", err)
	}

	_, err = qtx.CreatePostContent(ctx, CreatePostContentParams{
		ID:      postContentID,
		PostID:  id,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not create post: %w", err)
	}

	post, err := qtx.FindPostBySlug(ctx, FindPostBySlugParams{
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

	return &postcommand.PostUpdateResponse{
		ID:          post.ID,
		Slug:        post.Slug,
		Title:       post.Title,
		Content:     post.Content,
		PublishedAt: NewTimeFromNullTime(post.PublishedAt),
	}, nil
}

func (r *PostRepository) FindBySlug(ctx context.Context, namespace, slug string) (*postquery.PostFindResponse, error) {
	post, err := r.q.FindPostBySlug(ctx, FindPostBySlugParams{
		Namespace: namespace,
		Slug:      slug,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postquery.ErrPostNotFound
		}

		return nil, fmt.Errorf("could not find post: %w", err)
	}

	return &postquery.PostFindResponse{
		ID:          post.ID,
		Slug:        post.Slug,
		Title:       post.Title,
		Content:     post.Content,
		PublishedAt: NewTimeFromNullTime(post.PublishedAt),
	}, nil
}
