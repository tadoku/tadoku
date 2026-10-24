package permissions

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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

func TestCheckerRequire(t *testing.T) {
	t.Parallel()
	prodCtx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/checker-0000000a")
	writer := NewRelationships(db.AppPool, testModel)
	admin, branchAdmin, user := uuid.New(), uuid.New(), uuid.New()
	grantAll(t, prodCtx, writer, adminRelationship(admin))
	grantAll(t, branchCtx, writer, adminRelationship(branchAdmin))
	checker := NewChecker(db.AppPool, testModel)

	cases := []struct {
		name        string
		ctx         context.Context
		subject     string
		wantAllowed bool
		wantRequire errx.Kind
	}{
		{"admin in its tenant", prodCtx, admin.String(), true, granted},
		{"user", prodCtx, user.String(), false, errx.Forbidden},
		{"production admin on a test tenant", branchCtx, admin.String(), false, errx.Forbidden},
		{"test tenant admin in its tenant", branchCtx, branchAdmin.String(), true, granted},
		{"test tenant admin in production", prodCtx, branchAdmin.String(), false, errx.Forbidden},
		{"nil UUID subject", prodCtx, uuid.Nil.String(), false, errx.Forbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := signedIn(tc.ctx, tc.subject)

			allowed, err := checker.Allowed(ctx, siteAdmin)
			requireErr := checker.Require(ctx, siteAdmin)

			if err != nil || allowed != tc.wantAllowed {
				t.Errorf("Allowed = %t, %v; want %t, nil", allowed, err, tc.wantAllowed)
			}
			checkKind(t, "Require", requireErr, tc.wantRequire)
		})
	}

	t.Run("allowed for an explicit user", func(t *testing.T) {
		adminAllowed, adminErr := checker.AllowedFor(prodCtx, siteAdmin, admin)
		userAllowed, userErr := checker.AllowedFor(prodCtx, siteAdmin, user)

		if adminErr != nil || !adminAllowed {
			t.Errorf("AllowedFor(admin) = %t, %v; want true, nil", adminAllowed, adminErr)
		}
		if userErr != nil || userAllowed {
			t.Errorf("AllowedFor(user) = %t, %v; want false, nil", userAllowed, userErr)
		}
	})
}

func TestCheckerFailsClosed(t *testing.T) {
	t.Parallel()
	ctx := productionContext(t)

	db := newTestDatabase(t)
	admin := uuid.New()
	grantAll(t, ctx, NewRelationships(db.AppPool, testModel), adminRelationship(admin))
	closedPool, err := pgxpool.NewWithConfig(ctx, db.AppPool.Config())
	if err != nil {
		t.Fatal(err)
	}
	closedPool.Close()
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	lookupErr := errors.New("ban lookup failed")

	working := NewChecker(db.AppPool, testModel)
	unknownBan := WithBanState(ctx, BanUnknown(lookupErr))

	cases := []struct {
		name    string
		checker *Checker
		ctx     context.Context
		check   Check
		cause   error
	}{
		{"storage error", NewChecker(closedPool, testModel), ctx, siteAdmin, nil},
		{"cancelled context", working, cancelled, siteAdmin, context.Canceled},
		{"unknown ban", working, unknownBan, siteAdmin, lookupErr},
		{"unknown permit", working, ctx, sites.Check(siteRoot, "own"), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adminCtx := signedIn(tc.ctx, admin.String())

			allowed, allowedErr := tc.checker.Allowed(adminCtx, tc.check)
			requireErr := tc.checker.Require(adminCtx, tc.check)

			if allowed {
				t.Error("Allowed = true, want false")
			}
			for call, err := range map[string]error{"Allowed": allowedErr, "Require": requireErr} {
				if errx.KindOf(err) != errx.Unavailable || (tc.cause != nil && !errors.Is(err, tc.cause)) {
					t.Errorf("%s error = %v, want unavailable wrapping %v", call, err, tc.cause)
				}
			}
		})
	}

	t.Run("unknown ban allows only reads", func(t *testing.T) {
		userCtx := signedIn(WithBanState(ctx, BanUnknown(lookupErr)), uuid.NewString())
		authErr := working.RequireAuthenticated(userCtx)
		readErr := working.RequireAuthenticatedAllowingUnknownBan(userCtx)

		if errx.KindOf(authErr) != errx.Unavailable || !errors.Is(authErr, lookupErr) {
			t.Errorf("RequireAuthenticated error = %v, want unavailable wrapping %v", authErr, lookupErr)
		}
		if readErr != nil {
			t.Errorf("RequireAuthenticatedAllowingUnknownBan error = %v, want nil", readErr)
		}
	})
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
