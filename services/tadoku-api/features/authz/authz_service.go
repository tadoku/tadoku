package authz

import (
	"context"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type Service struct {
	permissions *permissions.Checker
}

func NewService(permissions *permissions.Checker) *Service {
	return &Service{permissions: permissions}
}

func (s *Service) CurrentUserRole(ctx context.Context) (Role, error) {
	user := identity.FromContext(ctx)
	if user == nil {
		return "", errx.NewUnauthorizedError("unauthorized")
	}
	if user.Subject == "" || user.Subject == "guest" {
		return RoleGuest, nil
	}

	admin, err := s.permissions.IsAdmin(ctx)
	if err != nil {
		return "", err
	}
	if admin {
		return RoleAdmin, nil
	}
	return RoleUser, nil
}

func (s *Service) CheckPermission(ctx context.Context, parameters PermissionCheckParameters) error {
	if err := s.permissions.RequireAuthenticated(ctx); err != nil {
		return err
	}
	if err := parameters.Validate(); err != nil {
		return err
	}

	// Production intentionally exposes no public permissions.
	return errx.NewForbiddenError("forbidden")
}
