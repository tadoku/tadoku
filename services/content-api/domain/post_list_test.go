package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockPostListRepo struct {
	listPostsFn func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error)
}

func (m *mockPostListRepo) ListPosts(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error) {
	if m.listPostsFn != nil {
		return m.listPostsFn(ctx, namespace, includeDrafts, pageSize, page)
	}
	return nil, nil
}

func TestPostList_Execute(t *testing.T) {
	t.Run("lists posts successfully", func(t *testing.T) {
		publishedAt := time.Now()
		posts := []contentdomain.Post{
			{
				ID:          uuid.New(),
				Namespace:   "blog",
				Slug:        "post-1",
				Title:       "Post 1",
				PublishedAt: &publishedAt,
			},
			{
				ID:          uuid.New(),
				Namespace:   "blog",
				Slug:        "post-2",
				Title:       "Post 2",
				PublishedAt: &publishedAt,
			},
		}

		repo := &mockPostListRepo{
			listPostsFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error) {
				return &contentdomain.PostListResult{
					Posts:         posts,
					TotalSize:     2,
					NextPageToken: "",
				}, nil
			},
		}

		svc := contentdomain.NewPostList(repo)

		resp, err := svc.Execute(context.Background(), &contentdomain.PostListRequest{
			Namespace: "blog",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Posts) != 2 {
			t.Errorf("expected 2 posts, got %d", len(resp.Posts))
		}
		if resp.TotalSize != 2 {
			t.Errorf("expected total size 2, got %d", resp.TotalSize)
		}
	})

	t.Run("returns forbidden when non-admin requests drafts", func(t *testing.T) {
		repo := &mockPostListRepo{}
		svc := contentdomain.NewPostList(repo)

		_, err := svc.Execute(userContext(), &contentdomain.PostListRequest{
			Namespace:     "blog",
			IncludeDrafts: true,
		})

		if !errors.Is(err, contentdomain.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("allows non-admin to list without drafts", func(t *testing.T) {
		repo := &mockPostListRepo{
			listPostsFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error) {
				return &contentdomain.PostListResult{
					Posts:     []contentdomain.Post{},
					TotalSize: 0,
				}, nil
			},
		}

		svc := contentdomain.NewPostList(repo)

		_, err := svc.Execute(userContext(), &contentdomain.PostListRequest{
			Namespace:     "blog",
			IncludeDrafts: false,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("allows admin to include drafts", func(t *testing.T) {
		var capturedIncludeDrafts bool
		repo := &mockPostListRepo{
			listPostsFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error) {
				capturedIncludeDrafts = includeDrafts
				return &contentdomain.PostListResult{
					Posts:     []contentdomain.Post{},
					TotalSize: 0,
				}, nil
			},
		}

		svc := contentdomain.NewPostList(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.PostListRequest{
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

	t.Run("returns error on invalid request - missing namespace", func(t *testing.T) {
		repo := &mockPostListRepo{}
		svc := contentdomain.NewPostList(repo)

		_, err := svc.Execute(context.Background(), &contentdomain.PostListRequest{})

		if !errors.Is(err, contentdomain.ErrRequestInvalid) {
			t.Errorf("expected ErrRequestInvalid, got %v", err)
		}
	})

	t.Run("uses default page size when not specified", func(t *testing.T) {
		var capturedPageSize int
		repo := &mockPostListRepo{
			listPostsFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error) {
				capturedPageSize = pageSize
				return &contentdomain.PostListResult{
					Posts:     []contentdomain.Post{},
					TotalSize: 0,
				}, nil
			},
		}

		svc := contentdomain.NewPostList(repo)

		_, err := svc.Execute(context.Background(), &contentdomain.PostListRequest{
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
		repo := &mockPostListRepo{
			listPostsFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error) {
				capturedPageSize = pageSize
				return &contentdomain.PostListResult{
					Posts:     []contentdomain.Post{},
					TotalSize: 0,
				}, nil
			},
		}

		svc := contentdomain.NewPostList(repo)

		_, err := svc.Execute(context.Background(), &contentdomain.PostListRequest{
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

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockPostListRepo{
			listPostsFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPostList(repo)

		_, err := svc.Execute(context.Background(), &contentdomain.PostListRequest{
			Namespace: "blog",
		})

		if err != repoErr {
			t.Errorf("expected repository error, got %v", err)
		}
	})

	t.Run("returns next page token for pagination", func(t *testing.T) {
		repo := &mockPostListRepo{
			listPostsFn: func(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*contentdomain.PostListResult, error) {
				return &contentdomain.PostListResult{
					Posts:         []contentdomain.Post{},
					TotalSize:     50,
					NextPageToken: "next-page-token",
				}, nil
			},
		}

		svc := contentdomain.NewPostList(repo)

		resp, err := svc.Execute(context.Background(), &contentdomain.PostListRequest{
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
