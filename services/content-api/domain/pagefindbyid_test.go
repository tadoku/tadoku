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

type mockPageFindByIDRepo struct {
	getPageByIDFn func(ctx context.Context, id uuid.UUID, namespace string) (*contentdomain.Page, error)
}

func (m *mockPageFindByIDRepo) GetPageByID(ctx context.Context, id uuid.UUID, namespace string) (*contentdomain.Page, error) {
	return m.getPageByIDFn(ctx, id, namespace)
}

func TestPageFindByID_Execute(t *testing.T) {
	pageID := uuid.New()

	t.Run("finds page successfully", func(t *testing.T) {
		expected := &contentdomain.Page{
			ID:        pageID,
			Namespace: "blog",
			Slug:      "hello-world",
			Title:     "Hello World",
			HTML:      "<p>Content</p>",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		repo := &mockPageFindByIDRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID, namespace string) (*contentdomain.Page, error) {
				assert.Equal(t, pageID, id)
				assert.Equal(t, "tadoku", namespace)
				return expected, nil
			},
		}

		svc := contentdomain.NewPageFindByID(repo)
		result, err := svc.Execute(adminContext(), pageID, "tadoku")

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageFindByIDRepo{}
		svc := contentdomain.NewPageFindByID(repo)

		_, err := svc.Execute(userContext(), pageID, "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns unauthorized when no session", func(t *testing.T) {
		repo := &mockPageFindByIDRepo{}
		svc := contentdomain.NewPageFindByID(repo)

		_, err := svc.Execute(context.Background(), pageID, "tadoku")

		assert.ErrorIs(t, err, contentdomain.ErrUnauthorized)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPageFindByIDRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID, namespace string) (*contentdomain.Page, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPageFindByID(repo)
		_, err := svc.Execute(adminContext(), pageID, "tadoku")

		assert.ErrorIs(t, err, repoErr)
	})
}
