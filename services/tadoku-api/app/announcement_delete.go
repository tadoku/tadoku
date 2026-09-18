package app

import (
	"context"

	"github.com/google/uuid"
)

func (a *Application) DeleteAnnouncement(ctx context.Context, namespace string, id uuid.UUID) error {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return err
	}

	return a.announcements.DeleteAnnouncement(ctx, namespace, id)
}
