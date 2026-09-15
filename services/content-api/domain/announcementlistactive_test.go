package domain_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	contentdomain "github.com/tadoku/tadoku/services/content-api/domain"
)

type mockAnnouncementListActiveRepo struct {
	listActiveAnnouncementsFn func(ctx context.Context, namespace string, now time.Time) ([]contentdomain.Announcement, error)
}

func (m *mockAnnouncementListActiveRepo) ListActiveAnnouncements(ctx context.Context, namespace string, now time.Time) ([]contentdomain.Announcement, error) {
	if m.listActiveAnnouncementsFn != nil {
		return m.listActiveAnnouncementsFn(ctx, namespace, now)
	}
	return []contentdomain.Announcement{}, nil
}

func TestAnnouncementListActive_Execute(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	t.Run("lists active announcements successfully", func(t *testing.T) {
		repo := &mockAnnouncementListActiveRepo{
			listActiveAnnouncementsFn: func(ctx context.Context, namespace string, gotNow time.Time) ([]contentdomain.Announcement, error) {
				if namespace != "tadoku" || !gotNow.Equal(now) {
					t.Errorf("repository inputs: namespace=%q now=%s, want tadoku at %s", namespace, gotNow, now)
				}
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

		svc := contentdomain.NewAnnouncementListActive(repo, commondomain.NewMockClock(now))

		resp, err := svc.Execute(context.Background(), &contentdomain.AnnouncementListActiveRequest{
			Namespace: "tadoku",
		})

		if err != nil {
			t.Fatal(err)
		}
		if len(resp.Announcements) != 1 {
			t.Fatalf("announcements=%d, want 1", len(resp.Announcements))
		}
		if resp.Announcements[0].Title != "Active Notice" {
			t.Errorf("title=%q, want Active Notice", resp.Announcements[0].Title)
		}
	})

	t.Run("does not require admin role", func(t *testing.T) {
		repo := &mockAnnouncementListActiveRepo{}
		svc := contentdomain.NewAnnouncementListActive(repo, commondomain.NewMockClock(now))

		resp, err := svc.Execute(userContext(), &contentdomain.AnnouncementListActiveRequest{
			Namespace: "tadoku",
		})

		if err != nil {
			t.Fatal(err)
		}
		if len(resp.Announcements) != 0 {
			t.Errorf("announcements=%d, want 0", len(resp.Announcements))
		}
	})

	t.Run("returns error on missing namespace", func(t *testing.T) {
		repo := &mockAnnouncementListActiveRepo{}
		svc := contentdomain.NewAnnouncementListActive(repo, commondomain.NewMockClock(now))

		_, err := svc.Execute(context.Background(), &contentdomain.AnnouncementListActiveRequest{})

		if !errors.Is(err, contentdomain.ErrRequestInvalid) {
			t.Errorf("error=%v, want ErrRequestInvalid", err)
		}
	})
}
