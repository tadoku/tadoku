package permissions

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

const granted = errx.Unknown

func signedIn(ctx context.Context, subject string) context.Context {
	return identity.WithUser(ctx, &identity.User{Subject: subject})
}

func checkKind(t *testing.T, call string, err error, want errx.Kind) {
	t.Helper()

	if errx.KindOf(err) != want || (err == nil) != (want == granted) {
		t.Errorf("%s error = %v, want kind %v (nil when granted)", call, err, want)
	}
}

func TestCheckerAuthentication(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	checker := NewChecker(nil, testModel)

	for name, userCtx := range map[string]context.Context{
		"missing identity": ctx,
		"empty subject":    signedIn(ctx, ""),
		"guest":            signedIn(ctx, "guest"),
	} {
		authErr := checker.RequireAuthenticated(userCtx)
		readErr := checker.RequireAuthenticatedAllowingUnknownBan(userCtx)
		requireErr := checker.Require(userCtx, siteAdmin)
		allowed, allowedErr := checker.Allowed(userCtx, siteAdmin)

		checkKind(t, name+" RequireAuthenticated", authErr, errx.Unauthorized)
		checkKind(t, name+" RequireAuthenticatedAllowingUnknownBan", readErr, errx.Unauthorized)
		checkKind(t, name+" Require", requireErr, errx.Unauthorized)
		checkKind(t, name+" Allowed", allowedErr, errx.Unauthorized)
		if allowed {
			t.Errorf("%s: Allowed = true, want false", name)
		}
	}
}

// TestCheckerBatchesWithoutStorage gives the checker no Tuples, so a batch that reached storage would panic.
func TestCheckerBatchesWithoutStorage(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	checker := NewChecker(nil, testModel)
	lookupErr := errors.New("ban lookup failed")
	userID := uuid.NewString()
	edit := contests.MustPermit("edit").On("c1")

	for name, userCtx := range map[string]context.Context{
		"missing identity":     ctx,
		"empty subject":        signedIn(ctx, ""),
		"guest":                signedIn(ctx, "guest"),
		"subject not a UUID":   signedIn(ctx, "admin"),
		"confirmed ban":        signedIn(WithBanState(ctx, ConfirmedBan()), userID),
		"guest with no tenant": signedIn(t.Context(), "guest"),
	} {
		allowed, err := checker.AllowedEach(userCtx, siteAdmin, edit)

		if err != nil || len(allowed) != 2 || allowed[siteAdmin] || allowed[edit] {
			t.Errorf("%s: AllowedEach = %v, %v; want every check false", name, allowed, err)
		}
	}

	t.Run("unknown ban", func(t *testing.T) {
		allowed, err := checker.AllowedEach(signedIn(WithBanState(ctx, BanUnknown(lookupErr)), userID), siteAdmin)

		if allowed != nil || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, lookupErr) {
			t.Errorf("AllowedEach = %v, %v; want unavailable wrapping %v", allowed, err, lookupErr)
		}
	})

	t.Run("missing tenant", func(t *testing.T) {
		_, eachErr := checker.AllowedEach(signedIn(t.Context(), userID))
		_, forEachErr := checker.AllowedForEach(t.Context(), siteAdmin, nil)
		_, forErr := checker.AllowedFor(t.Context(), siteAdmin, uuid.New())

		for call, err := range map[string]error{
			"AllowedEach":    eachErr,
			"AllowedForEach": forEachErr,
			"AllowedFor":     forErr,
		} {
			if errx.KindOf(err) != errx.Unavailable {
				t.Errorf("%s without a tenant error = %v, want unavailable", call, err)
			}
		}
	})
}
