package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
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
	t.Run("creates post successfully", func(t *testing.T) {
		var savedPost *contentdomain.Post
		repo := &mockPostCreateRepo{
			createPostFn: func(ctx context.Context, post *contentdomain.Post) error {
				savedPost = post
				return nil
			},
		}

		svc := contentdomain.NewPostCreate(repo)
		id := uuid.New()
		publishedAt := time.Now()

		resp, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:          id,
			Namespace:   "blog",
			Slug:        "hello-world",
			Title:       "Hello World",
			Content:     "Post content here",
			PublishedAt: &publishedAt,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Post == nil {
			t.Fatal("expected post in response")
		}
		if resp.Post.ID != id {
			t.Errorf("expected ID %v, got %v", id, resp.Post.ID)
		}
		if resp.Post.Namespace != "blog" {
			t.Errorf("expected namespace 'blog', got %q", resp.Post.Namespace)
		}
		if resp.Post.Slug != "hello-world" {
			t.Errorf("expected slug 'hello-world', got %q", resp.Post.Slug)
		}
		if resp.Post.Title != "Hello World" {
			t.Errorf("expected title 'Hello World', got %q", resp.Post.Title)
		}
		if resp.Post.Content != "Post content here" {
			t.Errorf("expected Content 'Post content here', got %q", resp.Post.Content)
		}
		if savedPost == nil {
			t.Fatal("expected post to be saved to repository")
		}
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPostCreateRepo{}
		svc := contentdomain.NewPostCreate(repo)

		_, err := svc.Execute(userContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			Content:   "Content",
		})

		if !errors.Is(err, contentdomain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		repo := &mockPostCreateRepo{}
		svc := contentdomain.NewPostCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Title:     "Hello World",
			Content:   "Content",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPost) {
			t.Errorf("expected ErrInvalidPost, got %v", err)
		}
	})

	t.Run("returns error on invalid request - uppercase slug", func(t *testing.T) {
		repo := &mockPostCreateRepo{}
		svc := contentdomain.NewPostCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "Hello-World",
			Title:     "Hello World",
			Content:   "Content",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPost) {
			t.Errorf("expected ErrInvalidPost, got %v", err)
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockPostCreateRepo{
			createPostFn: func(ctx context.Context, post *contentdomain.Post) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPostCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			Content:   "Content",
		})

		if err != repoErr {
			t.Errorf("expected repository error, got %v", err)
		}
	})

	t.Run("returns post already exists error", func(t *testing.T) {
		repo := &mockPostCreateRepo{
			createPostFn: func(ctx context.Context, post *contentdomain.Post) error {
				return contentdomain.ErrPostAlreadyExists
			},
		}

		svc := contentdomain.NewPostCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			Content:   "Content",
		})

		if !errors.Is(err, contentdomain.ErrPostAlreadyExists) {
			t.Errorf("expected ErrPostAlreadyExists, got %v", err)
		}
	})

	t.Run("creates post without published date (draft)", func(t *testing.T) {
		var savedPost *contentdomain.Post
		repo := &mockPostCreateRepo{
			createPostFn: func(ctx context.Context, post *contentdomain.Post) error {
				savedPost = post
				return nil
			},
		}

		svc := contentdomain.NewPostCreate(repo)

		resp, err := svc.Execute(adminContext(), &contentdomain.PostCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "draft-post",
			Title:     "Draft Post",
			Content:   "Draft content",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Post.PublishedAt != nil {
			t.Error("expected PublishedAt to be nil for draft")
		}
		if savedPost.PublishedAt != nil {
			t.Error("expected saved post PublishedAt to be nil")
		}
	})
}
