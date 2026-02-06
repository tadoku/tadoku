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

type mockAnnouncementCreateRepo struct {
	createAnnouncementFn func(ctx context.Context, a *contentdomain.Announcement) error
}

func (m *mockAnnouncementCreateRepo) CreateAnnouncement(ctx context.Context, a *contentdomain.Announcement) error {
	if m.createAnnouncementFn != nil {
		return m.createAnnouncementFn(ctx, a)
	}
	return nil
}

func TestAnnouncementCreate_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := &mockClock{now: now}
	startsAt := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 2, 7, 0, 0, 0, 0, time.UTC)

	t.Run("creates announcement successfully", func(t *testing.T) {
		var saved *contentdomain.Announcement
		repo := &mockAnnouncementCreateRepo{
			createAnnouncementFn: func(ctx context.Context, a *contentdomain.Announcement) error {
				saved = a
				return nil
			},
		}

		svc := contentdomain.NewAnnouncementCreate(repo, clock)
		id := uuid.New()
		href := "https://tadoku.app/blog"

		resp, err := svc.Execute(adminContext(), &contentdomain.AnnouncementCreateRequest{
			ID:        id,
			Namespace: "tadoku",
			Title:     "Maintenance Notice",
			Content:   "The site will be down **Friday**.",
			Style:     "warning",
			Href:      &href,
			StartsAt:  startsAt,
			EndsAt:    endsAt,
		})

		require.NoError(t, err)
		assert.Equal(t, &contentdomain.Announcement{
			ID:        id,
			Namespace: "tadoku",
			Title:     "Maintenance Notice",
			Content:   "The site will be down **Friday**.",
			Style:     "warning",
			Href:      &href,
			StartsAt:  startsAt,
			EndsAt:    endsAt,
			CreatedAt: now,
			UpdatedAt: now,
		}, resp.Announcement)
		assert.Equal(t, resp.Announcement, saved)
	})

	t.Run("creates announcement without href", func(t *testing.T) {
		repo := &mockAnnouncementCreateRepo{}
		svc := contentdomain.NewAnnouncementCreate(repo, clock)

		resp, err := svc.Execute(adminContext(), &contentdomain.AnnouncementCreateRequest{
			ID:        uuid.New(),
			Namespace: "tadoku",
			Title:     "Simple Notice",
			Content:   "Just a notice",
			Style:     "info",
			StartsAt:  startsAt,
			EndsAt:    endsAt,
		})

		require.NoError(t, err)
		assert.Nil(t, resp.Announcement.Href)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockAnnouncementCreateRepo{}
		svc := contentdomain.NewAnnouncementCreate(repo, clock)

		_, err := svc.Execute(userContext(), &contentdomain.AnnouncementCreateRequest{
			ID:        uuid.New(),
			Namespace: "tadoku",
			Title:     "Notice",
			Content:   "Content",
			Style:     "info",
			StartsAt:  startsAt,
			EndsAt:    endsAt,
		})

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("returns error on invalid style", func(t *testing.T) {
		repo := &mockAnnouncementCreateRepo{}
		svc := contentdomain.NewAnnouncementCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.AnnouncementCreateRequest{
			ID:        uuid.New(),
			Namespace: "tadoku",
			Title:     "Notice",
			Content:   "Content",
			Style:     "invalid-style",
			StartsAt:  startsAt,
			EndsAt:    endsAt,
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidAnnouncement)
	})

	t.Run("returns error when ends_at is before starts_at", func(t *testing.T) {
		repo := &mockAnnouncementCreateRepo{}
		svc := contentdomain.NewAnnouncementCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.AnnouncementCreateRequest{
			ID:        uuid.New(),
			Namespace: "tadoku",
			Title:     "Notice",
			Content:   "Content",
			Style:     "info",
			StartsAt:  endsAt,
			EndsAt:    startsAt,
		})

		assert.ErrorIs(t, err, contentdomain.ErrInvalidAnnouncement)
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database connection failed")
		repo := &mockAnnouncementCreateRepo{
			createAnnouncementFn: func(ctx context.Context, a *contentdomain.Announcement) error {
				return repoErr
			},
		}

		svc := contentdomain.NewAnnouncementCreate(repo, clock)

		_, err := svc.Execute(adminContext(), &contentdomain.AnnouncementCreateRequest{
			ID:        uuid.New(),
			Namespace: "tadoku",
			Title:     "Notice",
			Content:   "Content",
			Style:     "info",
			StartsAt:  startsAt,
			EndsAt:    endsAt,
		})

		assert.ErrorIs(t, err, repoErr)
	})
}
