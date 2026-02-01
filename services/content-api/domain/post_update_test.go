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

type mockPostUpdateRepo struct {
	getPostByIDFn func(ctx context.Context, id uuid.UUID) (*contentdomain.Post, error)
	updatePostFn  func(ctx context.Context, post *contentdomain.Post) error
}

func (m *mockPostUpdateRepo) GetPostByID(ctx context.Context, id uuid.UUID) (*contentdomain.Post, error) {
	if m.getPostByIDFn != nil {
		return m.getPostByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockPostUpdateRepo) UpdatePost(ctx context.Context, post *contentdomain.Post) error {
	if m.updatePostFn != nil {
		return m.updatePostFn(ctx, post)
	}
	return nil
}

func TestPostUpdate_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := &mockClock{now: now}

	t.Run("updates post successfully", func(t *testing.T) {
		id := uuid.New()
		createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
		existingPost := &contentdomain.Post{
			ID:        id,
			Namespace: "blog",
			Slug:      "old-slug",
			Title:     "Old Title",
			Content:   "Old content",
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}

		var updatedPost *contentdomain.Post
		repo := &mockPostUpdateRepo{
			getPostByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Post, error) {
				return existingPost, nil
			},
			updatePostFn: func(ctx context.Context, post *contentdomain.Post) error {
				updatedPost = post
				return nil
			},
		}

		svc := contentdomain.NewPostUpdate(repo, clock)
		publishedAt := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)

		resp, err := svc.Execute(adminContext(), id, &contentdomain.PostUpdateRequest{
			Namespace:   "blog",
			Slug:        "new-slug",
			Title:       "New Title",
			Content:     "New content",
			PublishedAt: &publishedAt,
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Post{
			ID:          id,
			Namespace:   "blog",
			Slug:        "new-slug",
			Title:       "New Title",
			Content:     "New content",
			PublishedAt: &publishedAt,
			CreatedAt:   createdAt,
			UpdatedAt:   now,
		}, resp.Post)
		assert.Equal(t, resp.Post, updatedPost)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPostUpdateRepo{}
		svc := contentdomain.NewPostUpdate(repo, clock)

		_, err := svc.Execute(userContext(), uuid.New(), &contentdomain.PostUpdateRequest{
			Namespace: "blog",
			Slug:      "test-slug",
			Title:     "Test Title",
			Content:   "Content",
		})

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		repo := &mockPostUpdateRepo{}
		svc := contentdomain.NewPostUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), uuid.New(), &contentdomain.PostUpdateRequest{
			Namespace: "blog",
			Title:     "Test Title",
			Content:   "Content",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPost)
	})

	t.Run("returns error when post not found", func(t *testing.T) {
		repo := &mockPostUpdateRepo{
			getPostByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Post, error) {
				return nil, contentdomain.ErrPostNotFound
			},
		}

		svc := contentdomain.NewPostUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), uuid.New(), &contentdomain.PostUpdateRequest{
			Namespace: "blog",
			Slug:      "test-slug",
			Title:     "Test Title",
			Content:   "Content",
		})

		assert.ErrorIs(t, err, contentdomain.ErrPostNotFound)
	})

	t.Run("returns repository error on update failure", func(t *testing.T) {
		id := uuid.New()
		existingPost := &contentdomain.Post{
			ID:        id,
			Namespace: "blog",
			Slug:      "old-slug",
			Title:     "Old Title",
			Content:   "Old content",
		}

		repoErr := errors.New("database connection failed")
		repo := &mockPostUpdateRepo{
			getPostByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Post, error) {
				return existingPost, nil
			},
			updatePostFn: func(ctx context.Context, post *contentdomain.Post) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPostUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), id, &contentdomain.PostUpdateRequest{
			Namespace: "blog",
			Slug:      "new-slug",
			Title:     "New Title",
			Content:   "New content",
		})

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("can unpublish a post", func(t *testing.T) {
		id := uuid.New()
		createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
		publishedAt := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)
		existingPost := &contentdomain.Post{
			ID:          id,
			Namespace:   "blog",
			Slug:        "published-post",
			Title:       "Published Post",
			Content:     "Content",
			PublishedAt: &publishedAt,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		}

		var updatedPost *contentdomain.Post
		repo := &mockPostUpdateRepo{
			getPostByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Post, error) {
				return existingPost, nil
			},
			updatePostFn: func(ctx context.Context, post *contentdomain.Post) error {
				updatedPost = post
				return nil
			},
		}

		svc := contentdomain.NewPostUpdate(repo, clock)

		resp, err := svc.Execute(adminContext(), id, &contentdomain.PostUpdateRequest{
			Namespace:   "blog",
			Slug:        "published-post",
			Title:       "Published Post",
			Content:     "Content",
			PublishedAt: nil,
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Post{
			ID:          id,
			Namespace:   "blog",
			Slug:        "published-post",
			Title:       "Published Post",
			Content:     "Content",
			PublishedAt: nil,
			CreatedAt:   createdAt,
			UpdatedAt:   now,
		}, resp.Post)
		assert.Equal(t, resp.Post, updatedPost)
	})
}
