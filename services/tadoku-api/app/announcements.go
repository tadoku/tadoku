package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

func (a *Application) ListActiveAnnouncements(ctx context.Context, namespace string) ([]content.Announcement, error) {
	return a.content.ListActiveAnnouncements(ctx, namespace)
}

func (a *Application) ListAnnouncements(ctx context.Context, namespace string, pageSize, page int) (*content.AnnouncementList, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	var result *content.AnnouncementList
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		var err error
		result, err = a.content.ListAnnouncements(ctx, namespace, pageSize, page)
		return err
	})
	return result, err
}
