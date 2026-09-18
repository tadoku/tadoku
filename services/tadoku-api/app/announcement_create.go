package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type CreateAnnouncementParameters = announcements.CreateAnnouncementParameters

func (a *Application) CreateAnnouncement(ctx context.Context, parameters CreateAnnouncementParameters) (*announcements.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	var result *announcements.Announcement
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		if err := a.announcements.CreateAnnouncement(ctx, parameters); err != nil {
			return err
		}
		result, err = a.announcements.FindAnnouncementByID(ctx, parameters.Namespace, parameters.ID)
		return err
	})
	if err != nil {
		// The readback precedes commit; discard it if the transaction fails.
		return nil, err
	}
	return result, nil
}
