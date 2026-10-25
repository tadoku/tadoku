package permissionmodel

import (
	"github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

// New builds the permission model from every feature's namespace. Every composition root builds its model here, so
// add a feature's namespace to this list and nowhere else.
func New() (permissions.Model, error) {
	return permissions.NewModel(authz.Permissions)
}
