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

type mockPostVersionListRepo struct {
	listPostVersionsFn func(ctx context.Context, postID uuid.UUID) ([]contentdomain.PostVersion, error)
}

func (m *mockPostVersionListRepo) ListPostVersions(ctx context.Context, postID uuid.UUID) ([]contentdomain.PostVersion, error) {
	return m.listPostVersionsFn(ctx, postID)
}

func TestPostVersionList_Execute(t *testing.T) {
	postID := uuid.New()

	t.Run("lists versions successfully", func(t *testing.T) {
		expected := []contentdomain.PostVersion{
			{ID: uuid.New(), Version: 2, Title: "Updated", CreatedAt: time.Now()},
			{ID: uuid.New(), Version: 1, Title: "Original", CreatedAt: time.Now()},
		}
		repo := &mockPostVersionListRepo{
			listPostVersionsFn: func(ctx context.Context, id uuid.UUID) ([]contentdomain.PostVersion, error) {
				assert.Equal(t, postID, id)
				return expected, nil
			},
		}

		svc := contentdomain.NewPostVersionList(repo)
		result, err := svc.Execute(adminContext(), postID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPostVersionListRepo{}
		svc := contentdomain.NewPostVersionList(repo)

		_, err := svc.Execute(userContext(), postID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns unauthorized when no session", func(t *testing.T) {
		repo := &mockPostVersionListRepo{}
		svc := contentdomain.NewPostVersionList(repo)

		_, err := svc.Execute(context.Background(), postID)

		assert.ErrorIs(t, err, contentdomain.ErrUnauthorized)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPostVersionListRepo{
			listPostVersionsFn: func(ctx context.Context, id uuid.UUID) ([]contentdomain.PostVersion, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPostVersionList(repo)
		_, err := svc.Execute(adminContext(), postID)

		assert.ErrorIs(t, err, repoErr)
	})
}
