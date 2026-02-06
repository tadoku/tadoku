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

type mockPageUpdateRepo struct {
	getPageByIDFn        func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error)
	updatePageFn         func(ctx context.Context, page *contentdomain.Page) error
	updatePageMetadataFn func(ctx context.Context, page *contentdomain.Page) error
}

func (m *mockPageUpdateRepo) GetPageByID(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
	if m.getPageByIDFn != nil {
		return m.getPageByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockPageUpdateRepo) UpdatePage(ctx context.Context, page *contentdomain.Page) error {
	if m.updatePageFn != nil {
		return m.updatePageFn(ctx, page)
	}
	return nil
}

func (m *mockPageUpdateRepo) UpdatePageMetadata(ctx context.Context, page *contentdomain.Page) error {
	if m.updatePageMetadataFn != nil {
		return m.updatePageMetadataFn(ctx, page)
	}
	return nil
}

func TestPageUpdate_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := &mockClock{now: now}

	t.Run("updates page successfully", func(t *testing.T) {
		id := uuid.New()
		createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
		existingPage := &contentdomain.Page{
			ID:        id,
			Namespace: "blog",
			Slug:      "old-slug",
			Title:     "Old Title",
			HTML:      "<p>Old content</p>",
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}

		var updatedPage *contentdomain.Page
		repo := &mockPageUpdateRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				return existingPage, nil
			},
			updatePageFn: func(ctx context.Context, page *contentdomain.Page) error {
				updatedPage = page
				return nil
			},
		}

		svc := contentdomain.NewPageUpdate(repo, clock)
		publishedAt := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)

		resp, err := svc.Execute(adminContext(), id, &contentdomain.PageUpdateRequest{
			Namespace:   "blog",
			Slug:        "new-slug",
			Title:       "New Title",
			HTML:        "<p>New content</p>",
			PublishedAt: &publishedAt,
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Page{
			ID:          id,
			Namespace:   "blog",
			Slug:        "new-slug",
			Title:       "New Title",
			HTML:        "<p>New content</p>",
			PublishedAt: &publishedAt,
			CreatedAt:   createdAt,
			UpdatedAt:   now,
		}, resp.Page)
		assert.Equal(t, resp.Page, updatedPage)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageUpdateRepo{}
		svc := contentdomain.NewPageUpdate(repo, clock)

		_, err := svc.Execute(userContext(), uuid.New(), &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Slug:      "test-slug",
			Title:     "Test Title",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		repo := &mockPageUpdateRepo{}
		svc := contentdomain.NewPageUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), uuid.New(), &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Title:     "Test Title",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidPage)
	})

	t.Run("returns error when page not found", func(t *testing.T) {
		repo := &mockPageUpdateRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				return nil, contentdomain.ErrPageNotFound
			},
		}

		svc := contentdomain.NewPageUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), uuid.New(), &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Slug:      "test-slug",
			Title:     "Test Title",
			HTML:      "<p>Content</p>",
		})

		assert.ErrorIs(t, err, contentdomain.ErrPageNotFound)
	})

	t.Run("returns repository error on update failure", func(t *testing.T) {
		id := uuid.New()
		existingPage := &contentdomain.Page{
			ID:        id,
			Namespace: "blog",
			Slug:      "old-slug",
			Title:     "Old Title",
			HTML:      "<p>Old content</p>",
		}

		repoErr := errors.New("database connection failed")
		repo := &mockPageUpdateRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				return existingPage, nil
			},
			updatePageFn: func(ctx context.Context, page *contentdomain.Page) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPageUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), id, &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Slug:      "new-slug",
			Title:     "New Title",
			HTML:      "<p>New content</p>",
		})

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("calls UpdatePage when content changes", func(t *testing.T) {
		id := uuid.New()
		existingPage := &contentdomain.Page{
			ID:        id,
			Namespace: "blog",
			Slug:      "test-page",
			Title:     "Old Title",
			HTML:      "<p>Old content</p>",
		}

		var calledUpdatePage, calledUpdateMetadata bool
		repo := &mockPageUpdateRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				return existingPage, nil
			},
			updatePageFn: func(ctx context.Context, page *contentdomain.Page) error {
				calledUpdatePage = true
				return nil
			},
			updatePageMetadataFn: func(ctx context.Context, page *contentdomain.Page) error {
				calledUpdateMetadata = true
				return nil
			},
		}

		svc := contentdomain.NewPageUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), id, &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Slug:      "test-page",
			Title:     "New Title",
			HTML:      "<p>Old content</p>",
		})

		require.NoError(t, err)
		assert.True(t, calledUpdatePage, "should call UpdatePage when title changes")
		assert.False(t, calledUpdateMetadata, "should not call UpdatePageMetadata when content changes")
	})

	t.Run("calls UpdatePageMetadata when only metadata changes", func(t *testing.T) {
		id := uuid.New()
		existingPage := &contentdomain.Page{
			ID:        id,
			Namespace: "blog",
			Slug:      "old-slug",
			Title:     "Same Title",
			HTML:      "<p>Same content</p>",
		}

		var calledUpdatePage, calledUpdateMetadata bool
		repo := &mockPageUpdateRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				return existingPage, nil
			},
			updatePageFn: func(ctx context.Context, page *contentdomain.Page) error {
				calledUpdatePage = true
				return nil
			},
			updatePageMetadataFn: func(ctx context.Context, page *contentdomain.Page) error {
				calledUpdateMetadata = true
				return nil
			},
		}

		svc := contentdomain.NewPageUpdate(repo, clock)
		publishedAt := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)

		_, err := svc.Execute(adminContext(), id, &contentdomain.PageUpdateRequest{
			Namespace:   "blog",
			Slug:        "new-slug",
			Title:       "Same Title",
			HTML:        "<p>Same content</p>",
			PublishedAt: &publishedAt,
		})

		require.NoError(t, err)
		assert.False(t, calledUpdatePage, "should not call UpdatePage when content is unchanged")
		assert.True(t, calledUpdateMetadata, "should call UpdatePageMetadata when only metadata changes")
	})

	t.Run("can unpublish a page", func(t *testing.T) {
		id := uuid.New()
		createdAt := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
		publishedAt := time.Date(2024, 1, 11, 0, 0, 0, 0, time.UTC)
		existingPage := &contentdomain.Page{
			ID:          id,
			Namespace:   "blog",
			Slug:        "published-page",
			Title:       "Published Page",
			HTML:        "<p>Content</p>",
			PublishedAt: &publishedAt,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		}

		var updatedPage *contentdomain.Page
		repo := &mockPageUpdateRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				return existingPage, nil
			},
			updatePageMetadataFn: func(ctx context.Context, page *contentdomain.Page) error {
				updatedPage = page
				return nil
			},
		}

		svc := contentdomain.NewPageUpdate(repo, clock)

		resp, err := svc.Execute(adminContext(), id, &contentdomain.PageUpdateRequest{
			Namespace:   "blog",
			Slug:        "published-page",
			Title:       "Published Page",
			HTML:        "<p>Content</p>",
			PublishedAt: nil,
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Page{
			ID:          id,
			Namespace:   "blog",
			Slug:        "published-page",
			Title:       "Published Page",
			HTML:        "<p>Content</p>",
			PublishedAt: nil,
			CreatedAt:   createdAt,
			UpdatedAt:   now,
		}, resp.Page)
		assert.Equal(t, resp.Page, updatedPage)
	})
}
