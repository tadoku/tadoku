package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockPageUpdateRepo struct {
	getPageByIDFn func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error)
	updatePageFn  func(ctx context.Context, page *contentdomain.Page) error
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

func TestPageUpdate_Execute(t *testing.T) {
	t.Run("updates page successfully", func(t *testing.T) {
		id := uuid.New()
		existingPage := &contentdomain.Page{
			ID:        id,
			Namespace: "blog",
			Slug:      "old-slug",
			Title:     "Old Title",
			HTML:      "<p>Old content</p>",
			CreatedAt: time.Now().Add(-time.Hour),
			UpdatedAt: time.Now().Add(-time.Hour),
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

		svc := contentdomain.NewPageUpdate(repo)
		publishedAt := time.Now()

		resp, err := svc.Execute(adminContext(), id, &contentdomain.PageUpdateRequest{
			Namespace:   "blog",
			Slug:        "new-slug",
			Title:       "New Title",
			HTML:        "<p>New content</p>",
			PublishedAt: &publishedAt,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Page.Slug != "new-slug" {
			t.Errorf("expected slug 'new-slug', got %q", resp.Page.Slug)
		}
		if resp.Page.Title != "New Title" {
			t.Errorf("expected title 'New Title', got %q", resp.Page.Title)
		}
		if resp.Page.HTML != "<p>New content</p>" {
			t.Errorf("expected HTML '<p>New content</p>', got %q", resp.Page.HTML)
		}
		if updatedPage == nil {
			t.Fatal("expected page to be updated in repository")
		}
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageUpdateRepo{}
		svc := contentdomain.NewPageUpdate(repo)

		_, err := svc.Execute(userContext(), uuid.New(), &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Slug:      "test-slug",
			Title:     "Test Title",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		repo := &mockPageUpdateRepo{}
		svc := contentdomain.NewPageUpdate(repo)

		_, err := svc.Execute(adminContext(), uuid.New(), &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Title:     "Test Title",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPage) {
			t.Errorf("expected ErrInvalidPage, got %v", err)
		}
	})

	t.Run("returns error when page not found", func(t *testing.T) {
		repo := &mockPageUpdateRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				return nil, contentdomain.ErrPageNotFound
			},
		}

		svc := contentdomain.NewPageUpdate(repo)

		_, err := svc.Execute(adminContext(), uuid.New(), &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Slug:      "test-slug",
			Title:     "Test Title",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrPageNotFound) {
			t.Errorf("expected ErrPageNotFound, got %v", err)
		}
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

		svc := contentdomain.NewPageUpdate(repo)

		_, err := svc.Execute(adminContext(), id, &contentdomain.PageUpdateRequest{
			Namespace: "blog",
			Slug:      "new-slug",
			Title:     "New Title",
			HTML:      "<p>New content</p>",
		})

		if err != repoErr {
			t.Errorf("expected repository error, got %v", err)
		}
	})

	t.Run("can unpublish a page", func(t *testing.T) {
		id := uuid.New()
		publishedAt := time.Now().Add(-time.Hour)
		existingPage := &contentdomain.Page{
			ID:          id,
			Namespace:   "blog",
			Slug:        "published-page",
			Title:       "Published Page",
			HTML:        "<p>Content</p>",
			PublishedAt: &publishedAt,
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

		svc := contentdomain.NewPageUpdate(repo)

		resp, err := svc.Execute(adminContext(), id, &contentdomain.PageUpdateRequest{
			Namespace:   "blog",
			Slug:        "published-page",
			Title:       "Published Page",
			HTML:        "<p>Content</p>",
			PublishedAt: nil,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Page.PublishedAt != nil {
			t.Error("expected PublishedAt to be nil after unpublishing")
		}
		if updatedPage.PublishedAt != nil {
			t.Error("expected saved page PublishedAt to be nil")
		}
	})
}
