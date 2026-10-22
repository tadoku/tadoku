package authz

import (
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleBanned Role = "banned"
	RoleUser   Role = "user"
	RoleGuest  Role = "guest"
)

var (
	ErrUserNotFound       = errx.NewNotFoundError("user not found")
	ErrAdminRoleProtected = errx.NewForbiddenError("cannot modify role of an admin user")
)

type RoleUpdateParameters struct {
	UserID uuid.UUID
	Role   Role
	Reason string
}

// RoleUpdate is a validated role change for an existing user; build it with Service.PrepareRoleUpdate.
type RoleUpdate struct {
	userID uuid.UUID
	banned bool
}

func (p RoleUpdateParameters) Validate() error {
	if p.UserID == uuid.Nil {
		return errx.NewInvalidInputError("user id is required")
	}
	if p.Role != RoleUser && p.Role != RoleBanned {
		return errx.NewInvalidInputError("role must be 'user' or 'banned'")
	}
	if p.Reason == "" {
		return errx.NewInvalidInputError("reason is required")
	}
	if len(p.Reason) > 1000 {
		return errx.NewInvalidInputError("reason must be at most 1000 bytes")
	}
	return nil
}
