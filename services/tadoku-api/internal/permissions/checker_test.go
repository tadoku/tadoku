package permissions

import (
	"context"
	"errors"
	"testing"

	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
)

func TestIsAdminSkipsKetoWithoutAuthenticatedUser(t *testing.T) {
	tests := []struct {
		name string
		user *identity.User
	}{
		{name: "missing identity"},
		{name: "empty subject", user: &identity.User{}},
		{name: "guest", user: &identity.User{Subject: "guest"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()
			if test.user != nil {
				ctx = identity.WithUser(ctx, test.user)
			}

			allowed, err := NewKetoChecker(nil).IsAdmin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if allowed {
				t.Error("IsAdmin allowed an unauthenticated identity")
			}
		})
	}
}

func TestIsAdminFailsClosedWithoutKeto(t *testing.T) {
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "admin"})
	for _, test := range []struct {
		name    string
		checker *Checker
	}{
		{name: "nil client", checker: NewKetoChecker(nil)},
		{name: "nil receiver"},
	} {
		t.Run(test.name, func(t *testing.T) {
			allowed, err := test.checker.IsAdmin(ctx)
			if allowed || errx.KindOf(err) != errx.Unavailable {
				t.Errorf("IsAdmin=(%t, %v), want false and unavailable kind", allowed, err)
			}
		})
	}
}

func TestCheckerClassifiesKetoFailures(t *testing.T) {
	fixture, err := testketo.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	checker := NewKetoChecker(ketoclient.NewReadClient(fixture.ReadURL()))
	checks := []struct {
		name  string
		check func(context.Context) (bool, error)
	}{
		{name: "admin", check: checker.IsAdmin},
		{name: "configured permission", check: func(ctx context.Context) (bool, error) {
			return checker.CheckPermission(ctx, "app", "tadoku", "admins")
		}},
	}

	canceledCtx, cancel := context.WithCancel(identity.WithUser(t.Context(), &identity.User{Subject: "admin"}))
	cancel()
	for _, check := range checks {
		t.Run(check.name+" canceled", func(t *testing.T) {
			allowed, err := check.check(canceledCtx)
			if allowed || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, context.Canceled) {
				t.Errorf("check=(%t, %v), want false and unavailable wrapping context.Canceled", allowed, err)
			}
		})
	}

	if err := fixture.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "admin"})
	for _, check := range checks {
		t.Run(check.name+" provider down", func(t *testing.T) {
			allowed, err := check.check(ctx)
			if allowed || errx.KindOf(err) != errx.Unavailable {
				t.Errorf("check=(%t, %v), want false and unavailable kind", allowed, err)
			}
		})
	}
	if checker.IsAdminOrFalse(ctx) {
		t.Error("provider failure granted administrator read visibility")
	}
	if checker.IsAdminOrFalse(WithBanState(ctx, BanUnknown(errors.New("ban lookup failed")))) {
		t.Error("unknown ban status granted administrator read visibility")
	}
}

func TestAuthenticationRequirements(t *testing.T) {
	tests := []struct {
		name           string
		user           *identity.User
		wantAuth       errx.Kind
		wantAdmin      errx.Kind
		wantPermission errx.Kind
	}{
		{name: "missing identity", wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized, wantPermission: errx.Unauthorized},
		{name: "empty subject", user: &identity.User{}, wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized, wantPermission: errx.Unauthorized},
		{name: "guest", user: &identity.User{Subject: "guest"}, wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized, wantPermission: errx.Unauthorized},
		{name: "authenticated user", user: &identity.User{Subject: "reader"}, wantAdmin: errx.Unavailable, wantPermission: errx.Unavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()
			if test.user != nil {
				ctx = identity.WithUser(ctx, test.user)
			}
			checker := NewKetoChecker(nil)

			if err := checker.RequireAuthenticated(ctx); errx.KindOf(err) != test.wantAuth || (err == nil) != (test.wantAuth == errx.Unknown) {
				t.Errorf("RequireAuthenticated error=%v, want kind %v (nil for Unknown)", err, test.wantAuth)
			}
			if err := checker.RequireAdmin(ctx); errx.KindOf(err) != test.wantAdmin {
				t.Errorf("RequireAdmin error=%v, want kind %v", err, test.wantAdmin)
			}
			if allowed, err := checker.CheckPermission(ctx, "app", "tadoku", "admins"); allowed || errx.KindOf(err) != test.wantPermission {
				t.Errorf("CheckPermission=(%t, %v), want false and kind %v", allowed, err, test.wantPermission)
			}
		})
	}
}

func TestAuthenticationRequirementsHandleUnknownBan(t *testing.T) {
	providerErr := errors.New("ban lookup failed")
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "user"})
	ctx = WithBanState(ctx, BanUnknown(providerErr))

	checker := NewKetoChecker(nil)
	if err := checker.RequireAuthenticated(ctx); errx.KindOf(err) != errx.Unavailable || !errors.Is(err, providerErr) {
		t.Errorf("RequireAuthenticated error=%v, want unavailable preserving ban lookup error", err)
	}
	if allowed, err := checker.IsAdmin(ctx); allowed || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, providerErr) {
		t.Errorf("IsAdmin=(%t, %v), want false and unavailable preserving ban lookup error", allowed, err)
	}
	if allowed, err := checker.CheckPermission(ctx, "app", "tadoku", "admins"); allowed || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, providerErr) {
		t.Errorf("CheckPermission=(%t, %v), want false and unavailable preserving ban lookup error", allowed, err)
	}
	if err := checker.RequireAuthenticatedAllowingUnknownBan(ctx); err != nil {
		t.Errorf("RequireAuthenticatedAllowingUnknownBan error=%v, want nil", err)
	}
}
