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

type CreateAnnouncementParameters = announcements.CreateAnnouncementParameters

func (a *Application) CreateAnnouncement(ctx context.Context, parameters CreateAnnouncementParameters) (*announcements.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return mutateThenReadBack(
		ctx,
		a.db,
		func(ctx context.Context) error {
			return a.announcements.CreateAnnouncement(ctx, parameters)
		},
		func(ctx context.Context) (*announcements.Announcement, error) {
			return a.announcements.FindAnnouncementByID(ctx, parameters.Namespace, parameters.ID)
		},
	)
}

type UpdateAnnouncementParameters = announcements.UpdateAnnouncementParameters

func (a *Application) UpdateAnnouncement(ctx context.Context, parameters UpdateAnnouncementParameters) (*announcements.Announcement, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return mutateThenReadBack(
		ctx,
		a.db,
		func(ctx context.Context) error {
			return a.announcements.UpdateAnnouncement(ctx, parameters)
		},
		func(ctx context.Context) (*announcements.Announcement, error) {
			return a.announcements.FindAnnouncementByID(ctx, parameters.Namespace, parameters.ID)
		},
	)
}

func (a *Application) DeleteAnnouncement(ctx context.Context, namespace string, id uuid.UUID) error {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return err
	}

	return a.announcements.DeleteAnnouncement(ctx, namespace, id)
}
