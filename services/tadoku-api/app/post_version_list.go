package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
)

func (a *Application) ListPostVersions(ctx context.Context, namespace string, id uuid.UUID) ([]content.PostVersion, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.content.ListPostVersions(ctx, namespace, id)
}
