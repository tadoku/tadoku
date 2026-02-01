package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
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
	clock := &mockClock{now: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)}

	t.Run("updates post successfully", func(t *testing.T) {
		id := uuid.New()
		existingPost := &contentdomain.Post{
			ID:        id,
			Namespace: "blog",
			Slug:      "old-slug",
			Title:     "Old Title",
			Content:   "Old content",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
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
		publishedAt := time.Now()

		resp, err := svc.Execute(adminContext(), id, &contentdomain.PostUpdateRequest{
			Namespace:   "blog",
			Slug:        "new-slug",
			Title:       "New Title",
			Content:     "New content",
			PublishedAt: &publishedAt,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Post.Slug != "new-slug" {
			t.Errorf("expected slug 'new-slug', got %q", resp.Post.Slug)
		}
		if resp.Post.Title != "New Title" {
			t.Errorf("expected title 'New Title', got %q", resp.Post.Title)
		}
		if resp.Post.Content != "New content" {
			t.Errorf("expected Content 'New content', got %q", resp.Post.Content)
		}
		if updatedPost == nil {
			t.Fatal("expected post to be updated in repository")
		}
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

		if !errors.Is(err, contentdomain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		repo := &mockPostUpdateRepo{}
		svc := contentdomain.NewPostUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), uuid.New(), &contentdomain.PostUpdateRequest{
			Namespace: "blog",
			Title:     "Test Title",
			Content:   "Content",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPost) {
			t.Errorf("expected ErrInvalidPost, got %v", err)
		}
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

		if !errors.Is(err, contentdomain.ErrPostNotFound) {
			t.Errorf("expected ErrPostNotFound, got %v", err)
		}
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

		if err != repoErr {
			t.Errorf("expected repository error, got %v", err)
		}
	})

	t.Run("can unpublish a post", func(t *testing.T) {
		id := uuid.New()
		publishedAt := time.Now().Add(-time.Hour)
		existingPost := &contentdomain.Post{
			ID:          id,
			Namespace:   "blog",
			Slug:        "published-post",
			Title:       "Published Post",
			Content:     "Content",
			PublishedAt: &publishedAt,
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

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Post.PublishedAt != nil {
			t.Error("expected PublishedAt to be nil after unpublishing")
		}
		if updatedPost.PublishedAt != nil {
			t.Error("expected saved post PublishedAt to be nil")
		}
	})
}
