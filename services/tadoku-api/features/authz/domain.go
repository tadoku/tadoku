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

type PermissionCheckParameters struct {
	Namespace string
	Object    string
	Relation  string
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
