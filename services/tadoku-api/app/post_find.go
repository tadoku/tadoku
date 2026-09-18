package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (a *Application) FindPostBySlug(ctx context.Context, namespace, slug string) (*content.Post, error) {
	post, err := a.content.FindPostBySlug(ctx, namespace, slug)
	if err == nil {
		return post, nil
	}
	if !errors.Is(err, content.ErrPostNotFound) && errx.KindOf(err) != errx.InvalidInput {
		return nil, err
	}

	id, err := uuid.Parse(slug)
	if err != nil {
		return nil, content.ErrPostNotFound
	}
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		if errx.KindOf(err) == errx.Forbidden {
			return nil, err
		}
		return nil, content.ErrPostNotFound
	}

	post, err = a.content.FindPostByID(ctx, namespace, id)
	if err != nil {
		return nil, content.ErrPostNotFound
	}
	return post, nil
}
