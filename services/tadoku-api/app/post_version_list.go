package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
)

func (a *Application) ListPostVersions(ctx context.Context, namespace string, id uuid.UUID) ([]posts.PostVersion, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.posts.ListPostVersions(ctx, namespace, id)
}
