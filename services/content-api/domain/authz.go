package domain

import (
	"context"

	"github.com/tadoku/tadoku/services/common/authz/roles"
)

func requireAdmin(ctx context.Context) error {
	return roles.RequireAdmin(ctx)
}

func isAdmin(ctx context.Context) bool {
	return roles.IsAdmin(ctx)
}

