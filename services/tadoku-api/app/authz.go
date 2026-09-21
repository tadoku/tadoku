package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
)

type Role = authz.Role

func (a *Application) CurrentUserRole(ctx context.Context) (Role, error) {
	return a.authorization.CurrentUserRole(ctx)
}

type PermissionCheckParameters = authz.PermissionCheckParameters

func (a *Application) CheckPermission(ctx context.Context, parameters PermissionCheckParameters) error {
	return a.authorization.CheckPermission(ctx, parameters)
}

type RoleUpdateParameters = authz.RoleUpdateParameters

func (a *Application) UpdateRole(ctx context.Context, parameters RoleUpdateParameters) error {
	return a.authorization.UpdateRole(ctx, parameters)
}
