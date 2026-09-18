package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (a *Application) FindPageBySlug(ctx context.Context, namespace, slug string) (*pages.Page, error) {
	page, err := a.pages.FindPageBySlug(ctx, namespace, slug)
	if err == nil {
		page.PublishedAt = nil
		page.CreatedAt = nil
		page.UpdatedAt = nil
		return page, nil
	}
	if !errors.Is(err, pages.ErrPageNotFound) && errx.KindOf(err) != errx.InvalidInput {
		return nil, err
	}

	id, err := uuid.Parse(slug)
	if err != nil {
		return nil, pages.ErrPageNotFound
	}
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	page, err = a.pages.FindPageByID(ctx, namespace, id)
	if err != nil {
		return nil, pages.ErrPageNotFound
	}
	return page, nil
}

func (a *Application) ListPages(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*pages.PageList, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.pages.ListPages(ctx, namespace, includeDrafts, pageSize, page)
}

type CreatePageParameters = pages.CreatePageParameters

func (a *Application) CreatePage(ctx context.Context, parameters CreatePageParameters) (*pages.Page, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	var result *pages.Page
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		if err := a.pages.CreatePage(ctx, parameters); err != nil {
			return err
		}
		result, err = a.pages.FindPageByID(ctx, parameters.Namespace, parameters.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

type UpdatePageParameters = pages.UpdatePageParameters

func (a *Application) UpdatePage(ctx context.Context, parameters UpdatePageParameters) (*pages.Page, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	var result *pages.Page
	err := postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) (err error) {
		if err := a.pages.UpdatePage(ctx, parameters); err != nil {
			return err
		}
		result, err = a.pages.FindPageByID(ctx, parameters.Namespace, parameters.ID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (a *Application) DeletePage(ctx context.Context, namespace string, id uuid.UUID) error {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return err
	}

	return a.pages.DeletePage(ctx, namespace, id)
}
