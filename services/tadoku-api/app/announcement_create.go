package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type CreateAnnouncementParameters = content.CreateAnnouncementParameters

func (a *Application) CreateAnnouncement(ctx context.Context, parameters CreateAnnouncementParameters) (*content.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	var result *content.Announcement
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		if err := a.content.CreateAnnouncement(ctx, parameters); err != nil {
			return err
		}
		result, err = a.content.FindAnnouncementByID(ctx, parameters.Namespace, parameters.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
