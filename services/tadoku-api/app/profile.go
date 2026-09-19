package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
)

func (a *Application) ListUsers(ctx context.Context, pageSize, page int, query string) (*profile.UserList, error) {
	return a.profile.ListUsers(ctx, pageSize, page, query)
}
