package roles

import (
	"context"
	"fmt"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type contextKey string

const ctxRolesKey contextKey = "roles.claims"

// Claims are authorization facts derived from an identity (typically via Keto).
// These are stored on the request context by middleware, and must be treated as
// request-scoped (not cached across requests).
type Claims struct {
	Subject       string
	Authenticated bool
	Admin         bool
	Banned        bool
	// Err is set when we could not evaluate authorization (e.g. Keto unavailable).
	Err error
}

func WithClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, ctxRolesKey, claims)
}

func FromContext(ctx context.Context) Claims {
	if v := ctx.Value(ctxRolesKey); v != nil {
		if c, ok := v.(Claims); ok {
			return c
		}
	}
	return Claims{}
}

func IsAuthenticated(ctx context.Context) bool { return FromContext(ctx).Authenticated }
func IsAdmin(ctx context.Context) bool         { return FromContext(ctx).Admin }
func IsBanned(ctx context.Context) bool        { return FromContext(ctx).Banned }

// RequireAuthenticated returns nil if the caller is authenticated.
// It returns commondomain.ErrUnauthorized if the caller is not authenticated.
// It returns commondomain.ErrAuthzUnavailable when we could not evaluate roles.
func RequireAuthenticated(ctx context.Context) error {
	c := FromContext(ctx)
	if !c.Authenticated {
		return commondomain.ErrUnauthorized
	}
	if c.Err != nil {
		return fmt.Errorf("%w: could not evaluate claims: %w", commondomain.ErrAuthzUnavailable, c.Err)
	}
	return nil
}

// RequireAdmin returns nil if the caller is an authenticated, non-banned admin.
// It returns commondomain.ErrAuthzUnavailable when we could not evaluate roles.
func RequireAdmin(ctx context.Context) error {
	c := FromContext(ctx)
	if !c.Authenticated {
		return commondomain.ErrUnauthorized
	}
	if c.Err != nil {
		return fmt.Errorf("%w: could not evaluate claims: %w", commondomain.ErrAuthzUnavailable, c.Err)
	}
	if c.Banned || !c.Admin {
		return commondomain.ErrForbidden
	}
	return nil
}
