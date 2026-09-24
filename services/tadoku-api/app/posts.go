package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (a *Application) FindPostBySlug(ctx context.Context, namespace, slug string) (*posts.Post, error) {
	post, err := a.posts.FindPostBySlug(ctx, namespace, slug)
	if err == nil {
		post.CreatedAt = nil
		post.UpdatedAt = nil
		return post, nil
	}
	if !errors.Is(err, posts.ErrPostNotFound) {
		return nil, err
	}

	id, err := uuid.Parse(slug)
	if err != nil {
		return nil, posts.ErrPostNotFound
	}
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.posts.FindPostByID(ctx, namespace, id)
}

func (a *Application) ListPosts(ctx context.Context, namespace string, includeDrafts *bool, pageSize, page int) (*posts.PostList, error) {
	if includeDrafts != nil && *includeDrafts {
		allowed, err := a.permissions.IsAdmin(ctx)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, errx.NewForbiddenError("drafts require admin access")
		}
	}

	return a.posts.ListPosts(ctx, namespace, includeDrafts, pageSize, page)
}

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
		return nil, err
	}
	return result, nil
}

type UpdatePostParameters = posts.UpdatePostParameters

func (a *Application) UpdatePost(ctx context.Context, parameters UpdatePostParameters) (*posts.Post, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	var result *posts.Post
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		if err := a.posts.UpdatePost(ctx, parameters); err != nil {
			return err
		}
		result, err = a.posts.FindPostByID(ctx, parameters.Namespace, parameters.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Application) DeletePost(ctx context.Context, namespace string, id uuid.UUID) error {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return err
	}

	return a.posts.DeletePost(ctx, namespace, id)
}

func (a *Application) ListPostVersions(ctx context.Context, namespace string, id uuid.UUID) ([]posts.PostVersion, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.posts.ListPostVersions(ctx, namespace, id)
}

func (a *Application) GetPostVersion(ctx context.Context, namespace string, postID, contentID uuid.UUID) (*posts.PostVersion, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.posts.GetPostVersion(ctx, namespace, postID, contentID)
}
