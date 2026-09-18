package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
)

func (a *Application) GetPostVersion(ctx context.Context, namespace string, postID, contentID uuid.UUID) (*posts.PostVersion, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.posts.GetPostVersion(ctx, namespace, postID, contentID)
}
