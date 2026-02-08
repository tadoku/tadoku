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

type mockPageVersionGetRepo struct {
	getPageVersionFn func(ctx context.Context, pageID uuid.UUID, contentID uuid.UUID) (*contentdomain.PageVersion, error)
}

func (m *mockPageVersionGetRepo) GetPageVersion(ctx context.Context, pageID uuid.UUID, contentID uuid.UUID) (*contentdomain.PageVersion, error) {
	return m.getPageVersionFn(ctx, pageID, contentID)
}

func TestPageVersionGet_Execute(t *testing.T) {
	pageID := uuid.New()
	contentID := uuid.New()

	t.Run("gets version successfully", func(t *testing.T) {
		expected := &contentdomain.PageVersion{
			ID:        contentID,
			Version:   3,
			Title:     "Version 3",
			HTML:      "<p>Content</p>",
			CreatedAt: time.Now(),
		}
		repo := &mockPageVersionGetRepo{
			getPageVersionFn: func(ctx context.Context, pID uuid.UUID, cID uuid.UUID) (*contentdomain.PageVersion, error) {
				assert.Equal(t, pageID, pID)
				assert.Equal(t, contentID, cID)
				return expected, nil
			},
		}

		svc := contentdomain.NewPageVersionGet(repo)
		result, err := svc.Execute(adminContext(), pageID, contentID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageVersionGetRepo{}
		svc := contentdomain.NewPageVersionGet(repo)

		_, err := svc.Execute(userContext(), pageID, contentID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns unauthorized when no session", func(t *testing.T) {
		repo := &mockPageVersionGetRepo{}
		svc := contentdomain.NewPageVersionGet(repo)

		_, err := svc.Execute(context.Background(), pageID, contentID)

		assert.ErrorIs(t, err, contentdomain.ErrUnauthorized)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPageVersionGetRepo{
			getPageVersionFn: func(ctx context.Context, pID uuid.UUID, cID uuid.UUID) (*contentdomain.PageVersion, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPageVersionGet(repo)
		_, err := svc.Execute(adminContext(), pageID, contentID)

		assert.ErrorIs(t, err, repoErr)
	})
}
