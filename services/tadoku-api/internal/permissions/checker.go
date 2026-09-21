// Package permissions evaluates authorization for the verified request identity.
package permissions

import (
	"context"

	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// Checker evaluates identity and admin-role requirements for the verified user.
// A shared ban gate may record an inconclusive lookup so privileges fail closed.
type Checker struct {
	client *ketoclient.Client
}

type banLookupErrorKey struct{}
type bannedKey struct{}

// WithBanLookupError records a failed shared ban lookup for later privilege checks.
func WithBanLookupError(ctx context.Context, err error) context.Context {
	if err == nil {
		return ctx
	}
	return context.WithValue(ctx, banLookupErrorKey{}, err)
}

// WithBanned records a confirmed ban for operations that need to report it.
func WithBanned(ctx context.Context) context.Context {
	return context.WithValue(ctx, bannedKey{}, true)
}

// IsBanned reports whether the shared ban lookup confirmed a ban.
func IsBanned(ctx context.Context) bool {
	banned, _ := ctx.Value(bannedKey{}).(bool)
	return banned
}

// NewKetoChecker checks admin membership using the shared application relation.
func NewKetoChecker(client *ketoclient.Client) *Checker {
	return &Checker{client: client}
}

func (c *Checker) CheckPermission(ctx context.Context, namespace, object, relation string) (bool, error) {
	if err := c.RequireAuthenticated(ctx); err != nil {
		return false, err
	}
	if c == nil || c.client == nil {
		return false, errx.NewUnavailableError("permissions unavailable", nil)
	}

	user := identity.FromContext(ctx)
	allowed, err := c.client.CheckPermission(ctx, namespace, object, relation, ketoclient.Subject{ID: user.Subject})
	if err != nil {
		return false, errx.NewUnavailableError("check permission", err)
	}
	if err := ctx.Err(); err != nil {
		return false, errx.NewUnavailableError("check permission", err)
	}
	return allowed, nil
}

func (c *Checker) IsAdmin(ctx context.Context) (bool, error) {
	user := identity.FromContext(ctx)
	if user == nil || user.Subject == "" || user.Subject == "guest" {
		return false, nil
	}
	if err, _ := ctx.Value(banLookupErrorKey{}).(error); err != nil {
		return false, errx.NewUnavailableError("check ban permission", err)
	}
	if c == nil || c.client == nil {
		return false, errx.NewUnavailableError("permissions unavailable", nil)
	}

	allowed, err := c.client.CheckPermission(ctx, "app", "tadoku", "admins", ketoclient.Subject{ID: user.Subject})
	if err != nil {
		return false, errx.NewUnavailableError("check admin permission", err)
	}
	if err := ctx.Err(); err != nil {
		return false, errx.NewUnavailableError("check admin permission", err)
	}
	return allowed, nil
}

func (c *Checker) RequireAuthenticated(ctx context.Context) error {
	if err := c.RequireAuthenticatedAllowingUnknownBan(ctx); err != nil {
		return err
	}
	if err, _ := ctx.Value(banLookupErrorKey{}).(error); err != nil {
		return errx.NewUnavailableError("check ban permission", err)
	}
	return nil
}

// RequireAuthenticatedAllowingUnknownBan requires a verified user but permits an
// inconclusive shared ban lookup. It is only safe for read-only operations.
func (c *Checker) RequireAuthenticatedAllowingUnknownBan(ctx context.Context) error {
	user := identity.FromContext(ctx)
	if user == nil || user.Subject == "" || user.Subject == "guest" {
		return errx.NewUnauthorizedError("unauthorized")
	}
	return nil
}

func (c *Checker) RequireAdmin(ctx context.Context) error {
	if err := c.RequireAuthenticated(ctx); err != nil {
		return err
	}

	allowed, err := c.IsAdmin(ctx)
	if err != nil {
		return err
	}
	if !allowed {
		return errx.NewForbiddenError("forbidden")
	}
	return nil
}
