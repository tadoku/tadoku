package authz

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	kratosclient "github.com/tadoku/tadoku/services/common/client/kratos"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type Service struct {
	permissions *permissions.Checker
	users       *kratosclient.Client
	roles       *commonroles.KetoService
	roleManager *commonroles.KetoManager
	audit       *AuthzRepository
}

func NewService(
	permissions *permissions.Checker,
	users *kratosclient.Client,
	roles *commonroles.KetoService,
	roleManager *commonroles.KetoManager,
	audit *AuthzRepository,
) *Service {
	return &Service{
		permissions: permissions,
		users:       users,
		roles:       roles,
		roleManager: roleManager,
		audit:       audit,
	}
}

func (s *Service) CurrentUserRole(ctx context.Context) (Role, error) {
	user := identity.FromContext(ctx)
	if user == nil {
		return "", errx.NewUnauthorizedError("unauthorized")
	}
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

func (s *Service) UpdateRole(ctx context.Context, parameters RoleUpdateParameters) error {
	if err := s.permissions.RequireAdmin(ctx); err != nil {
		return err
	}

	user := identity.FromContext(ctx)
	moderatorUserID, err := uuid.Parse(user.Subject)
	if err != nil {
		return errx.NewUnauthorizedError("unauthorized")
	}
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

	claims, err := s.roles.ClaimsForSubject(ctx, parameters.UserID.String())
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

	action := ModerationActionUnbanUser
	if banned {
		action = ModerationActionBanUser
	}
	if err := s.audit.CreateModerationAudit(ctx, ModerationAudit{
		ModeratorUserID: moderatorUserID,
		Action:          action,
		TargetUserID:    parameters.UserID,
		NewRole:         parameters.Role,
		Description:     parameters.Reason,
		CreatedAt:       timex.Now(),
	}); err != nil {
		return fmt.Errorf("create moderation audit: %w", err)
	}

	return nil
}
