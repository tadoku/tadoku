package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockPostCreateRepo struct {
	createPostFn func(ctx context.Context, post *contentdomain.Post) error
}

func (m *mockPostCreateRepo) CreatePost(ctx context.Context, post *contentdomain.Post) error {
	if m.createPostFn != nil {
		return m.createPostFn(ctx, post)
	}
	return nil
}

func TestPostCreate_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := &mockClock{now: now}

	t.Run("creates post successfully", func(t *testing.T) {
		var savedPost *contentdomain.Post
		repo := &mockPostCreateRepo{
			createPostFn: func(ctx context.Context, post *contentdomain.Post) error {
				savedPost = post
				return nil
			},
		}

		svc := contentdomain.NewPostCreate(repo, clock)
		id := uuid.New()
		publishedAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)

		resp, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:          id,
			Namespace:   "blog",
			Slug:        "hello-world",
			Title:       "Hello World",
			Content:     "Post content here",
			PublishedAt: &publishedAt,
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Post{
			ID:          id,
			Namespace:   "blog",
			Slug:        "hello-world",
			Title:       "Hello World",
			Content:     "Post content here",
			PublishedAt: &publishedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, resp.Post)
		assert.Equal(t, resp.Post, savedPost)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPostCreateRepo{}
		svc := contentdomain.NewPostCreate(repo, clock)

		_, err := svc.Execute(userContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			Content:   "Content",
		})

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		repo := &mockPostCreateRepo{}
		svc := contentdomain.NewPostCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Title:     "Hello World",
			Content:   "Content",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPost)
	})

	t.Run("returns error on invalid request - uppercase slug", func(t *testing.T) {
		repo := &mockPostCreateRepo{}
		svc := contentdomain.NewPostCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "Hello-World",
			Title:     "Hello World",
			Content:   "Content",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPost)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockPostCreateRepo{
			createPostFn: func(ctx context.Context, post *contentdomain.Post) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPostCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			Content:   "Content",
		})

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("returns post already exists error", func(t *testing.T) {
		repo := &mockPostCreateRepo{
			createPostFn: func(ctx context.Context, post *contentdomain.Post) error {
				return contentdomain.ErrPostAlreadyExists
			},
		}

		svc := contentdomain.NewPostCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			Content:   "Content",
		})

		assert.ErrorIs(t, err, contentdomain.ErrPostAlreadyExists)
	})

	t.Run("creates post without published date (draft)", func(t *testing.T) {
		var savedPost *contentdomain.Post
		repo := &mockPostCreateRepo{
			createPostFn: func(ctx context.Context, post *contentdomain.Post) error {
				savedPost = post
				return nil
			},
		}

		svc := contentdomain.NewPostCreate(repo, clock)
		id := uuid.New()

		resp, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        id,
			Namespace: "blog",
			Slug:      "draft-post",
			Title:     "Draft Post",
			Content:   "Draft content",
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Post{
			ID:          id,
			Namespace:   "blog",
			Slug:        "draft-post",
			Title:       "Draft Post",
			Content:     "Draft content",
			PublishedAt: nil,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, resp.Post)
		assert.Equal(t, resp.Post, savedPost)
	})
}
