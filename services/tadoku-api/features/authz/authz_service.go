package authz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	kratosclient "github.com/tadoku/tadoku/services/tadoku-api/infra/kratos"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

type Service struct {
	permissions       *permissions.Checker
	users             *kratosclient.Client
	roles             *permissions.KetoService
	roleManager       *permissions.KetoManager
	publicPermissions PublicPermissionAllowlist
}

func NewService(
	permissions *permissions.Checker,
	users *kratosclient.Client,
	roles *permissions.KetoService,
	roleManager *permissions.KetoManager,
	publicPermissions PublicPermissionAllowlist,
) *Service {
	return &Service{
		permissions:       permissions,
		users:             users,
		roles:             roles,
		roleManager:       roleManager,
		publicPermissions: publicPermissions,
	}
}

func (s *Service) ProxyAdminCheck(ctx context.Context, subject uuid.UUID) (bool, error) {
	if subject == uuid.Nil {
		return false, errx.NewInvalidInputError("subject must be a UUID")
	}

	return s.permissions.CheckAdmin(ctx, subject.String())
}

func (s *Service) CurrentUserRole(ctx context.Context) (Role, error) {
	user := identity.FromContext(ctx)
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

func (s *Service) CheckPermission(ctx context.Context, parameters PermissionCheckParameters) (bool, error) {
	if err := parameters.Validate(); err != nil {
		return false, err
	}
	if !s.publicPermissions.Allows(parameters.Namespace, parameters.Relation) {
		return false, errx.NewForbiddenError("forbidden")
	}

	return s.permissions.CheckPermission(ctx, parameters.Namespace, parameters.Object, parameters.Relation)
}

func (s *Service) UpdateRole(ctx context.Context, parameters RoleUpdateParameters) error {
	if err := parameters.Validate(); err != nil {
		return err
	}

	exists, err := s.users.UserExists(ctx, parameters.UserID)
	if err != nil {
		return fmt.Errorf("check whether target user exists: %w", err)
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
