package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type CreatePostParameters = posts.CreatePostParameters

func (a *Application) CreatePost(ctx context.Context, parameters CreatePostParameters) (*posts.Post, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	var result *posts.Post
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		if err := a.posts.CreatePost(ctx, parameters); err != nil {
			return err
		}
		result, err = a.posts.FindPostByID(ctx, parameters.Namespace, parameters.ID)
		return err
	})
	if err != nil {
		// The readback precedes commit; discard it if the transaction fails.
		return nil, err
	}
	return result, nil
}
