package permissions

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

// Checker evaluates the model's checks in the context tenant. Nothing is inherited from another tenant.
type Checker struct {
	model  Model
	tuples Tuples
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

func ConfirmedBan() BanState {
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

func NewChecker(tuples Tuples, model Model) *Checker {
	return &Checker{model: model, tuples: tuples}
}

// Require returns unauthorized when nobody is signed in, forbidden when the check is denied, and unavailable when
// the check fails or the request's ban state is unknown.
func (c *Checker) Require(ctx context.Context, check Check) error {
	allowed, err := c.Allowed(ctx, check)
	if err != nil {
		return err
	}
	if !allowed {
		return errx.NewForbiddenError("forbidden")
	}
	return nil
}

// Allowed evaluates the check for the signed-in user in the context tenant, with Require's errors.
func (c *Checker) Allowed(ctx context.Context, check Check) (bool, error) {
	if err := c.RequireAuthenticated(ctx); err != nil {
		return false, err
	}

	userID, ok := identity.ActorID(ctx)
	if !ok {
		return false, nil
	}
	return c.AllowedFor(ctx, check, userID)
}

// AllowedFor evaluates the check for userID in the context tenant without reading the request identity or ban
// state. A failed check, or one whose context ends before it returns, is unavailable.
func (c *Checker) AllowedFor(ctx context.Context, check Check, userID uuid.UUID) (bool, error) {
	allowed, err := c.model.check(ctx, c.tuples, check, userID)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return false, errx.NewUnavailableError("check "+check.permit+" permission", err)
	}
	return allowed, nil
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
