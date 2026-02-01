package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockPostFindRepo struct {
	findPostBySlugFn func(ctx context.Context, namespace, slug string) (*contentdomain.Post, error)
}

func (m *mockPostFindRepo) FindPostBySlug(ctx context.Context, namespace, slug string) (*contentdomain.Post, error) {
	if m.findPostBySlugFn != nil {
		return m.findPostBySlugFn(ctx, namespace, slug)
	}
	return nil, nil
}

func TestPostFind_Execute(t *testing.T) {
	t.Run("finds published post successfully", func(t *testing.T) {
		now := time.Now()
		publishedAt := now.Add(-time.Hour)
		post := &contentdomain.Post{
			ID:          uuid.New(),
			Namespace:   "blog",
			Slug:        "hello-world",
			Title:       "Hello World",
			Content:     "Post content",
			PublishedAt: &publishedAt,
		}

		repo := &mockPostFindRepo{
			findPostBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Post, error) {
				if namespace == "blog" && slug == "hello-world" {
					return post, nil
				}
				return nil, contentdomain.ErrPostNotFound
			},
		}

		clock := &mockClock{now: now}
		svc := contentdomain.NewPostFind(repo, clock)

		resp, err := svc.Execute(context.Background(), &contentdomain.PostFindRequest{
			Namespace: "blog",
			Slug:      "hello-world",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Post.Title != "Hello World" {
			t.Errorf("expected title 'Hello World', got %q", resp.Post.Title)
		}
	})

	t.Run("returns not found for unpublished post", func(t *testing.T) {
		now := time.Now()
		post := &contentdomain.Post{
			ID:          uuid.New(),
			Namespace:   "blog",
			Slug:        "draft",
			Title:       "Draft Post",
			Content:     "Content",
			PublishedAt: nil,
		}

		repo := &mockPostFindRepo{
			findPostBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Post, error) {
				return post, nil
			},
		}

		clock := &mockClock{now: now}
		svc := contentdomain.NewPostFind(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PostFindRequest{
			Namespace: "blog",
			Slug:      "draft",
		})

		if !errors.Is(err, contentdomain.ErrPostNotFound) {
			t.Errorf("expected ErrPostNotFound, got %v", err)
		}
	})

	t.Run("returns not found for future published post", func(t *testing.T) {
		now := time.Now()
		futurePublish := now.Add(time.Hour)
		post := &contentdomain.Post{
			ID:          uuid.New(),
			Namespace:   "blog",
			Slug:        "scheduled",
			Title:       "Scheduled Post",
			Content:     "Content",
			PublishedAt: &futurePublish,
		}

		repo := &mockPostFindRepo{
			findPostBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Post, error) {
				return post, nil
			},
		}

		clock := &mockClock{now: now}
		svc := contentdomain.NewPostFind(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PostFindRequest{
			Namespace: "blog",
			Slug:      "scheduled",
		})

		if !errors.Is(err, contentdomain.ErrPostNotFound) {
			t.Errorf("expected ErrPostNotFound, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing namespace", func(t *testing.T) {
		clock := &mockClock{now: time.Now()}
		svc := contentdomain.NewPostFind(nil, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PostFindRequest{
			Slug: "hello-world",
		})

		if !errors.Is(err, contentdomain.ErrRequestInvalid) {
			t.Errorf("expected ErrRequestInvalid, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		clock := &mockClock{now: time.Now()}
		svc := contentdomain.NewPostFind(nil, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PostFindRequest{
			Namespace: "blog",
		})

		if !errors.Is(err, contentdomain.ErrRequestInvalid) {
			t.Errorf("expected ErrRequestInvalid, got %v", err)
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockPostFindRepo{
			findPostBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Post, error) {
				return nil, repoErr
			},
		}

		clock := &mockClock{now: time.Now()}
		svc := contentdomain.NewPostFind(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PostFindRequest{
			Namespace: "blog",
			Slug:      "hello-world",
		})

		if err != repoErr {
			t.Errorf("expected repository error, got %v", err)
		}
	})

	t.Run("returns post not found from repository", func(t *testing.T) {
		repo := &mockPostFindRepo{
			findPostBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Post, error) {
				return nil, contentdomain.ErrPostNotFound
			},
		}

		clock := &mockClock{now: time.Now()}
		svc := contentdomain.NewPostFind(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PostFindRequest{
			Namespace: "blog",
			Slug:      "nonexistent",
		})

		if !errors.Is(err, contentdomain.ErrPostNotFound) {
			t.Errorf("expected ErrPostNotFound, got %v", err)
		}
	})
}
