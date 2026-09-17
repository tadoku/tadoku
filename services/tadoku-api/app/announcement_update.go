package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type UpdateAnnouncementParameters = content.UpdateAnnouncementParameters

func (a *Application) UpdateAnnouncement(ctx context.Context, namespace string, id uuid.UUID, parameters UpdateAnnouncementParameters) (*content.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := parameters.Validate(namespace); err != nil {
		return nil, err
	}

	var result *content.Announcement
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		result, err = a.content.UpdateAnnouncement(ctx, namespace, id, parameters)
		return
	})
	return result, err
}
