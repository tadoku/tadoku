package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// Call after RequireAuthenticated; the owner path performs no ban check.
func (a *Application) requireOwnerOrAdmin(ctx context.Context, ownerID uuid.UUID) error {
	if actorID, ok := identity.ActorID(ctx); ok && actorID == ownerID {
		return nil
	}
	return a.permissions.RequireAdmin(ctx)
}
