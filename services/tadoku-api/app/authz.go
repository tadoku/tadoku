package app

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
)

func (a *Application) CurrentUserRole(ctx context.Context) (authz.Role, error) {
	return a.authorization.CurrentUserRole(ctx)
}

type PermissionCheckParameters = authz.PermissionCheckParameters

func (a *Application) CheckPermission(ctx context.Context, parameters PermissionCheckParameters) error {
	return a.authorization.CheckPermission(ctx, parameters)
}
