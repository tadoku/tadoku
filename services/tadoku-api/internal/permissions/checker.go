package permissions

import (
	"context"

	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

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

type BanState struct {
	status banStatus
	err    error
}

func Banned() BanState {
	return BanState{status: banned}
}

func BanUnknown(err error) BanState {
	return BanState{status: banUnknown, err: err}
}

func WithBanState(ctx context.Context, state BanState) context.Context {
	return context.WithValue(ctx, banStateKey{}, state)
}

func banStateFromContext(ctx context.Context) BanState {
	state, _ := ctx.Value(banStateKey{}).(BanState)
	return state
}

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

func (c *Checker) CheckBanned(ctx context.Context, subjectID string) (bool, error) {
	return c.lookup(ctx, "app", "tadoku", "banned", subjectID)
}

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

// RequireAuthenticatedAllowingUnknownBan is safe only for read-only operations.
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
