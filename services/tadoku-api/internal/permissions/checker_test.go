package permissions

import (
	"context"
	"errors"
	"testing"

	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

type ketoLookup func(context.Context, string, string, string, ketoclient.Subject) (bool, error)

func (f ketoLookup) CheckPermission(ctx context.Context, namespace, object, relation string, subject ketoclient.Subject) (bool, error) {
	return f(ctx, namespace, object, relation, subject)
}

func TestKetoCheckerUsesAdminRelation(t *testing.T) {
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "reader"})
	providerErr := errors.New("keto unavailable")
	calls := 0
	checker := NewKetoChecker(ketoLookup(func(gotCtx context.Context, namespace, object, relation string, subject ketoclient.Subject) (bool, error) {
		calls++
		if gotCtx != ctx {
			t.Error("Keto did not receive request context")
		}
		if namespace != "app" || object != "tadoku" || relation != "admins" || subject.ID != "reader" || subject.Set != nil {
			t.Errorf("unexpected Keto tuple: %q %q %q %+v", namespace, object, relation, subject)
		}
		return true, providerErr
	}))
	allowed, err := checker.IsAdmin(ctx)
	if allowed || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, providerErr) {
		t.Errorf("IsAdmin=(%t, %v), want false and wrapped provider failure", allowed, err)
	}
	if calls != 1 {
		t.Errorf("Keto calls=%d, want 1", calls)
	}

	allowed, err = NewKetoChecker(nil).IsAdmin(ctx)
	if allowed || errx.KindOf(err) != errx.Unavailable {
		t.Errorf("nil Keto client: IsAdmin=(%t, %v), want false and unavailable kind", allowed, err)
	}
}

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
	if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, providerErr) {
		t.Errorf("error=%v, want unavailable kind preserving provider failure", err)
	}

	canceledCtx, cancel := context.WithCancel(identity.WithUser(t.Context(), &identity.User{Subject: "admin"}))
	checker = NewChecker(func(context.Context, string) (bool, error) {
		cancel()
		return true, nil
	})
	allowed, err = checker.IsAdmin(canceledCtx)
	if allowed || errx.KindOf(err) != errx.Unavailable || !errors.Is(err, context.Canceled) {
		t.Errorf("IsAdmin=(%t, %v), want false and unavailable kind wrapping context.Canceled", allowed, err)
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
			if allowed || errx.KindOf(err) != errx.Unavailable {
				t.Errorf("IsAdmin=(%t, %v), want false and unavailable kind", allowed, err)
			}
		})
	}
}

func TestRequireAuthenticatedIsIdentityOnly(t *testing.T) {
	ctx := identity.WithUser(t.Context(), &identity.User{Subject: "user"})
	ctx = WithBanLookupError(ctx, errors.New("ban lookup failed"))

	if err := (*Checker)(nil).RequireAuthenticated(ctx); err != nil {
		t.Errorf("RequireAuthenticated error=%v, want nil", err)
	}
}

func TestRequirements(t *testing.T) {
	tests := []struct {
		name      string
		user      *identity.User
		allowed   bool
		lookupErr error
		wantAuth  errx.Kind
		wantAdmin errx.Kind
		wantCalls int
	}{
		{name: "missing identity", wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized},
		{name: "empty subject", user: &identity.User{}, wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized},
		{name: "guest", user: &identity.User{Subject: "guest"}, wantAuth: errx.Unauthorized, wantAdmin: errx.Unauthorized},
		{name: "non-admin", user: &identity.User{Subject: "reader"}, wantAdmin: errx.Forbidden, wantCalls: 1},
		{name: "admin", user: &identity.User{Subject: "admin"}, allowed: true, wantCalls: 1},
		{name: "provider unavailable", user: &identity.User{Subject: "admin"}, lookupErr: context.DeadlineExceeded, wantAdmin: errx.Unavailable, wantCalls: 1},
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

			if err := checker.RequireAuthenticated(ctx); errx.KindOf(err) != test.wantAuth || (err == nil) != (test.wantAuth == errx.Unknown) {
				t.Errorf("RequireAuthenticated error=%v, want kind %v (nil for Unknown)", err, test.wantAuth)
			}
			err := checker.RequireAdmin(ctx)
			if errx.KindOf(err) != test.wantAdmin || (err == nil) != (test.wantAdmin == errx.Unknown) {
				t.Errorf("RequireAdmin error=%v, want kind %v (nil for Unknown)", err, test.wantAdmin)
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
