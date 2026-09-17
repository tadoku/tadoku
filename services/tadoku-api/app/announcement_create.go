package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
)

type CreateAnnouncementParameters = content.CreateAnnouncementParameters

func (a *Application) CreateAnnouncement(ctx context.Context, parameters CreateAnnouncementParameters) (*content.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.content.CreateAnnouncement(ctx, parameters)
}
