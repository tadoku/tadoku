package authz

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	requestidentity "github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type Service struct {
	roles *Roles
	users identity.Reader
}

func NewService(roles *Roles, users identity.Reader) *Service {
	return &Service{roles: roles, users: users}
}

func (s *Service) ProxyAdminCheck(ctx context.Context, subject uuid.UUID) (bool, error) {
	if subject == uuid.Nil {
		return false, errx.NewInvalidInputError("subject must be a UUID")
	}

	return s.roles.IsProductionAdmin(ctx, subject.String())
}

func (s *Service) CurrentUserRole(ctx context.Context) (Role, error) {
	user := requestidentity.FromContext(ctx)
	if user.Subject == "" || user.Subject == "guest" {
		return RoleGuest, nil
	}
	if permissions.IsBanned(ctx) {
		return RoleBanned, nil
	}

	admin, err := s.roles.IsAdmin(ctx)
	if err != nil {
		return "", err
	}
	if admin {
		return RoleAdmin, nil
	}
	return RoleUser, nil
}

func (s *Service) UpdateRole(ctx context.Context, parameters RoleUpdateParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	exists, err := s.users.UserExists(ctx, parameters.UserID)
	if err != nil {
		return errx.NewUnavailableError("check whether target user exists", err)
	}
	if !exists {
		return ErrUserNotFound
	}

	roles, err := s.roles.UserRoles(ctx, []uuid.UUID{parameters.UserID})
	if err != nil {
		return errx.NewUnavailableError("fetch target claims", err)
	}
	if roles[parameters.UserID] == RoleAdmin {
		return ErrAdminRoleProtected
	}

	if err := s.roles.SetBanned(ctx, parameters.UserID, parameters.Role == RoleBanned); err != nil {
		return errx.NewUnavailableError("update banned role", err)
	}

	return nil
}
