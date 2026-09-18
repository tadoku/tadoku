package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type CreatePostParameters = content.CreatePostParameters

func (a *Application) CreatePost(ctx context.Context, parameters CreatePostParameters) (*content.Post, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	var result *content.Post
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		if err := a.content.CreatePost(ctx, parameters); err != nil {
			return err
		}
		result, err = a.content.FindPostByID(ctx, parameters.Namespace, parameters.ID)
		return err
	})
	if err != nil {
		// The readback precedes commit; discard it if the transaction fails.
		return nil, err
	}
	return result, nil
}
