package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func (a *Application) ListPosts(ctx context.Context, namespace string, includeDrafts bool, pageSize, page int) (*posts.PostList, error) {
	if includeDrafts {
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
