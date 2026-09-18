package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (a *Application) FindPostBySlug(ctx context.Context, namespace, slug string) (*posts.Post, error) {
	post, err := a.posts.FindPostBySlug(ctx, namespace, slug)
	if err == nil {
		post.CreatedAt = nil
		post.UpdatedAt = nil
		return post, nil
	}
	if !errors.Is(err, posts.ErrPostNotFound) && errx.KindOf(err) != errx.InvalidInput {
		return nil, err
	}

	id, err := uuid.Parse(slug)
	if err != nil {
		return nil, posts.ErrPostNotFound
	}
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		if errx.KindOf(err) == errx.Forbidden {
			return nil, err
		}
		return nil, posts.ErrPostNotFound
	}

	post, err = a.posts.FindPostByID(ctx, namespace, id)
	if err != nil {
		return nil, posts.ErrPostNotFound
	}
	return post, nil
}
