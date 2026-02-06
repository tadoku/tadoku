package domain_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockAnnouncementFindByIDRepo struct {
	getAnnouncementByIDFn func(ctx context.Context, id uuid.UUID) (*contentdomain.Announcement, error)
}

func (m *mockAnnouncementFindByIDRepo) GetAnnouncementByID(ctx context.Context, id uuid.UUID) (*contentdomain.Announcement, error) {
	if m.getAnnouncementByIDFn != nil {
		return m.getAnnouncementByIDFn(ctx, id)
	}
	return &contentdomain.Announcement{}, nil
}

func TestAnnouncementFindByID_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	id := uuid.New()

	t.Run("finds announcement by ID", func(t *testing.T) {
		expected := &contentdomain.Announcement{
			ID:        id,
			Namespace: "tadoku",
			Title:     "Notice",
			Content:   "Content",
			Style:     "info",
			StartsAt:  now,
			EndsAt:    now.Add(24 * time.Hour),
		}

		repo := &mockAnnouncementFindByIDRepo{
			getAnnouncementByIDFn: func(ctx context.Context, reqID uuid.UUID) (*contentdomain.Announcement, error) {
				assert.Equal(t, id, reqID)
				return expected, nil
			},
		}

		svc := contentdomain.NewAnnouncementFindByID(repo)
		result, err := svc.Execute(adminContext(), id)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockAnnouncementFindByIDRepo{}
		svc := contentdomain.NewAnnouncementFindByID(repo)

		_, err := svc.Execute(userContext(), id)

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns not found", func(t *testing.T) {
		repo := &mockAnnouncementFindByIDRepo{
			getAnnouncementByIDFn: func(ctx context.Context, reqID uuid.UUID) (*contentdomain.Announcement, error) {
				return nil, contentdomain.ErrAnnouncementNotFound
			},
		}

		svc := contentdomain.NewAnnouncementFindByID(repo)
		_, err := svc.Execute(adminContext(), id)

		assert.ErrorIs(t, err, contentdomain.ErrAnnouncementNotFound)
	})
}
