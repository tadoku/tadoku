package posts

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type PostsRepository struct {
	db *pgxpool.Pool
}

func NewPostsRepository(db *pgxpool.Pool) *PostsRepository {
	return &PostsRepository{
		db: db,
	}
}

func (r *PostsRepository) CreatePost(ctx context.Context, item *Post, contentID uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var publishedAt pgtype.Timestamp
	if item.PublishedAt != nil {
		publishedAt = postgres.Timestamp(*item.PublishedAt)
	}
	err = queries.New(executor).CreatePost(ctx, queries.CreatePostParams{
		ID:               pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace:        item.Namespace,
		Slug:             item.Slug,
		CurrentContentID: pgtype.UUID{Bytes: contentID, Valid: true},
		PublishedAt:      publishedAt,
		CreatedAt:        postgres.Timestamp(*item.CreatedAt),
		UpdatedAt:        postgres.Timestamp(*item.UpdatedAt),
	})
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) && pgError.Code == pgerrcode.UniqueViolation {
		return ErrPostAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}

	return nil
}

func (r *PostsRepository) CreatePostContent(ctx context.Context, postID, contentID uuid.UUID, title, content string, createdAt time.Time) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	err = queries.New(executor).CreatePostContent(ctx, queries.CreatePostContentParams{
		ID:        pgtype.UUID{Bytes: contentID, Valid: true},
		PostID:    pgtype.UUID{Bytes: postID, Valid: true},
		Title:     title,
		Content:   content,
		CreatedAt: postgres.Timestamp(createdAt),
	})
	if err != nil {
		return fmt.Errorf("create post content: %w", err)
	}

	return nil
}

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

func (r *PostsRepository) ListPosts(ctx context.Context, namespace string, includeDrafts bool, cutoff time.Time, limit int32, offset int64) ([]Post, int, error) {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return nil, 0, err
	}

	rows, err := queries.New(executor).ListPosts(ctx, queries.ListPostsParams{
		Namespace:     namespace,
		IncludeDrafts: includeDrafts,
		Cutoff:        postgres.Timestamp(cutoff),
		ResultLimit:   limit,
		StartFrom:     offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("list posts: %w", err)
	}

	result := make([]Post, 0, len(rows))
	var total int
	for _, row := range rows {
		total = int(row.TotalSize)
		if !row.ID.Valid {
			continue
		}

		var publishedAt *time.Time
		if row.PublishedAt.Valid {
			publishedAt = &row.PublishedAt.Time
		}

		result = append(result, Post{
			ID:          uuid.UUID(row.ID.Bytes),
			Namespace:   row.Namespace.String,
			Slug:        row.Slug.String,
			Title:       row.Title.String,
			Content:     row.Content.String,
			PublishedAt: publishedAt,
			CreatedAt:   &row.CreatedAt.Time,
			UpdatedAt:   &row.UpdatedAt.Time,
		})
	}

	return result, total, nil
}

func (r *PostsRepository) UpdatePost(ctx context.Context, item *Post, contentID *uuid.UUID) error {
	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}

	var publishedAt pgtype.Timestamp
	if item.PublishedAt != nil {
		publishedAt = postgres.Timestamp(*item.PublishedAt)
	}
	_, err = queries.New(executor).UpdatePost(ctx, queries.UpdatePostParams{
		ID:               pgtype.UUID{Bytes: item.ID, Valid: true},
		Namespace:        item.Namespace,
		Slug:             item.Slug,
		CurrentContentID: postgres.NullableUUID(contentID),
		PublishedAt:      publishedAt,
		UpdatedAt:        postgres.Timestamp(*item.UpdatedAt),
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

	return nil
}

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
