package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/tadoku/tadoku/services/content-api/domain"
)

// PostRepository implements all post-related domain interfaces:
// - domain.PostCreateRepository
// - domain.PostUpdateRepository
// - domain.PostFindRepository
// - domain.PostListRepository
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

// CreatePost implements domain.PostCreateRepository
func (r *PostRepository) CreatePost(ctx context.Context, post *domain.Post) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not create post: %w", err)
	}

	postContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	_, err = qtx.CreatePost(ctx, CreatePostParams{
		ID:               post.ID,
		Namespace:        post.Namespace,
		Slug:             post.Slug,
		CurrentContentID: postContentID,
		PublishedAt:      NewNullTime(post.PublishedAt),
	})
	if err != nil {
		_ = tx.Rollback()

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrPostAlreadyExists
		}

		return fmt.Errorf("could not create post: %w", err)
	}

	_, err = qtx.CreatePostContent(ctx, CreatePostContentParams{
		ID:      postContentID,
		PostID:  post.ID,
		Title:   post.Title,
		Content: post.Content,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not create post: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not create post: %w", err)
	}

	return nil
}

// GetPostByID implements domain.PostUpdateRepository
func (r *PostRepository) GetPostByID(ctx context.Context, id uuid.UUID, namespace string) (*domain.Post, error) {
	post, err := r.q.FindPostByID(ctx, FindPostByIDParams{ID: id, Namespace: namespace})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}
		return nil, fmt.Errorf("could not find post: %w", err)
	}

	return &domain.Post{
		ID:          post.ID,
		Namespace:   post.Namespace,
		Slug:        post.Slug,
		Title:       post.Title,
		Content:     post.Content,
		PublishedAt: NewTimeFromNullTime(post.PublishedAt),
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}, nil
}

// UpdatePost implements domain.PostUpdateRepository
func (r *PostRepository) UpdatePost(ctx context.Context, post *domain.Post) error {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not update post: %w", err)
	}

	postContentID := uuid.New()

	qtx := r.q.WithTx(tx)

	_, err = qtx.UpdatePost(ctx, UpdatePostParams{
		ID:               post.ID,
		Slug:             post.Slug,
		CurrentContentID: postContentID,
		PublishedAt:      NewNullTime(post.PublishedAt),
	})
	if err != nil {
		_ = tx.Rollback()

		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrPostNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrPostAlreadyExists
		}

		return fmt.Errorf("could not update post: %w", err)
	}

	_, err = qtx.CreatePostContent(ctx, CreatePostContentParams{
		ID:      postContentID,
		PostID:  post.ID,
		Title:   post.Title,
		Content: post.Content,
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not update post: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not update post: %w", err)
	}

	return nil
}

// UpdatePostMetadata implements domain.PostUpdateRepository
func (r *PostRepository) UpdatePostMetadata(ctx context.Context, post *domain.Post) error {
	_, err := r.q.UpdatePostMetadata(ctx, UpdatePostMetadataParams{
		ID:          post.ID,
		Slug:        post.Slug,
		PublishedAt: NewNullTime(post.PublishedAt),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrPostNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.ErrPostAlreadyExists
		}

		return fmt.Errorf("could not update post metadata: %w", err)
	}

	return nil
}

// DeletePost implements domain.PostDeleteRepository
func (r *PostRepository) DeletePost(ctx context.Context, id uuid.UUID, namespace string) error {
	if err := r.q.DeletePost(ctx, DeletePostParams{ID: id, Namespace: namespace}); err != nil {
		return fmt.Errorf("could not delete post: %w", err)
	}
	return nil
}

// ListPostVersions implements domain.PostVersionListRepository
func (r *PostRepository) ListPostVersions(ctx context.Context, postID uuid.UUID) ([]domain.PostVersion, error) {
	rows, err := r.q.ListPostVersions(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("could not list post versions: %w", err)
	}

	versions := make([]domain.PostVersion, len(rows))
	for i, row := range rows {
		versions[i] = domain.PostVersion{
			ID:        row.ID,
			Version:   i + 1,
			Title:     row.Title,
			CreatedAt: row.CreatedAt,
		}
	}

	return versions, nil
}

// GetPostVersion implements domain.PostVersionListRepository
func (r *PostRepository) GetPostVersion(ctx context.Context, postID uuid.UUID, contentID uuid.UUID) (*domain.PostVersion, error) {
	row, err := r.q.GetPostVersion(ctx, GetPostVersionParams{
		ID:     contentID,
		PostID: postID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}
		return nil, fmt.Errorf("could not get post version: %w", err)
	}

	return &domain.PostVersion{
		ID:        row.ID,
		Title:     row.Title,
		Content:   row.Content,
		CreatedAt: row.CreatedAt,
	}, nil
}

// FindPostBySlug implements domain.PostFindRepository
func (r *PostRepository) FindPostBySlug(ctx context.Context, namespace, slug string) (*domain.Post, error) {
	post, err := r.q.FindPostBySlug(ctx, FindPostBySlugParams{
		Namespace: namespace,
		Slug:      slug,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}

		return nil, fmt.Errorf("could not find post: %w", err)
	}

	return &domain.Post{
		ID:          post.ID,
		Namespace:   post.Namespace,
		Slug:        post.Slug,
		Title:       post.Title,
		Content:     post.Content,
		PublishedAt: NewTimeFromNullTime(post.PublishedAt),
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
	}, nil
}

// ListPosts implements domain.PostListRepository
func (r *PostRepository) ListPosts(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*domain.PostListResult, error) {
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list posts: %w", err)
	}

	qtx := r.q.WithTx(tx)

	meta, err := qtx.PostsMetadata(ctx, PostsMetadataParams{
		IncludeDrafts: includeDrafts,
		Namespace:     namespace,
	})
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list posts: %w", err)
	}

	posts, err := qtx.ListPosts(ctx, ListPostsParams{
		StartFrom:     int32(page * pageSize),
		PageSize:      int32(pageSize),
		Namespace:     namespace,
		IncludeDrafts: includeDrafts,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		return nil, fmt.Errorf("could not list posts: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("could not list posts: %w", err)
	}

	result := make([]domain.Post, len(posts))
	for i, p := range posts {
		result[i] = domain.Post{
			ID:          p.ID,
			Namespace:   p.Namespace,
			Slug:        p.Slug,
			Title:       p.Title,
			Content:     p.Content,
			PublishedAt: NewTimeFromNullTime(p.PublishedAt),
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		}
	}

	nextPageToken := ""
	if (page*pageSize)+pageSize < int(meta.TotalSize) {
		nextPageToken = fmt.Sprint(page + 1)
	}

	return &domain.PostListResult{
		Posts:         result,
		TotalSize:     int(meta.TotalSize),
		NextPageToken: nextPageToken,
	}, nil
}
