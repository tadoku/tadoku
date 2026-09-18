package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
)

func (a *Application) FindAnnouncementByID(ctx context.Context, namespace string, id uuid.UUID) (*announcements.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.announcements.FindAnnouncementByID(ctx, namespace, id)
}

func (a *Application) ListActiveAnnouncements(ctx context.Context, namespace string) ([]announcements.Announcement, error) {
	return a.announcements.ListActiveAnnouncements(ctx, namespace)
}

func (a *Application) ListAnnouncements(ctx context.Context, namespace string, pageSize, page int) (*announcements.AnnouncementList, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.announcements.ListAnnouncements(ctx, namespace, pageSize, page)
}
