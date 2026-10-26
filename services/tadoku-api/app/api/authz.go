package api

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
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

type RoleUpdateParameters = authz.RoleUpdateParameters

func (a *Application) UpdateRole(ctx context.Context, parameters RoleUpdateParameters) error {
	if err := a.roles.RequireAdmin(ctx); err != nil {
		return err
	}

	actorID, err := identity.RequireActorID(ctx)
	if err != nil {
		return err
	}
	update, err := a.authorization.PrepareRoleUpdate(ctx, parameters)
	if err != nil {
		return err
	}

	action := "unban_user"
	if parameters.Role == authz.RoleBanned {
		action = "ban_user"
	}
	event := audit.Event{
		ActorID: actorID,
		Action:  action,
		Metadata: map[string]any{
			"target_user_id": parameters.UserID.String(),
			"new_role":       string(parameters.Role),
		},
		Description: parameters.Reason,
	}

	return postgres.RunInTransaction(ctx, a.db, func(ctx context.Context) error {
		if err := a.profile.LockUser(ctx, parameters.UserID); err != nil {
			return err
		}
		if err := a.authorization.UpdateRole(ctx, update); err != nil {
			return err
		}
		return a.audit.Record(ctx, event)
	})
}
