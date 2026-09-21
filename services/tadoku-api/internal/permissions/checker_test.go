package permissions

import (
	"context"
	"errors"
	"testing"

	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
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

func TestIsAdminClassifiesKetoFailures(t *testing.T) {
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
	canceledCtx, cancel := context.WithCancel(identity.WithUser(t.Context(), &identity.User{Subject: "admin"}))
	cancel()

	allowed, err := checker.IsAdmin(canceledCtx)
	if allowed || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, context.Canceled) {
		t.Errorf("IsAdmin with canceled context=(%t, %v), want false and unavailable wrapping context.Canceled", allowed, err)
	}

	if err := fixture.Close(); err != nil {
		t.Fatal(err)
	}
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "admin"})
	allowed, err = checker.IsAdmin(ctx)
	if allowed || errx.KindOf(err) != errx.Unavailable {
		t.Errorf("IsAdmin after Keto stopped=(%t, %v), want false and unavailable kind", allowed, err)
	}
}

func TestAuthenticationRequirements(t *testing.T) {
	tests := []struct {
		name      string
		user      *identity.User
		wantAuth  errx.Kind
		wantAdmin errx.Kind
	}{
		{name: "missing identity", wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized},
		{name: "empty subject", user: &identity.User{}, wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized},
		{name: "guest", user: &identity.User{Subject: "guest"}, wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized},
		{name: "authenticated user", user: &identity.User{Subject: "reader"}, wantAdmin: errx.Unavailable},
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
		})
	}
}

func TestAuthenticationRequirementsHandleUnknownBan(t *testing.T) {
	providerErr := errors.New("ban lookup failed")
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "user"})
	ctx = WithBanLookupError(ctx, providerErr)

	checker := NewKetoChecker(nil)
	if err := checker.RequireAuthenticated(ctx); errx.KindOf(err) != errx.Unavailable || !errors.Is(err, providerErr) {
		t.Errorf("RequireAuthenticated error=%v, want unavailable preserving ban lookup error", err)
	}
	if allowed, err := checker.IsAdmin(ctx); allowed || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, providerErr) {
		t.Errorf("IsAdmin=(%t, %v), want false and unavailable preserving ban lookup error", allowed, err)
	}
	if err := checker.RequireAuthenticatedAllowingUnknownBan(ctx); err != nil {
		t.Errorf("RequireAuthenticatedAllowingUnknownBan error=%v, want nil", err)
	}
}
