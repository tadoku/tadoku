package app

import (
	"context"

	"github.com/google/uuid"
)

func (a *Application) DeletePost(ctx context.Context, namespace string, id uuid.UUID) error {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return err
	}

	return a.posts.DeletePost(ctx, namespace, id)
}
