package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// requireOwnerOrAdmin allows the actor when they own the resource and otherwise
// requires administrator access. The owner path performs no ban check, so call
// it after RequireAuthenticated.
func (a *Application) requireOwnerOrAdmin(ctx context.Context, ownerID uuid.UUID) error {
	if actorID, ok := identity.ActorID(ctx); ok && actorID == ownerID {
		return nil
	}
	return a.permissions.RequireAdmin(ctx)
}
