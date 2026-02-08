package domain_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockPostDeleteRepo struct {
	deletePostFn func(ctx context.Context, id uuid.UUID, namespace string) error
}

func (m *mockPostDeleteRepo) DeletePost(ctx context.Context, id uuid.UUID, namespace string) error {
	return m.deletePostFn(ctx, id, namespace)
}

func TestPostDelete_Execute(t *testing.T) {
	postID := uuid.New()

	t.Run("deletes post successfully", func(t *testing.T) {
		repo := &mockPostDeleteRepo{
			deletePostFn: func(ctx context.Context, id uuid.UUID, namespace string) error {
				assert.Equal(t, postID, id)
				assert.Equal(t, "tadoku", namespace)
				return nil
			},
		}

		svc := contentdomain.NewPostDelete(repo)
		err := svc.Execute(adminContext(), postID, "tadoku")

		require.NoError(t, err)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPostDeleteRepo{}
		svc := contentdomain.NewPostDelete(repo)

		err := svc.Execute(userContext(), postID, "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPostDeleteRepo{}
		svc := contentdomain.NewPostDelete(repo)

		err := svc.Execute(context.Background(), postID, "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrUnauthorized)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPostDeleteRepo{
			deletePostFn: func(ctx context.Context, id uuid.UUID, namespace string) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPostDelete(repo)
		err := svc.Execute(adminContext(), postID, "tadoku")

		assert.ErrorIs(t, err, repoErr)
	})
}
