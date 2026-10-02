package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

func TestBanLookupFailureBlocksPrivilegeChecks(t *testing.T) {
	providerErr := errors.New("ban lookup failed")
	for _, test := range []struct {
		name  string
		check func(*permissions.Checker, context.Context) (bool, error)
	}{
		{name: "conditional admin privilege", check: func(checker *permissions.Checker, ctx context.Context) (bool, error) {
			return checker.IsAdmin(ctx)
		}},
		{name: "admin-only operation", check: func(checker *permissions.Checker, ctx context.Context) (bool, error) {
			err := checker.RequireAdmin(ctx)
			return err == nil, err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			banChecks := 0
			checker := permissions.NewKetoChecker(nil)
			rejectBanned := RejectBannedUsers(func(context.Context) permissions.Admission {
				banChecks++
				return permissions.AdmittedBanUnknown{Err: providerErr}
			}, slog.New(slog.NewTextHandler(io.Discard, nil)))
			handler := rejectBanned(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
				allowed, err := test.check(checker, r.Context())
				if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, providerErr) {
					t.Errorf("permission error=%v, want unavailable preserving ban lookup error", err)
				}
				if allowed {
					w.WriteHeader(stdhttp.StatusNoContent)
					return
				}
				w.WriteHeader(stdhttp.StatusServiceUnavailable)
			}))
			request := httptest.NewRequest(stdhttp.MethodGet, "/test/permissions", nil)
			request = request.WithContext(identity.WithUser(request.Context(), &identity.User{Subject: "admin"}))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != stdhttp.StatusServiceUnavailable {
				t.Errorf("status=%d, want %d", response.Code, stdhttp.StatusServiceUnavailable)
			}
			if banChecks != 1 {
				t.Errorf("ban checks=%d, want 1", banChecks)
			}
		})
	}
}

func TestRejectBannedUsersAllowsOnlyRoleIntrospection(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		pattern    string
		want       int
		wantCalled bool
	}{
		{
			name:       "authz role get pattern",
			method:     stdhttp.MethodGet,
			path:       "/authz/current-user/role",
			pattern:    authzRoleGetPattern,
			want:       stdhttp.StatusNoContent,
			wantCalled: true,
		},
		{
			name:   "matching path without pattern",
			method: stdhttp.MethodGet,
			path:   "/authz/current-user/role",
			want:   stdhttp.StatusForbidden,
		},
		{
			name:    "matching path with wrong pattern",
			method:  stdhttp.MethodGet,
			path:    "/authz/current-user/role",
			pattern: "POST /authz/current-user/role",
			want:    stdhttp.StatusForbidden,
		},
		{
			name:    "role subpath with role pattern still requires exact mux match",
			method:  stdhttp.MethodGet,
			path:    "/authz/current-user/role/extra",
			pattern: "GET /authz/current-user/role/extra",
			want:    stdhttp.StatusForbidden,
		},
		{
			name:    "permission check pattern",
			method:  stdhttp.MethodPost,
			path:    "/authz/permission/check",
			pattern: "POST /authz/permission/check",
			want:    stdhttp.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			called := false
			middleware := RejectBannedUsers(func(context.Context) permissions.Admission {
				return permissions.Banned{}
			}, slog.New(slog.NewTextHandler(io.Discard, nil)))
			handler := middleware(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
				called = true
				if !permissions.IsBanned(r.Context()) {
					t.Error("downstream request is missing confirmed ban")
				}
				w.WriteHeader(stdhttp.StatusNoContent)
			}))
			request := httptest.NewRequest(test.method, test.path, nil)
			request.Pattern = test.pattern
			request = request.WithContext(identity.WithUser(request.Context(), &identity.User{Subject: "banned"}))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.want {
				t.Errorf("status = %d, want %d", response.Code, test.want)
			}
			if called != test.wantCalled {
				t.Errorf("downstream called = %t, want %t", called, test.wantCalled)
			}
		})
	}
}

func TestRejectBannedUsersUsesRequestContext(t *testing.T) {
	var checkedContext context.Context
	rejectBanned := RejectBannedUsers(func(ctx context.Context) permissions.Admission {
		checkedContext = ctx
		return permissions.Admitted{}
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	handler := withRequestTimeout(100*time.Millisecond, rejectBanned(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	})))
	request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	request = request.WithContext(identity.WithUser(request.Context(), &identity.User{Subject: "user"}))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if checkedContext == nil {
		t.Fatal("ban check did not receive a context")
	}
	if _, ok := checkedContext.Deadline(); !ok {
		t.Error("ban check did not receive the request deadline")
	}
	if !errors.Is(checkedContext.Err(), context.Canceled) {
		t.Errorf("check context error=%v, want canceled after request", checkedContext.Err())
	}
}

func TestRejectBannedUsersLogsAndAllowsProviderTimeout(t *testing.T) {
	var logs strings.Builder
	providerErr := context.DeadlineExceeded
	rejectBanned := RejectBannedUsers(func(context.Context) permissions.Admission {
		return permissions.AdmittedBanUnknown{Err: providerErr}
	}, slog.New(slog.NewTextHandler(&logs, nil)))
	request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
	request = request.WithContext(identity.WithUser(request.Context(), &identity.User{Subject: "timed-out-user"}))
	response := httptest.NewRecorder()

	rejectBanned(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusNoContent)
	})).ServeHTTP(response, request)

	if response.Code != stdhttp.StatusNoContent {
		t.Errorf("status=%d, want %d", response.Code, stdhttp.StatusNoContent)
	}
	for _, text := range []string{"banned-user check unavailable", "timed-out-user", providerErr.Error()} {
		if !strings.Contains(logs.String(), text) {
			t.Errorf("log %q does not contain %q", logs.String(), text)
		}
	}
}
