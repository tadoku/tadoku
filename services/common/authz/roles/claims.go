package roles

import (
	"context"
	"fmt"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type contextKey string

const ctxRolesKey contextKey = "roles.claims"

// Claims are request-scoped; do not cache them across requests.
type Claims struct {
	Subject       string
	Authenticated bool
	Admin         bool
	Banned        bool
	Err           error
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
