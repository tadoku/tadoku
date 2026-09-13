package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
)

func (a *Application) ListActiveAnnouncements(ctx context.Context, namespace string) ([]content.Announcement, error) {
	return a.content.ListActiveAnnouncements(ctx, namespace)
}
