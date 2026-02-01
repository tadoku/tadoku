package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockPageListRepo struct {
	listPagesFn func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PageListResult, error)
}

func (m *mockPageListRepo) ListPages(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PageListResult, error) {
	if m.listPagesFn != nil {
		return m.listPagesFn(ctx, namespace, includeDrafts, pageSize, page)
	}
	return nil, nil
}

func TestPageList_Execute(t *testing.T) {
	t.Run("lists pages successfully", func(t *testing.T) {
		publishedAt := time.Now()
		pages := []contentdomain.Page{
			{
				ID:          uuid.New(),
				Namespace:   "blog",
				Slug:        "page-1",
				Title:       "Page 1",
				PublishedAt: &publishedAt,
			},
			{
				ID:          uuid.New(),
				Namespace:   "blog",
				Slug:        "page-2",
				Title:       "Page 2",
				PublishedAt: &publishedAt,
			},
		}

		repo := &mockPageListRepo{
			listPagesFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PageListResult, error) {
				return &contentdomain.PageListResult{
					Pages:         pages,
					TotalSize:     2,
					NextPageToken: "",
				}, nil
			},
		}

		svc := contentdomain.NewPageList(repo)

		resp, err := svc.Execute(adminContext(), &contentdomain.PageListRequest{
			Namespace: "blog",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Pages) != 2 {
			t.Errorf("expected 2 pages, got %d", len(resp.Pages))
		}
		if resp.TotalSize != 2 {
			t.Errorf("expected total size 2, got %d", resp.TotalSize)
		}
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageListRepo{}
		svc := contentdomain.NewPageList(repo)

		_, err := svc.Execute(userContext(), &contentdomain.PageListRequest{
			Namespace: "blog",
		})

		if !errors.Is(err, contentdomain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPageListRepo{}
		svc := contentdomain.NewPageList(repo)

		_, err := svc.Execute(context.Background(), &contentdomain.PageListRequest{
			Namespace: "blog",
		})

		if !errors.Is(err, contentdomain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("returns error on invalid request - missing namespace", func(t *testing.T) {
		repo := &mockPageListRepo{}
		svc := contentdomain.NewPageList(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageListRequest{})

		if !errors.Is(err, contentdomain.ErrRequestInvalid) {
			t.Errorf("expected ErrRequestInvalid, got %v", err)
		}
	})

	t.Run("uses default page size when not specified", func(t *testing.T) {
		var capturedPageSize int
		repo := &mockPageListRepo{
			listPagesFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PageListResult, error) {
				capturedPageSize = pageSize
				return &contentdomain.PageListResult{
					Pages:     []contentdomain.Page{},
					TotalSize: 0,
				}, nil
			},
		}

		svc := contentdomain.NewPageList(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageListRequest{
			Namespace: "blog",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedPageSize != 10 {
			t.Errorf("expected default page size 10, got %d", capturedPageSize)
		}
	})

	t.Run("caps page size at 100", func(t *testing.T) {
		var capturedPageSize int
		repo := &mockPageListRepo{
			listPagesFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PageListResult, error) {
				capturedPageSize = pageSize
				return &contentdomain.PageListResult{
					Pages:     []contentdomain.Page{},
					TotalSize: 0,
				}, nil
			},
		}

		svc := contentdomain.NewPageList(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageListRequest{
			Namespace: "blog",
			PageSize:  500,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedPageSize != 100 {
			t.Errorf("expected capped page size 100, got %d", capturedPageSize)
		}
	})

	t.Run("passes include drafts flag", func(t *testing.T) {
		var capturedIncludeDrafts bool
		repo := &mockPageListRepo{
			listPagesFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PageListResult, error) {
				capturedIncludeDrafts = includeDrafts
				return &contentdomain.PageListResult{
					Pages:     []contentdomain.Page{},
					TotalSize: 0,
				}, nil
			},
		}

		svc := contentdomain.NewPageList(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageListRequest{
			Namespace:     "blog",
			IncludeDrafts: true,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !capturedIncludeDrafts {
			t.Error("expected includeDrafts to be true")
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockPageListRepo{
			listPagesFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PageListResult, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPageList(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PageListRequest{
			Namespace: "blog",
		})

		if err != repoErr {
			t.Errorf("expected repository error, got %v", err)
		}
	})

	t.Run("returns next page token for pagination", func(t *testing.T) {
		repo := &mockPageListRepo{
			listPagesFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PageListResult, error) {
				return &contentdomain.PageListResult{
					Pages:         []contentdomain.Page{},
					TotalSize:     50,
					NextPageToken: "next-page-token",
				}, nil
			},
		}

		svc := contentdomain.NewPageList(repo)

		resp, err := svc.Execute(adminContext(), &contentdomain.PageListRequest{
			Namespace: "blog",
			PageSize:  10,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.NextPageToken != "next-page-token" {
			t.Errorf("expected next page token 'next-page-token', got %q", resp.NextPageToken)
		}
	})
}
