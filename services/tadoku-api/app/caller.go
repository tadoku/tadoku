package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// requireOwnerOrAdmin allows the owner of a resource and administrators. The
// owner path performs no ban check, so call it after RequireAuthenticated.
func (a *Application) requireOwnerOrAdmin(ctx context.Context, ownerID uuid.UUID) error {
	if callerID, ok := identity.CallerID(ctx); ok && callerID == ownerID {
		return nil
	}
	return a.permissions.RequireAdmin(ctx)
}
