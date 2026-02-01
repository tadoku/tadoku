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

type mockPageFindRepo struct {
	findPageBySlugFn func(ctx context.Context, namespace, slug string) (*contentdomain.Page, error)
}

func (m *mockPageFindRepo) FindPageBySlug(ctx context.Context, namespace, slug string) (*contentdomain.Page, error) {
	if m.findPageBySlugFn != nil {
		return m.findPageBySlugFn(ctx, namespace, slug)
	}
	return nil, nil
}

type mockClock struct {
	now time.Time
}

func (c *mockClock) Now() time.Time {
	return c.now
}

func TestPageFind_Execute(t *testing.T) {
	t.Run("finds published page successfully", func(t *testing.T) {
		now := time.Now()
		publishedAt := now.Add(-time.Hour)
		page := &contentdomain.Page{
			ID:          uuid.New(),
			Namespace:   "blog",
			Slug:        "hello-world",
			Title:       "Hello World",
			HTML:        "<p>Content</p>",
			PublishedAt: &publishedAt,
		}

		repo := &mockPageFindRepo{
			findPageBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Page, error) {
				if namespace == "blog" && slug == "hello-world" {
					return page, nil
				}
				return nil, contentdomain.ErrPageNotFound
			},
		}

		clock := &mockClock{now: now}
		svc := contentdomain.NewPageFind(repo, clock)

		resp, err := svc.Execute(context.Background(), &contentdomain.PageFindRequest{
			Namespace: "blog",
			Slug:      "hello-world",
		})

		require.NoError(t, err)
		assert.Equal(t, "Hello World", resp.Page.Title)
	})

	t.Run("returns not found for unpublished page", func(t *testing.T) {
		now := time.Now()
		page := &contentdomain.Page{
			ID:          uuid.New(),
			Namespace:   "blog",
			Slug:        "draft",
			Title:       "Draft Page",
			HTML:        "<p>Content</p>",
			PublishedAt: nil,
		}

		repo := &mockPageFindRepo{
			findPageBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Page, error) {
				return page, nil
			},
		}

		clock := &mockClock{now: now}
		svc := contentdomain.NewPageFind(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PageFindRequest{
			Namespace: "blog",
			Slug:      "draft",
		})

		assert.ErrorIs(t, err, contentdomain.ErrPageNotFound)
	})

	t.Run("returns not found for future published page", func(t *testing.T) {
		now := time.Now()
		futurePublish := now.Add(time.Hour)
		page := &contentdomain.Page{
			ID:          uuid.New(),
			Namespace:   "blog",
			Slug:        "scheduled",
			Title:       "Scheduled Page",
			HTML:        "<p>Content</p>",
			PublishedAt: &futurePublish,
		}

		repo := &mockPageFindRepo{
			findPageBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Page, error) {
				return page, nil
			},
		}

		clock := &mockClock{now: now}
		svc := contentdomain.NewPageFind(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PageFindRequest{
			Namespace: "blog",
			Slug:      "scheduled",
		})

		assert.ErrorIs(t, err, contentdomain.ErrPageNotFound)
	})

	t.Run("returns error on invalid request - missing namespace", func(t *testing.T) {
		clock := &mockClock{now: time.Now()}
		svc := contentdomain.NewPageFind(nil, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PageFindRequest{
			Slug: "hello-world",
		})

		assert.ErrorIs(t, err, contentdomain.ErrRequestInvalid)
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		clock := &mockClock{now: time.Now()}
		svc := contentdomain.NewPageFind(nil, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PageFindRequest{
			Namespace: "blog",
		})

		assert.ErrorIs(t, err, contentdomain.ErrRequestInvalid)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockPageFindRepo{
			findPageBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Page, error) {
				return nil, repoErr
			},
		}

		clock := &mockClock{now: time.Now()}
		svc := contentdomain.NewPageFind(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PageFindRequest{
			Namespace: "blog",
			Slug:      "hello-world",
		})

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("returns page not found from repository", func(t *testing.T) {
		repo := &mockPageFindRepo{
			findPageBySlugFn: func(ctx context.Context, namespace, slug string) (*contentdomain.Page, error) {
				return nil, contentdomain.ErrPageNotFound
			},
		}

		clock := &mockClock{now: time.Now()}
		svc := contentdomain.NewPageFind(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PageFindRequest{
			Namespace: "blog",
			Slug:      "nonexistent",
		})

		assert.ErrorIs(t, err, contentdomain.ErrPageNotFound)
	})
}
