// Package permissions evaluates authorization for the verified request identity.
package permissions

import (
	"context"
	"fmt"

	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// Checker evaluates identity and admin-role requirements for the verified user.
// A shared ban gate may record an inconclusive lookup so privileges fail closed.
type Checker struct {
	lookupAdmin func(context.Context, string) (bool, error)
}

type banLookupErrorKey struct{}

// WithBanLookupError records a failed shared ban lookup for later privilege checks.
func WithBanLookupError(ctx context.Context, err error) context.Context {
	if err == nil {
		return ctx
	}
	return context.WithValue(ctx, banLookupErrorKey{}, err)
}

func NewChecker(lookupAdmin func(context.Context, string) (bool, error)) *Checker {
	return &Checker{lookupAdmin: lookupAdmin}
}

// KetoReader is the permission lookup required by the checker.
type KetoReader interface {
	CheckPermission(context.Context, string, string, string, ketoclient.Subject) (bool, error)
}

// NewKetoChecker checks admin membership using the shared application relation.
func NewKetoChecker(client KetoReader) *Checker {
	if client == nil {
		return NewChecker(nil)
	}
	return NewChecker(func(ctx context.Context, subjectID string) (bool, error) {
		return client.CheckPermission(ctx, "app", "tadoku", "admins", ketoclient.Subject{ID: subjectID})
	})
}

func (c *Checker) IsAdmin(ctx context.Context) (bool, error) {
	user := identity.FromContext(ctx)
	if user == nil || user.Subject == "" || user.Subject == "guest" {
		return false, nil
	}
	if err, _ := ctx.Value(banLookupErrorKey{}).(error); err != nil {
		return false, fmt.Errorf("%w: check ban permission: %w", commondomain.ErrAuthzUnavailable, err)
	}
	if c == nil || c.lookupAdmin == nil {
		return false, commondomain.ErrAuthzUnavailable
	}

	allowed, err := c.lookupAdmin(ctx, user.Subject)
	if err != nil {
		return false, fmt.Errorf("%w: check admin permission: %w", commondomain.ErrAuthzUnavailable, err)
	}
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("%w: check admin permission: %w", commondomain.ErrAuthzUnavailable, err)
	}
	return allowed, nil
}

func (c *Checker) RequireAuthenticated(ctx context.Context) error {
	user := identity.FromContext(ctx)
	if user == nil || user.Subject == "" || user.Subject == "guest" {
		return commondomain.ErrUnauthorized
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
		return commondomain.ErrForbidden
	}
	return nil
}
