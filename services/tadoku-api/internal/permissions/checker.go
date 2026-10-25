package permissions

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

// Checker evaluates the model's checks in the context tenant, and fails as unavailable without one. Nothing is
// inherited from another tenant.
type Checker struct {
	model  Model
	tuples Tuples
}

var errNoTenant = errors.New("permissions require a tenant")

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
	answers, err := c.answer(ctx, []question{{check: check, user: userID}})
	if err != nil {
		return false, errx.NewUnavailableError("check "+check.permit+" permission", err)
	}
	return answers[0], nil
}

// AllowedEach evaluates every check for the signed-in user, reading storage once per rule level whatever the number
// of checks, for flags shown to the actor such as whether each row is editable. A guest or a user with a confirmed
// ban gets false for every check. An unknown ban state, or an error in any check, makes the whole call unavailable.
// Otherwise each answer equals Allowed's.
func (c *Checker) AllowedEach(ctx context.Context, checks ...Check) (map[Check]bool, error) {
	allowed := make(map[Check]bool, len(checks))
	if c.RequireAuthenticatedAllowingUnknownBan(ctx) != nil {
		return denyAll(allowed, checks), nil
	}
	if err := banLookupError(ctx); err != nil {
		return nil, errx.NewUnavailableError("check ban permission", err)
	}
	userID, ok := identity.ActorID(ctx)
	if !ok || IsBanned(ctx) {
		return denyAll(allowed, checks), nil
	}

	questions := make([]question, len(checks))
	for i, check := range checks {
		questions[i] = question{check: check, user: userID}
	}
	answers, err := c.answer(ctx, questions)
	if err != nil {
		return nil, errx.NewUnavailableError("check permissions", err)
	}
	for i, check := range checks {
		allowed[check] = answers[i]
	}
	return allowed, nil
}

// AllowedForEach evaluates the check for each user, like AllowedFor, reading storage once per rule level whatever
// the number of users. It reads no request identity or ban state and never authorizes the actor. An error for any
// user makes the whole call unavailable.
func (c *Checker) AllowedForEach(ctx context.Context, check Check, userIDs []uuid.UUID) (map[uuid.UUID]bool, error) {
	questions := make([]question, len(userIDs))
	for i, userID := range userIDs {
		questions[i] = question{check: check, user: userID}
	}
	answers, err := c.answer(ctx, questions)
	if err != nil {
		return nil, errx.NewUnavailableError("check "+check.permit+" permission", err)
	}

	allowed := make(map[uuid.UUID]bool, len(userIDs))
	for i, userID := range userIDs {
		allowed[userID] = answers[i]
	}
	return allowed, nil
}

func (c *Checker) answer(ctx context.Context, questions []question) ([]bool, error) {
	if _, ok := tenant.FromContext(ctx); !ok {
		return nil, errNoTenant
	}

	answers, err := c.model.evaluate(ctx, c.tuples, questions)
	if err == nil {
		err = ctx.Err()
	}
	return answers, err
}

func denyAll(allowed map[Check]bool, checks []Check) map[Check]bool {
	for _, check := range checks {
		allowed[check] = false
	}
	return allowed
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
