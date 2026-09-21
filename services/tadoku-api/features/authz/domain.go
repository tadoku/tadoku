// Package authz owns public authorization behavior.
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

type PermissionCheckParameters struct {
	Namespace string
	Object    string
	Relation  string
}

type PublicPermission struct {
	Namespace string
	Relation  string
}

type PublicPermissionAllowlist []PublicPermission

func (a PublicPermissionAllowlist) Allows(namespace, relation string) bool {
	for _, permission := range a {
		if permission.Namespace == namespace && permission.Relation == relation {
			return true
		}
	}
	return false
}

func (p PermissionCheckParameters) Validate() error {
	if p.Namespace == "" {
		return errx.NewInvalidInputError("namespace is required")
	}
	if p.Object == "" {
		return errx.NewInvalidInputError("object is required")
	}
	if p.Relation == "" {
		return errx.NewInvalidInputError("relation is required")
	}
	return nil
}

type RoleUpdateParameters struct {
	UserID uuid.UUID
	Role   Role
	Reason string
}

func (p RoleUpdateParameters) Validate() error {
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
