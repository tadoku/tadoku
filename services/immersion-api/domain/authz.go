package domain

import (
	"context"

	"github.com/tadoku/tadoku/services/common/authz/roles"
)

func requireAdmin(ctx context.Context) error { return roles.RequireAdmin(ctx) }
func requireAuthentication(ctx context.Context) error {
	return roles.RequireAuthenticated(ctx)
}
func isAdmin(ctx context.Context) bool { return roles.IsAdmin(ctx) }
func isGuest(ctx context.Context) bool { return !roles.IsAuthenticated(ctx) }
