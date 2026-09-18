package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (a *Application) FindPostBySlug(ctx context.Context, namespace, slug string) (*content.Post, bool, error) {
	post, err := a.content.FindPostBySlug(ctx, namespace, slug)
	if err == nil {
		return post, false, nil
	}
	if !errors.Is(err, content.ErrPostNotFound) && errx.KindOf(err) != errx.InvalidInput {
		return nil, false, err
	}

	id, err := uuid.Parse(slug)
	if err != nil {
		return nil, false, content.ErrPostNotFound
	}
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		if errx.KindOf(err) == errx.Forbidden {
			return nil, false, err
		}
		return nil, false, content.ErrPostNotFound
	}

	post, err = a.content.FindPostByID(ctx, namespace, id)
	if err != nil {
		return nil, false, content.ErrPostNotFound
	}
	return post, true, nil
}
