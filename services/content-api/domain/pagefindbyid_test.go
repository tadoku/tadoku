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
	getPageByIDFn func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error)
}

func (m *mockPageFindByIDRepo) GetPageByID(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
	return m.getPageByIDFn(ctx, id)
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
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				assert.Equal(t, pageID, id)
				return expected, nil
			},
		}

		svc := contentdomain.NewPageFindByID(repo)
		result, err := svc.Execute(adminContext(), pageID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockPageFindByIDRepo{}
		svc := contentdomain.NewPageFindByID(repo)

		_, err := svc.Execute(userContext(), pageID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns forbidden when no session", func(t *testing.T) {
		repo := &mockPageFindByIDRepo{}
		svc := contentdomain.NewPageFindByID(repo)

		_, err := svc.Execute(context.Background(), pageID)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database error")
		repo := &mockPageFindByIDRepo{
			getPageByIDFn: func(ctx context.Context, id uuid.UUID) (*contentdomain.Page, error) {
				return nil, repoErr
			},
		}

		svc := contentdomain.NewPageFindByID(repo)
		_, err := svc.Execute(adminContext(), pageID)

		assert.ErrorIs(t, err, repoErr)
	})
}
