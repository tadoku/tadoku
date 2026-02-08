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

type mockPostVersionGetRepo struct {
	getPostVersionFn func(ctx context.Context, postID uuid.UUID, contentID uuid.UUID) (*contentdomain.PostVersion, error)
}

func (m *mockPostVersionGetRepo) GetPostVersion(ctx context.Context, postID uuid.UUID, contentID uuid.UUID) (*contentdomain.PostVersion, error) {
	return m.getPostVersionFn(ctx, postID, contentID)
}

func TestPostVersionGet_Execute(t *testing.T) {
	postID := uuid.New()
	contentID := uuid.New()

	t.Run("gets version successfully", func(t *testing.T) {
		expected := &contentdomain.PostVersion{
			ID:        contentID,
			Version:   3,
			Title:     "Version 3",
			Content:   "Post content",
			CreatedAt: time.Now(),
		}
		repo := &mockPostVersionGetRepo{
			getPostVersionFn: func(ctx context.Context, pID uuid.UUID, cID uuid.UUID) (*contentdomain.PostVersion, error) {
				assert.Equal(t, postID, pID)
				assert.Equal(t, contentID, cID)
				return expected, nil
			},
		}

		svc := contentdomain.NewPostVersionGet(repo)
		result, err := svc.Execute(adminContext(), postID, contentID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPostVersionGetRepo{}
		svc := contentdomain.NewPostVersionGet(repo)

		_, err := svc.Execute(userContext(), postID, contentID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPostVersionGetRepo{}
		svc := contentdomain.NewPostVersionGet(repo)

		_, err := svc.Execute(context.Background(), postID, contentID)

		assert.ErrorIs(t, err, contentdomain.ErrUnauthorized)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPostVersionGetRepo{
			getPostVersionFn: func(ctx context.Context, pID uuid.UUID, cID uuid.UUID) (*contentdomain.PostVersion, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPostVersionGet(repo)
		_, err := svc.Execute(adminContext(), postID, contentID)

		assert.ErrorIs(t, err, repoErr)
	})
}
