package domain

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type UpdateUserRoleRepository interface {
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
	GetUserRole(ctx context.Context, userID string) (string, error)
	UpdateUserRole(ctx context.Context, req *UpdateUserRoleRequest, moderatorUserID uuid.UUID) error
}

type UpdateUserRoleRequest struct {
	UserID uuid.UUID
	Role   string // "user" or "banned"
	Reason string
}

type UpdateUserRole struct {
	repo UpdateUserRoleRepository
}

func NewUpdateUserRole(repo UpdateUserRoleRepository) *UpdateUserRole {
	return &UpdateUserRole{repo: repo}
}

func (s *UpdateUserRole) Execute(ctx context.Context, req *UpdateUserRoleRequest) error {
	// Check if user is authenticated
	if isGuest(ctx) {
		return ErrUnauthorized
	}

	// Only admins can update roles
	if err := requireAdmin(ctx); err != nil {
		return err
	}

	// Get session to extract moderator user ID
	session := commondomain.ParseUserIdentity(ctx)
	if session == nil {
		return ErrUnauthorized
	}
	moderatorUserID := uuid.MustParse(session.Subject)

	// Validate role
	if req.Role != "user" && req.Role != "banned" {
		return fmt.Errorf("%w: role must be 'user' or 'banned'", ErrRequestInvalid)
	}

	// Validate reason
	if req.Reason == "" {
		return fmt.Errorf("%w: reason is required", ErrRequestInvalid)
	}
	if len(req.Reason) > 1000 {
		return fmt.Errorf("%w: reason must be 1000 characters or less", ErrRequestInvalid)
	}

	// Verify target user exists
	exists, err := s.repo.UserExists(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("could not check if user exists: %w", err)
	}
	if !exists {
		return ErrNotFound
	}

	// Check if target user is an admin - admins cannot be banned
	targetRole, err := s.repo.GetUserRole(ctx, req.UserID.String())
	if err != nil {
		return fmt.Errorf("could not get target user role: %w", err)
	}
	if targetRole == "admin" {
		return fmt.Errorf("%w: cannot modify role of an admin user", ErrForbidden)
	}

	// Update the role
	return s.repo.UpdateUserRole(ctx, req, moderatorUserID)
}
