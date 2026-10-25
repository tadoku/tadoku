package authz

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

func TestCheckerRequire(t *testing.T) {
	t.Parallel()
	prodCtx := productionContext(t)

	db := newTestDatabase(t)
	branchCtx := registerTestTenant(t, db, "e2e/checker-0000000a")
	repository := NewRepository(db.AppPool)
	writer := permissions.NewRelationships(repository, testModel)
	admin, branchAdmin, user := uuid.New(), uuid.New(), uuid.New()
	grantAll(t, prodCtx, writer, siteAdminRelationship(admin))
	grantAll(t, branchCtx, writer, siteAdminRelationship(branchAdmin))
	checker := permissions.NewChecker(repository, testModel)

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
	repository := NewRepository(db.AppPool)
	admin := uuid.New()
	grantAll(t, ctx, permissions.NewRelationships(repository, testModel), siteAdminRelationship(admin))
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	lookupErr := errors.New("ban lookup failed")

	working := permissions.NewChecker(repository, testModel)
	unknownBan := permissions.WithBanState(ctx, permissions.BanUnknown(lookupErr))

	cases := []struct {
		name    string
		checker *permissions.Checker
		ctx     context.Context
		check   permissions.Check
		cause   error
	}{
		{"storage error", permissions.NewChecker(NewRepository(closedPool(t, db)), testModel), ctx, siteAdmin, nil},
		{"cancelled context", working, cancelled, siteAdmin, context.Canceled},
		{"unknown ban", working, unknownBan, siteAdmin, lookupErr},
		{"unknown permit", working, ctx, permissions.Check{}, nil},
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
		userCtx := signedIn(permissions.WithBanState(ctx, permissions.BanUnknown(lookupErr)), uuid.NewString())
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
