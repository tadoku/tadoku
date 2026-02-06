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

type mockPageDeleteRepo struct {
	deletePageFn func(ctx context.Context, id uuid.UUID, namespace string) error
}

func (m *mockPageDeleteRepo) DeletePage(ctx context.Context, id uuid.UUID, namespace string) error {
	return m.deletePageFn(ctx, id, namespace)
}

func TestPageDelete_Execute(t *testing.T) {
	pageID := uuid.New()

	t.Run("deletes page successfully", func(t *testing.T) {
		repo := &mockPageDeleteRepo{
			deletePageFn: func(ctx context.Context, id uuid.UUID, namespace string) error {
				assert.Equal(t, pageID, id)
				assert.Equal(t, "tadoku", namespace)
				return nil
			},
		}

		svc := contentdomain.NewPageDelete(repo)
		err := svc.Execute(adminContext(), pageID, "tadoku")

		require.NoError(t, err)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageDeleteRepo{}
		svc := contentdomain.NewPageDelete(repo)

		err := svc.Execute(userContext(), pageID, "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPageDeleteRepo{}
		svc := contentdomain.NewPageDelete(repo)

		err := svc.Execute(context.Background(), pageID, "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPageDeleteRepo{
			deletePageFn: func(ctx context.Context, id uuid.UUID, namespace string) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPageDelete(repo)
		err := svc.Execute(adminContext(), pageID, "tadoku")

		assert.ErrorIs(t, err, repoErr)
	})
}
