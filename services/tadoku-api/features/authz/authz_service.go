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
	permissions *permissions.Checker
	users       identity.Reader
	roles       *permissions.KetoService
	roleManager *permissions.KetoManager
}

func NewService(
	permissions *permissions.Checker,
	users identity.Reader,
	roles *permissions.KetoService,
	roleManager *permissions.KetoManager,
) *Service {
	return &Service{
		permissions: permissions,
		users:       users,
		roles:       roles,
		roleManager: roleManager,
	}
}

func (s *Service) ProxyAdminCheck(ctx context.Context, subject uuid.UUID) (bool, error) {
	if subject == uuid.Nil {
		return false, errx.NewInvalidInputError("subject must be a UUID")
	}

	return s.permissions.IsProductionAdmin(ctx, subject.String())
}

func (s *Service) CurrentUserRole(ctx context.Context) (Role, error) {
	user := requestidentity.FromContext(ctx)
	if user.Subject == "" || user.Subject == "guest" {
		return RoleGuest, nil
	}
	if permissions.IsBanned(ctx) {
		return RoleBanned, nil
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

	claims, err := s.roles.RolesForSubject(ctx, parameters.UserID.String())
	if err != nil {
		return errx.NewUnavailableError("fetch target claims", err)
	}
	if claims.Admin {
		return ErrAdminRoleProtected
	}

	banned := parameters.Role == RoleBanned
	if err := s.roleManager.SetBanned(ctx, parameters.UserID.String(), banned); err != nil {
		return errx.NewUnavailableError("update banned role", err)
	}

	return nil
}
