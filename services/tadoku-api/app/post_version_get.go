package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
)

func (a *Application) GetPostVersion(ctx context.Context, namespace string, postID, contentID uuid.UUID) (*content.PostVersion, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.content.GetPostVersion(ctx, namespace, postID, contentID)
}
