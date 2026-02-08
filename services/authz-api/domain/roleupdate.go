package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type RoleUpdateUserDirectory interface {
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
}

type RoleUpdateRequest struct {
	UserID uuid.UUID
	Role   string // "user" or "banned"
	Reason string
}

type RoleUpdate struct {
	users    RoleUpdateUserDirectory
	audit    ModerationAuditRepository
	roles    commonroles.Service
	roleMgmt commonroles.Manager
}

func NewRoleUpdate(users RoleUpdateUserDirectory, audit ModerationAuditRepository, roles commonroles.Service, roleMgmt commonroles.Manager) *RoleUpdate {
	return &RoleUpdate{users: users, audit: audit, roles: roles, roleMgmt: roleMgmt}
}

func (s *RoleUpdate) Execute(ctx context.Context, req *RoleUpdateRequest) error {
	// Extract moderator user ID
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil || session.Subject == "guest" {
		return commondomain.ErrUnauthorized
	}

	moderatorClaims, err := s.roles.ClaimsForSubject(ctx, session.Subject)
	if err != nil {
		return commondomain.ErrAuthzUnavailable
	}
	if !moderatorClaims.Admin {
		return commondomain.ErrForbidden
	}

	moderatorUserID, err := uuid.Parse(session.Subject)
	if err != nil {
		return commondomain.ErrUnauthorized
	}

	// Validate role
	if req.Role != "user" && req.Role != "banned" {
		return fmt.Errorf("%w: role must be 'user' or 'banned'", commondomain.ErrRequestInvalid)
	}

	// Validate reason
	if req.Reason == "" {
		return fmt.Errorf("%w: reason is required", commondomain.ErrRequestInvalid)
	}
	if len(req.Reason) > 1000 {
		return fmt.Errorf("%w: reason must be 1000 characters or less", commondomain.ErrRequestInvalid)
	}

	// Verify target user exists
	exists, err := s.users.UserExists(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("could not check if user exists: %w", err)
	}
	if !exists {
		return commondomain.ErrNotFound
	}

	// Check if target user is an admin - admins cannot be banned
	targetClaims, err := s.roles.ClaimsForSubject(ctx, req.UserID.String())
	if err != nil {
		return commondomain.ErrAuthzUnavailable
	}
	if targetClaims.Admin {
		return fmt.Errorf("%w: cannot modify role of an admin user", commondomain.ErrForbidden)
	}

	// Update the role in Keto first (source of truth), then audit.
	switch req.Role {
	case "banned":
		if err := s.roleMgmt.SetBanned(ctx, req.UserID.String(), true); err != nil {
			return commondomain.ErrAuthzUnavailable
		}
	case "user":
		if err := s.roleMgmt.SetBanned(ctx, req.UserID.String(), false); err != nil {
			return commondomain.ErrAuthzUnavailable
		}
	default:
		return fmt.Errorf("%w: role must be 'user' or 'banned'", commondomain.ErrRequestInvalid)
	}

	action := "unban_user"
	if req.Role == "banned" {
		action = "ban_user"
	}

	auditReq := &ModerationAuditLogCreateRequest{
		ModeratorUserID: moderatorUserID,
		Action:          action,
		Metadata: map[string]any{
			"target_user_id": req.UserID.String(),
			"new_role":       req.Role,
		},
		Description: &req.Reason,
	}
	if err := s.audit.CreateModerationAuditLog(ctx, auditReq); err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return nil
}

