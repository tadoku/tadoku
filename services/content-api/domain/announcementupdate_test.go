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

type mockAnnouncementUpdateRepo struct {
	getAnnouncementByIDFn func(ctx context.Context, id uuid.UUID) (*contentdomain.Announcement, error)
	updateAnnouncementFn  func(ctx context.Context, a *contentdomain.Announcement) error
}

func (m *mockAnnouncementUpdateRepo) GetAnnouncementByID(ctx context.Context, id uuid.UUID) (*contentdomain.Announcement, error) {
	if m.getAnnouncementByIDFn != nil {
		return m.getAnnouncementByIDFn(ctx, id)
	}
	return &contentdomain.Announcement{}, nil
}

func (m *mockAnnouncementUpdateRepo) UpdateAnnouncement(ctx context.Context, a *contentdomain.Announcement) error {
	if m.updateAnnouncementFn != nil {
		return m.updateAnnouncementFn(ctx, a)
	}
	return nil
}

func TestAnnouncementUpdate_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := &mockClock{now: now}
	startsAt := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 2, 7, 0, 0, 0, 0, time.UTC)
	id := uuid.New()

	t.Run("updates announcement successfully", func(t *testing.T) {
		existing := &contentdomain.Announcement{
			ID:        id,
			Namespace: "tadoku",
			Title:     "Old Title",
			Content:   "Old content",
			Style:     "info",
			StartsAt:  startsAt,
			EndsAt:    endsAt,
			CreatedAt: now.Add(-24 * time.Hour),
			UpdatedAt: now.Add(-24 * time.Hour),
		}

		var updated *contentdomain.Announcement
		repo := &mockAnnouncementUpdateRepo{
			getAnnouncementByIDFn: func(ctx context.Context, reqID uuid.UUID) (*contentdomain.Announcement, error) {
				return existing, nil
			},
			updateAnnouncementFn: func(ctx context.Context, a *contentdomain.Announcement) error {
				updated = a
				return nil
			},
		}

		svc := contentdomain.NewAnnouncementUpdate(repo, clock)
		href := "https://example.com"

		resp, err := svc.Execute(adminContext(), id, &contentdomain.AnnouncementUpdateRequest{
			Namespace: "tadoku",
			Title:     "New Title",
			Content:   "New content",
			Style:     "warning",
			Href:      &href,
			StartsAt:  startsAt,
			EndsAt:    endsAt,
		})

		require.NoError(t, err)
		assert.Equal(t, "New Title", resp.Announcement.Title)
		assert.Equal(t, "New content", resp.Announcement.Content)
		assert.Equal(t, "warning", resp.Announcement.Style)
		assert.Equal(t, &href, resp.Announcement.Href)
		assert.Equal(t, now, resp.Announcement.UpdatedAt)
		assert.Equal(t, updated, resp.Announcement)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockAnnouncementUpdateRepo{}
		svc := contentdomain.NewAnnouncementUpdate(repo, clock)

		_, err := svc.Execute(userContext(), id, &contentdomain.AnnouncementUpdateRequest{
			Namespace: "tadoku",
			Title:     "Title",
			Content:   "Content",
			Style:     "info",
			StartsAt:  startsAt,
			EndsAt:    endsAt,
		})

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns not found when announcement does not exist", func(t *testing.T) {
		repo := &mockAnnouncementUpdateRepo{
			getAnnouncementByIDFn: func(ctx context.Context, reqID uuid.UUID) (*contentdomain.Announcement, error) {
				return nil, contentdomain.ErrAnnouncementNotFound
			},
		}

		svc := contentdomain.NewAnnouncementUpdate(repo, clock)

		_, err := svc.Execute(adminContext(), id, &contentdomain.AnnouncementUpdateRequest{
			Namespace: "tadoku",
			Title:     "Title",
			Content:   "Content",
			Style:     "info",
			StartsAt:  startsAt,
			EndsAt:    endsAt,
		})

		assert.ErrorIs(t, err, contentdomain.ErrAnnouncementNotFound)
	})
}
