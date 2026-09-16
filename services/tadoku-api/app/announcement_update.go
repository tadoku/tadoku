package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type AnnouncementUpdateRequest = content.AnnouncementUpdateRequest

func (a *Application) UpdateAnnouncement(ctx context.Context, namespace string, id uuid.UUID, req AnnouncementUpdateRequest) (*content.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := req.Validate(namespace); err != nil {
		return nil, err
	}

	var result *content.Announcement
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		result, err = a.content.UpdateAnnouncement(ctx, namespace, id, req)
		return
	})
	return result, err
}
