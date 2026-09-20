// Package authz owns public authorization behavior.
package authz

import "github.com/tadoku/tadoku/services/tadoku-api/internal/errx"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleBanned Role = "banned"
	RoleUser   Role = "user"
	RoleGuest  Role = "guest"
)

var (
	ErrPermissionNamespaceRequired = errx.NewInvalidInputError("namespace is required")
	ErrPermissionObjectRequired    = errx.NewInvalidInputError("object is required")
	ErrPermissionRelationRequired  = errx.NewInvalidInputError("relation is required")
)

type PermissionCheckParameters struct {
	Namespace string
	Object    string
	Relation  string
}

func (p PermissionCheckParameters) Validate() error {
	if p.Namespace == "" {
		return ErrPermissionNamespaceRequired
	}
	if p.Object == "" {
		return ErrPermissionObjectRequired
	}
	if p.Relation == "" {
		return ErrPermissionRelationRequired
	}
	return nil
}
