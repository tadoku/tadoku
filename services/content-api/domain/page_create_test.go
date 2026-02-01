package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/common/domain"
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
	session := &domain.SessionToken{Role: domain.RoleAdmin}
	return context.WithValue(context.Background(), domain.CtxSessionKey, session)
}

func userContext() context.Context {
	session := &domain.SessionToken{Role: domain.RoleUser}
	return context.WithValue(context.Background(), domain.CtxSessionKey, session)
}

func TestPageCreate_Execute(t *testing.T) {
	t.Run("creates page successfully", func(t *testing.T) {
		var savedPage *contentdomain.Page
		repo := &mockPageCreateRepo{
			createPageFn: func(ctx context.Context, page *contentdomain.Page) error {
				savedPage = page
				return nil
			},
		}

		svc := contentdomain.NewPageCreate(repo)
		id := uuid.New()
		publishedAt := time.Now()

		resp, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:          id,
			Namespace:   "blog",
			Slug:        "hello-world",
			Title:       "Hello World",
			HTML:        "<p>Content</p>",
			PublishedAt: &publishedAt,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Page == nil {
			t.Fatal("expected page in response")
		}
		if resp.Page.ID != id {
			t.Errorf("expected ID %v, got %v", id, resp.Page.ID)
		}
		if resp.Page.Namespace != "blog" {
			t.Errorf("expected namespace 'blog', got %q", resp.Page.Namespace)
		}
		if resp.Page.Slug != "hello-world" {
			t.Errorf("expected slug 'hello-world', got %q", resp.Page.Slug)
		}
		if resp.Page.Title != "Hello World" {
			t.Errorf("expected title 'Hello World', got %q", resp.Page.Title)
		}
		if resp.Page.HTML != "<p>Content</p>" {
			t.Errorf("expected HTML '<p>Content</p>', got %q", resp.Page.HTML)
		}
		if resp.Page.PublishedAt == nil {
			t.Error("expected PublishedAt to be set")
		}
		if savedPage == nil {
			t.Fatal("expected page to be saved to repository")
		}
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(userContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(context.Background(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing ID", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPage) {
			t.Errorf("expected ErrInvalidPage, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing namespace", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:    uuid.New(),
			Slug:  "hello-world",
			Title: "Hello World",
			HTML:  "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPage) {
			t.Errorf("expected ErrInvalidPage, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing slug", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPage) {
			t.Errorf("expected ErrInvalidPage, got %v", err)
		}
	})

	t.Run("returns error on invalid request - uppercase slug", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "Hello-World",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPage) {
			t.Errorf("expected ErrInvalidPage, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing title", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPage) {
			t.Errorf("expected ErrInvalidPage, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing HTML", func(t *testing.T) {
		repo := &mockPageCreateRepo{}
		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
		})

		if !errors.Is(err, contentdomain.ErrInvalidPage) {
			t.Errorf("expected ErrInvalidPage, got %v", err)
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockPageCreateRepo{
			createPageFn: func(ctx context.Context, page *contentdomain.Page) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		if err != repoErr {
			t.Errorf("expected repository error, got %v", err)
		}
	})

	t.Run("returns page already exists error", func(t *testing.T) {
		repo := &mockPageCreateRepo{
			createPageFn: func(ctx context.Context, page *contentdomain.Page) error {
				return contentdomain.ErrPageAlreadyExists
			},
		}

		svc := contentdomain.NewPageCreate(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
		})

		if !errors.Is(err, contentdomain.ErrPageAlreadyExists) {
			t.Errorf("expected ErrPageAlreadyExists, got %v", err)
		}
	})

	t.Run("creates page without published date", func(t *testing.T) {
		var savedPage *contentdomain.Page
		repo := &mockPageCreateRepo{
			createPageFn: func(ctx context.Context, page *contentdomain.Page) error {
				savedPage = page
				return nil
			},
		}

		svc := contentdomain.NewPageCreate(repo)

		resp, err := svc.Execute(adminContext(), &contentdomain.PageCreateRequest{
			ID:        uuid.New(),
			Namespace: "blog",
			Slug:      "draft-page",
			Title:     "Draft Page",
			HTML:      "<p>Draft content</p>",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Page.PublishedAt != nil {
			t.Error("expected PublishedAt to be nil for draft")
		}
		if savedPage.PublishedAt != nil {
			t.Error("expected saved page PublishedAt to be nil")
		}
	})
}
