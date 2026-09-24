// Package permissions evaluates authorization for the verified request identity.
package permissions

import (
	"context"

	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// Checker owns the shared application Keto relations. CheckBanned and CheckAdmin
// look up a given subject; the other methods evaluate the verified actor.
// A shared ban gate may record an inconclusive lookup so privileges fail closed.
type Checker struct {
	client *ketoclient.Client
}

type banStateKey struct{}

type banStatus int

const (
	notBanned banStatus = iota
	banned
	banUnknown
)

// BanState is the outcome of the shared ban lookup. The zero value means not
// banned; only the unknown state carries the lookup error.
type BanState struct {
	status banStatus
	err    error
}

// Banned records a confirmed ban for operations that need to report it.
func Banned() BanState {
	return BanState{status: banned}
}

// BanUnknown records a failed shared ban lookup so privileges fail closed.
func BanUnknown(err error) BanState {
	return BanState{status: banUnknown, err: err}
}

// WithBanState records the shared ban lookup outcome for the request.
func WithBanState(ctx context.Context, state BanState) context.Context {
	return context.WithValue(ctx, banStateKey{}, state)
}

func banStateFromContext(ctx context.Context) BanState {
	state, _ := ctx.Value(banStateKey{}).(BanState)
	return state
}

// IsBanned reports whether the shared ban lookup confirmed a ban.
func IsBanned(ctx context.Context) bool {
	return banStateFromContext(ctx).status == banned
}

func banLookupError(ctx context.Context) error {
	state := banStateFromContext(ctx)
	if state.status != banUnknown {
		return nil
	}
	return state.err
}

// NewKetoChecker checks the shared application relations in Keto.
func NewKetoChecker(client *ketoclient.Client) *Checker {
	return &Checker{client: client}
}

func (c *Checker) CheckPermission(ctx context.Context, namespace, object, relation string) (bool, error) {
	if err := c.RequireAuthenticated(ctx); err != nil {
		return false, err
	}

	user := identity.FromContext(ctx)
	return c.lookup(ctx, namespace, object, relation, user.Subject)
}

// CheckBanned reports whether the subject holds the shared application ban relation.
func (c *Checker) CheckBanned(ctx context.Context, subjectID string) (bool, error) {
	return c.lookup(ctx, "app", "tadoku", "banned", subjectID)
}

// CheckAdmin reports whether the subject holds the shared application admin relation.
func (c *Checker) CheckAdmin(ctx context.Context, subjectID string) (bool, error) {
	return c.lookup(ctx, "app", "tadoku", "admins", subjectID)
}

func (c *Checker) lookup(ctx context.Context, namespace, object, relation, subjectID string) (bool, error) {
	if c == nil || c.client == nil {
		return false, errx.NewUnavailableError("permissions unavailable", nil)
	}

	allowed, err := c.client.CheckPermission(ctx, namespace, object, relation, ketoclient.Subject{ID: subjectID})
	if err != nil {
		return false, errx.NewUnavailableError("check "+relation+" permission", err)
	}
	if err := ctx.Err(); err != nil {
		return false, errx.NewUnavailableError("check "+relation+" permission", err)
	}
	return allowed, nil
}

func (c *Checker) IsAdmin(ctx context.Context) (bool, error) {
	user := identity.FromContext(ctx)
	if user == nil || user.Subject == "" || user.Subject == "guest" {
		return false, nil
	}
	if err := banLookupError(ctx); err != nil {
		return false, errx.NewUnavailableError("check ban permission", err)
	}

	return c.CheckAdmin(ctx, user.Subject)
}

// IsAdminOrFalse reports administrator access only after a successful lookup.
// Provider failures return false.
func (c *Checker) IsAdminOrFalse(ctx context.Context) bool {
	allowed, err := c.IsAdmin(ctx)
	return err == nil && allowed
}

func (c *Checker) RequireAuthenticated(ctx context.Context) error {
	if err := c.RequireAuthenticatedAllowingUnknownBan(ctx); err != nil {
		return err
	}
	if err := banLookupError(ctx); err != nil {
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
