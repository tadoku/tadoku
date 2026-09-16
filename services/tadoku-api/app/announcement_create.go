package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
)

var ErrInvalidAnnouncement = content.ErrInvalidAnnouncement

type CreateAnnouncementRequest = content.CreateAnnouncementRequest

func (a *Application) CreateAnnouncement(ctx context.Context, request content.CreateAnnouncementRequest) (*content.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.content.CreateAnnouncement(ctx, request)
}
