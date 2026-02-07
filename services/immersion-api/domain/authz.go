package domain

import (
	"context"

	"github.com/tadoku/tadoku/services/common/authz/roles"
)

func requireAdmin(ctx context.Context) error { return roles.RequireAdmin(ctx) }
func isAdmin(ctx context.Context) bool       { return roles.IsAdmin(ctx) }
func isBanned(ctx context.Context) bool      { return roles.IsBanned(ctx) }
func isGuest(ctx context.Context) bool       { return !roles.IsAuthenticated(ctx) }

func requireNotBanned(ctx context.Context) error {
	claims := roles.FromContext(ctx)
	if claims.Err != nil {
		return ErrAuthzUnavailable
	}
	if claims.Banned {
		return ErrForbidden
	}
	return nil
}

func requireNotBannedUnauthorized(ctx context.Context) error {
	claims := roles.FromContext(ctx)
	if claims.Err != nil {
		return ErrAuthzUnavailable
	}
	if claims.Banned {
		return ErrUnauthorized
	}
	return nil
}
