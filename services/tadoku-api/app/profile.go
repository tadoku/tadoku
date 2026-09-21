package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
)

func (a *Application) ListUsers(ctx context.Context, pageSize, page int, query string) (*profile.UserList, error) {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.profile.ListUsers(ctx, pageSize, page, query)
}

func (a *Application) FindProfile(ctx context.Context, userID uuid.UUID) (*profile.PublicProfile, error) {
	return a.profile.FindProfile(ctx, userID)
}
