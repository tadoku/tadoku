package permissions

import (
	"context"
	"errors"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

func TestIsAdminIdentityAndLookupBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		user    *identity.User
		allowed bool
	}{
		{name: "missing identity"},
		{name: "empty subject", user: &identity.User{}},
		{name: "guest", user: &identity.User{Subject: "guest"}},
		{name: "non-admin", user: &identity.User{Subject: "reader"}},
		{name: "admin", user: &identity.User{Subject: "admin"}, allowed: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()
			if test.user != nil {
				ctx = identity.WithUser(ctx, test.user)
			}

			calls := 0
			checker := NewChecker(func(gotCtx context.Context, subject string) (bool, error) {
				calls++
				if gotCtx != ctx {
					t.Error("lookup did not receive request context")
				}
				if subject != test.user.Subject {
					t.Errorf("subject=%q, want %q", subject, test.user.Subject)
				}
				return test.allowed, nil
			})

			got, err := checker.IsAdmin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.allowed {
				t.Errorf("IsAdmin=%t, want %t", got, test.allowed)
			}
			wantCalls := 1
			if test.user == nil || test.user.Subject == "" || test.user.Subject == "guest" {
				wantCalls = 0
			}
			if calls != wantCalls {
				t.Errorf("lookup calls=%d, want %d", calls, wantCalls)
			}
		})
	}
}

func TestIsAdminFailsClosed(t *testing.T) {
	providerErr := errors.New("provider failed")
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "admin"})

	checker := NewChecker(func(got context.Context, _ string) (bool, error) {
		if got != ctx {
			t.Error("lookup did not receive request context")
		}
		return true, providerErr
	})
	allowed, err := checker.IsAdmin(ctx)
	if allowed {
		t.Error("IsAdmin allowed a provider result with an error")
	}
	for _, target := range []error{ErrUnavailable, providerErr} {
		if !errors.Is(err, target) {
			t.Errorf("error %v does not match %v", err, target)
		}
	}

	canceledCtx, cancel := context.WithCancel(identity.WithUser(t.Context(), &identity.User{Subject: "admin"}))
	checker = NewChecker(func(context.Context, string) (bool, error) {
		cancel()
		return true, nil
	})
	allowed, err = checker.IsAdmin(canceledCtx)
	if allowed || !errors.Is(err, ErrUnavailable) || !errors.Is(err, context.Canceled) {
		t.Errorf("IsAdmin=(%t, %v), want (false, ErrUnavailable wrapping context.Canceled)", allowed, err)
	}

	for _, test := range []struct {
		name    string
		checker *Checker
	}{
		{name: "zero value", checker: &Checker{}},
		{name: "nil receiver"},
	} {
		t.Run(test.name, func(t *testing.T) {
			allowed, err := test.checker.IsAdmin(identity.WithUser(t.Context(), &identity.User{Subject: "admin"}))
			if allowed || !errors.Is(err, ErrUnavailable) {
				t.Errorf("IsAdmin=(%t, %v), want (false, ErrUnavailable)", allowed, err)
			}
		})
	}
}

func TestRequirements(t *testing.T) {
	tests := []struct {
		name      string
		user      *identity.User
		allowed   bool
		lookupErr error
		wantAuth  error
		wantAdmin error
		wantCalls int
	}{
		{name: "missing identity", wantAuth: ErrUnauthorized, wantAdmin: ErrUnauthorized},
		{name: "empty subject", user: &identity.User{}, wantAuth: ErrUnauthorized, wantAdmin: ErrUnauthorized},
		{name: "guest", user: &identity.User{Subject: "guest"}, wantAuth: ErrUnauthorized, wantAdmin: ErrUnauthorized},
		{name: "non-admin", user: &identity.User{Subject: "reader"}, wantAdmin: ErrForbidden, wantCalls: 1},
		{name: "admin", user: &identity.User{Subject: "admin"}, allowed: true, wantCalls: 1},
		{name: "provider unavailable", user: &identity.User{Subject: "admin"}, lookupErr: context.DeadlineExceeded, wantAdmin: ErrUnavailable, wantCalls: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := t.Context()
			if test.user != nil {
				ctx = identity.WithUser(ctx, test.user)
			}

			calls := 0
			checker := NewChecker(func(context.Context, string) (bool, error) {
				calls++
				return test.allowed, test.lookupErr
			})

			if err := checker.RequireAuthenticated(ctx); !errors.Is(err, test.wantAuth) || (err != nil) != (test.wantAuth != nil) {
				t.Errorf("RequireAuthenticated error=%v, want %v", err, test.wantAuth)
			}
			err := checker.RequireAdmin(ctx)
			if !errors.Is(err, test.wantAdmin) || (err != nil) != (test.wantAdmin != nil) {
				t.Errorf("RequireAdmin error=%v, want %v", err, test.wantAdmin)
			}
			if test.lookupErr != nil && !errors.Is(err, test.lookupErr) {
				t.Errorf("RequireAdmin error=%v does not preserve %v", err, test.lookupErr)
			}
			if calls != test.wantCalls {
				t.Errorf("lookup calls=%d, want %d", calls, test.wantCalls)
			}
		})
	}
}
