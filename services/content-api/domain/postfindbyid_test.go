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

type mockPostFindByIDRepo struct {
	getPostByIDFn func(ctx context.Context, id uuid.UUID, namespace string) (*contentdomain.Post, error)
}

func (m *mockPostFindByIDRepo) GetPostByID(ctx context.Context, id uuid.UUID, namespace string) (*contentdomain.Post, error) {
	return m.getPostByIDFn(ctx, id, namespace)
}

func TestPostFindByID_Execute(t *testing.T) {
	postID := uuid.New()

	t.Run("finds post successfully", func(t *testing.T) {
		expected := &contentdomain.Post{
			ID:        postID,
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			Content:   "Post content",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo := &mockPostFindByIDRepo{
			getPostByIDFn: func(ctx context.Context, id uuid.UUID, namespace string) (*contentdomain.Post, error) {
				assert.Equal(t, postID, id)
				assert.Equal(t, "tadoku", namespace)
				return expected, nil
			},
		}

		svc := contentdomain.NewPostFindByID(repo)
		result, err := svc.Execute(adminContext(), postID, "tadoku")

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPostFindByIDRepo{}
		svc := contentdomain.NewPostFindByID(repo)

		_, err := svc.Execute(userContext(), postID, "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPostFindByIDRepo{}
		svc := contentdomain.NewPostFindByID(repo)

		_, err := svc.Execute(context.Background(), postID, "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrUnauthorized)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPostFindByIDRepo{
			getPostByIDFn: func(ctx context.Context, id uuid.UUID, namespace string) (*contentdomain.Post, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPostFindByID(repo)
		_, err := svc.Execute(adminContext(), postID, "tadoku")

		assert.ErrorIs(t, err, repoErr)
	})
}
