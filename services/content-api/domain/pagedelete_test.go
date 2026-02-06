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
	deletePageFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPageDeleteRepo) DeletePage(ctx context.Context, id uuid.UUID) error {
	return m.deletePageFn(ctx, id)
}

func TestPageDelete_Execute(t *testing.T) {
	pageID := uuid.New()

	t.Run("deletes page successfully", func(t *testing.T) {
		repo := &mockPageDeleteRepo{
			deletePageFn: func(ctx context.Context, id uuid.UUID) error {
				assert.Equal(t, pageID, id)
				return nil
			},
		}

		svc := contentdomain.NewPageDelete(repo)
		err := svc.Execute(adminContext(), pageID)

		require.NoError(t, err)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageDeleteRepo{}
		svc := contentdomain.NewPageDelete(repo)

		err := svc.Execute(userContext(), pageID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPageDeleteRepo{}
		svc := contentdomain.NewPageDelete(repo)

		err := svc.Execute(context.Background(), pageID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPageDeleteRepo{
			deletePageFn: func(ctx context.Context, id uuid.UUID) error {
				return repoErr
			},
		}

		svc := contentdomain.NewPageDelete(repo)
		err := svc.Execute(adminContext(), pageID)

		assert.ErrorIs(t, err, repoErr)
	})
}
