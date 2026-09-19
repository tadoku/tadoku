// Package authz owns public authorization behavior.
package authz

import "github.com/tadoku/tadoku/services/tadoku-api/internal/errx"

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	RoleGuest Role = "guest"
)

type PermissionCheckParameters struct {
	Namespace string
	Object    string
	Relation  string
}

func (p PermissionCheckParameters) Validate() error {
	if p.Namespace == "" || p.Object == "" || p.Relation == "" {
		return errx.NewInvalidInputError("namespace, object, and relation are required")
	}
	return nil
}
