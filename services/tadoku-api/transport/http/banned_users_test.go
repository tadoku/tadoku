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

	commondomain "github.com/tadoku/tadoku/services/common/domain"
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
			adminChecks := 0
			checker := permissions.NewChecker(func(context.Context, string) (bool, error) {
				adminChecks++
				return true, nil
			})
			rejectBanned := RejectBannedUsers(func(context.Context, string) (bool, error) {
				banChecks++
				return false, providerErr
			}, slog.New(slog.NewTextHandler(io.Discard, nil)))
			handler := rejectBanned(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
				allowed, err := test.check(checker, r.Context())
				if !errors.Is(err, commondomain.ErrAuthzUnavailable) || !errors.Is(err, providerErr) {
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
			if adminChecks != 0 {
				t.Errorf("admin checks=%d, want 0", adminChecks)
			}
		})
	}
}

func TestRejectBannedUsersIdentityGuardsAndProviderErrors(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	for _, test := range []struct {
		name        string
		user        *identity.User
		banned      bool
		checkErr    error
		wantChecks  int
		wantSubject string
	}{
		{name: "missing identity skips check"},
		{name: "empty subject skips check", user: &identity.User{}},
		{name: "provider error wins over banned result", user: &identity.User{Subject: "uncertain"}, banned: true, checkErr: providerErr, wantChecks: 1, wantSubject: "uncertain"},
		{name: "provider failure fails open", user: &identity.User{Subject: "outage"}, checkErr: providerErr, wantChecks: 1, wantSubject: "outage"},
	} {
		t.Run(test.name, func(t *testing.T) {
			checks := 0
			var subject string
			middleware := RejectBannedUsers(func(_ context.Context, got string) (bool, error) {
				checks++
				subject = got
				return test.banned, test.checkErr
			}, slog.New(slog.NewTextHandler(io.Discard, nil)))
			downstreamCalls := 0
			handler := middleware(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
				downstreamCalls++
				w.WriteHeader(stdhttp.StatusNoContent)
			}))
			request := httptest.NewRequest(stdhttp.MethodGet, "/", nil)
			if test.user != nil {
				request = request.WithContext(identity.WithUser(request.Context(), test.user))
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != stdhttp.StatusNoContent {
				t.Errorf("status=%d, want %d", response.Code, stdhttp.StatusNoContent)
			}
			if checks != test.wantChecks || subject != test.wantSubject {
				t.Errorf("checks=%d subject=%q, want %d %q", checks, subject, test.wantChecks, test.wantSubject)
			}
			if downstreamCalls != 1 {
				t.Errorf("downstream calls=%d, want 1", downstreamCalls)
			}
			if response.Body.Len() != 0 {
				t.Errorf("response body=%q, want empty", response.Body.String())
			}
		})
	}
}

func TestRejectBannedUsersUsesRequestContext(t *testing.T) {
	var checkedContext context.Context
	rejectBanned := RejectBannedUsers(func(ctx context.Context, _ string) (bool, error) {
		checkedContext = ctx
		return false, nil
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
	rejectBanned := RejectBannedUsers(func(context.Context, string) (bool, error) {
		return false, providerErr
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
