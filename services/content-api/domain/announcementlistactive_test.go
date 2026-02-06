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

type mockAnnouncementListActiveRepo struct {
	listActiveAnnouncementsFn func(ctx context.Context, namespace string) ([]contentdomain.Announcement, error)
}

func (m *mockAnnouncementListActiveRepo) ListActiveAnnouncements(ctx context.Context, namespace string) ([]contentdomain.Announcement, error) {
	if m.listActiveAnnouncementsFn != nil {
		return m.listActiveAnnouncementsFn(ctx, namespace)
	}
	return []contentdomain.Announcement{}, nil
}

func TestAnnouncementListActive_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	clock := &mockClock{now: now}

	t.Run("lists active announcements successfully", func(t *testing.T) {
		repo := &mockAnnouncementListActiveRepo{
			listActiveAnnouncementsFn: func(ctx context.Context, namespace string) ([]contentdomain.Announcement, error) {
				return []contentdomain.Announcement{
					{
						ID:        uuid.New(),
						Namespace: "tadoku",
						Title:     "Active Notice",
						Content:   "This is active",
						Style:     "info",
						StartsAt:  now.Add(-24 * time.Hour),
						EndsAt:    now.Add(24 * time.Hour),
					},
				}, nil
			},
		}

		svc := contentdomain.NewAnnouncementListActive(repo, clock)

		resp, err := svc.Execute(context.Background(), &contentdomain.AnnouncementListActiveRequest{
			Namespace: "tadoku",
		})

		require.NoError(t, err)
		assert.Len(t, resp.Announcements, 1)
		assert.Equal(t, "Active Notice", resp.Announcements[0].Title)
	})

	t.Run("does not require admin role", func(t *testing.T) {
		repo := &mockAnnouncementListActiveRepo{}
		svc := contentdomain.NewAnnouncementListActive(repo, clock)

		resp, err := svc.Execute(userContext(), &contentdomain.AnnouncementListActiveRequest{
			Namespace: "tadoku",
		})

		require.NoError(t, err)
		assert.Empty(t, resp.Announcements)
	})

	t.Run("returns error on missing namespace", func(t *testing.T) {
		repo := &mockAnnouncementListActiveRepo{}
		svc := contentdomain.NewAnnouncementListActive(repo, clock)

		_, err := svc.Execute(context.Background(), &contentdomain.AnnouncementListActiveRequest{})

		assert.ErrorIs(t, err, contentdomain.ErrRequestInvalid)
	})
}
