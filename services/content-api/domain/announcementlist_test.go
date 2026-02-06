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

type mockAnnouncementListRepo struct {
	listAnnouncementsFn func(ctx context.Context, namespace string, pageSize, page int) (*contentdomain.AnnouncementListResult, error)
}

func (m *mockAnnouncementListRepo) ListAnnouncements(ctx context.Context, namespace string, pageSize, page int) (*contentdomain.AnnouncementListResult, error) {
	if m.listAnnouncementsFn != nil {
		return m.listAnnouncementsFn(ctx, namespace, pageSize, page)
	}
	return &contentdomain.AnnouncementListResult{}, nil
}

func TestAnnouncementList_Execute(t *testing.T) {
	startsAt := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	endsAt := time.Date(2024, 2, 7, 0, 0, 0, 0, time.UTC)

	t.Run("lists announcements successfully", func(t *testing.T) {
		repo := &mockAnnouncementListRepo{
			listAnnouncementsFn: func(ctx context.Context, namespace string, pageSize, page int) (*contentdomain.AnnouncementListResult, error) {
				return &contentdomain.AnnouncementListResult{
					Announcements: []contentdomain.Announcement{
						{
							ID:        uuid.New(),
							Namespace: "tadoku",
							Title:     "Notice 1",
							Content:   "Content 1",
							Style:     "info",
							StartsAt:  startsAt,
							EndsAt:    endsAt,
						},
					},
					TotalSize:     1,
					NextPageToken: "",
				}, nil
			},
		}

		svc := contentdomain.NewAnnouncementList(repo)

		resp, err := svc.Execute(adminContext(), &contentdomain.AnnouncementListRequest{
			Namespace: "tadoku",
			PageSize:  10,
			Page:      0,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Announcements, 1)
		assert.Equal(t, 1, resp.TotalSize)
	})

	t.Run("returns forbidden when not admin", func(t *testing.T) {
		repo := &mockAnnouncementListRepo{}
		svc := contentdomain.NewAnnouncementList(repo)

		_, err := svc.Execute(userContext(), &contentdomain.AnnouncementListRequest{
			Namespace: "tadoku",
		})

		assert.ErrorIs(t, err, contentdomain.ErrForbidden)
	})

	t.Run("defaults page size", func(t *testing.T) {
		var capturedPageSize int
		repo := &mockAnnouncementListRepo{
			listAnnouncementsFn: func(ctx context.Context, namespace string, pageSize, page int) (*contentdomain.AnnouncementListResult, error) {
				capturedPageSize = pageSize
				return &contentdomain.AnnouncementListResult{}, nil
			},
		}

		svc := contentdomain.NewAnnouncementList(repo)

		_, err := svc.Execute(adminContext(), &contentdomain.AnnouncementListRequest{
			Namespace: "tadoku",
		})

		require.NoError(t, err)
		assert.Equal(t, 10, capturedPageSize)
	})
}
