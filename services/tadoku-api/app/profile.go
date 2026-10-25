package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/profile"
)

func (a *Application) ListUsers(ctx context.Context, pageSize, page int, query string) (*profile.UserList, error) {
	if err := a.roles.RequireAdmin(ctx); err != nil {
		return nil, err
	}

	return a.profile.ListUsers(ctx, pageSize, page, query, a.roleNames)
}

func (a *Application) roleNames(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	roles, err := a.roles.UserRoles(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	names := make(map[uuid.UUID]string, len(roles))
	for userID, role := range roles {
		names[userID] = string(role)
	}
	return names, nil
}

func (a *Application) FindProfile(ctx context.Context, userID uuid.UUID) (*profile.PublicProfile, error) {
	return a.profile.FindProfile(ctx, userID)
}
