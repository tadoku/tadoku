package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tadoku/tadoku/services/common/testutil/authzctx"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockPageCreateRepo struct {
	createPageFn func(ctx context.Context, page *contentdomain.Page) error
}

func (m *mockPageCreateRepo) CreatePage(ctx context.Context, page *contentdomain.Page) error {
	if m.createPageFn != nil {
		return m.createPageFn(ctx, page)
	}
	return nil
}

func adminContext() context.Context {
	return authzctx.AdminSubject("kratos-admin-id")
}

func userContext() context.Context {
	return authzctx.UserSubject("kratos-user-id")
}

func TestPageCreate_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := &mockClock{now: now}

	t.Run("creates page successfully", func(t *testing.T) {
		var savedPage *contentdomain.Page
		repo := &mockPageCreateRepo{
			createPageFn: func(ctx context.Context, page *contentdomain.Page) error {
				savedPage = page
				return nil
			},
		}

		svc := contentdomain.NewPageCreate(repo, clock)
		id := uuid.New()
		publishedAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)

		resp, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:          id,
			Namespace:   "blog",
			Slug:        "hello-world",
			Title:       "Hello World",
			HTML:        "<p>Content</p>",
			PublishedAt: &publishedAt,
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Page{
			ID:          id,
			Namespace:   "blog",
			Slug:        "hello-world",
			Title:       "Hello World",
			HTML:        "<p>Content</p>",
			PublishedAt: &publishedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, resp.Page)
		assert.Equal(t, resp.Page, savedPage)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(userContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns unauthorized when no session", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrUnauthorized)
	})

	t.Run("returns error on invalid request - missing ID", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPage)
	})

	t.Run("returns error on invalid request - missing namespace", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:    uuid.New(),
			Slug:  "hello-world",
			Title: "Hello World",
			HTML:  "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPage)
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPage)
	})

	t.Run("returns error on invalid request - uppercase slug", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "Hello-World",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPage)
	})

	t.Run("returns error on invalid request - missing title", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPage)
	})

	t.Run("returns error on invalid request - missing HTML", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPage)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockPageCreateRepo{
			createPageFn: func(ctx context.Context, page *contentdomain.Page) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("returns page already exists error", func(t *testing.T) {
		repo := &mockPageCreateRepo{
			createPageFn: func(ctx context.Context, page *contentdomain.Page) error {
				return contentdomain.ErrPageAlreadyExists
			},
		}

		svc := contentdomain.NewPageCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrPageAlreadyExists)
	})

	t.Run("creates page without published date", func(t *testing.T) {
		var savedPage *contentdomain.Page
		repo := &mockPageCreateRepo{
			createPageFn: func(ctx context.Context, page *contentdomain.Page) error {
				savedPage = page
				return nil
			},
		}

		svc := contentdomain.NewPageCreate(repo, clock)
		id := uuid.New()

		resp, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        id,
			Namespace: "blog",
			Slug:      "draft-page",
			Title:     "Draft Page",
			HTML:      "<p>Draft content</p>",
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Page{
			ID:          id,
			Namespace:   "blog",
			Slug:        "draft-page",
			Title:       "Draft Page",
			HTML:        "<p>Draft content</p>",
			PublishedAt: nil,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, resp.Page)
		assert.Equal(t, resp.Page, savedPage)
	})
}
