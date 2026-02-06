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

type mockPageVersionListRepo struct {
	listPageVersionsFn func(ctx context.Context, pageID uuid.UUID) ([]contentdomain.PageVersion, error)
}

func (m *mockPageVersionListRepo) ListPageVersions(ctx context.Context, pageID uuid.UUID) ([]contentdomain.PageVersion, error) {
	return m.listPageVersionsFn(ctx, pageID)
}

func TestPageVersionList_Execute(t *testing.T) {
	pageID := uuid.New()

	t.Run("lists versions successfully", func(t *testing.T) {
		expected := []contentdomain.PageVersion{
			{ID: uuid.New(), Version: 2, Title: "Updated", CreatedAt: time.Now()},
			{ID: uuid.New(), Version: 1, Title: "Original", CreatedAt: time.Now()},
		}
		repo := &mockPageVersionListRepo{
			listPageVersionsFn: func(ctx context.Context, id uuid.UUID) ([]contentdomain.PageVersion, error) {
				assert.Equal(t, pageID, id)
				return expected, nil
			},
		}

		svc := contentdomain.NewPageVersionList(repo)
		result, err := svc.Execute(adminContext(), pageID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageVersionListRepo{}
		svc := contentdomain.NewPageVersionList(repo)

		_, err := svc.Execute(userContext(), pageID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPageVersionListRepo{}
		svc := contentdomain.NewPageVersionList(repo)

		_, err := svc.Execute(context.Background(), pageID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPageVersionListRepo{
			listPageVersionsFn: func(ctx context.Context, id uuid.UUID) ([]contentdomain.PageVersion, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPageVersionList(repo)
		_, err := svc.Execute(adminContext(), pageID)

		assert.ErrorIs(t, err, repoErr)
	})
}
