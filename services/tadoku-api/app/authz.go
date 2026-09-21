package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/callbackauth"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type Role = authz.Role

func (a *Application) ProxyAdminCheck(ctx context.Context, subject uuid.UUID) (bool, error) {
	if !callbackauth.IsAuthenticated(ctx) {
		return false, errx.NewUnauthorizedError("unauthorized")
	}

	return a.authorization.ProxyAdminCheck(ctx, subject)
}

func (a *Application) CurrentUserRole(ctx context.Context) (Role, error) {
	if identity.FromContext(ctx) == nil {
		return "", errx.NewUnauthorizedError("unauthorized")
	}

	return a.authorization.CurrentUserRole(ctx)
}

type PermissionCheckParameters = authz.PermissionCheckParameters

func (a *Application) CheckPermission(ctx context.Context, parameters PermissionCheckParameters) (bool, error) {
	if err := a.permissions.RequireAuthenticated(ctx); err != nil {
		return false, err
	}

	return a.authorization.CheckPermission(ctx, parameters)
}

type RoleUpdateParameters = authz.RoleUpdateParameters

func (a *Application) UpdateRole(ctx context.Context, parameters RoleUpdateParameters) error {
	if err := a.permissions.RequireAdmin(ctx); err != nil {
		return err
	}

	actorID, err := uuid.Parse(identity.FromContext(ctx).Subject)
	if err != nil {
		return errx.NewUnauthorizedError("unauthorized")
	}

	if err := a.authorization.UpdateRole(ctx, parameters); err != nil {
		return err
	}

	action := "unban_user"
	if parameters.Role == authz.RoleBanned {
		action = "ban_user"
	}

	if err := a.audit.Record(ctx, audit.Event{
		ActorID: actorID,
		Action:  action,
		Metadata: map[string]string{
			"target_user_id": parameters.UserID.String(),
			"new_role":       string(parameters.Role),
		},
		Description: parameters.Reason,
	}); err != nil {
		return fmt.Errorf("create audit: %w", err)
	}

	return nil
}
