package permissions

import (
	"context"
	"testing"

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
