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

	"github.com/tadoku/tadoku/services/tadoku-api/internal/identity"
)

func TestRejectBannedUsers(t *testing.T) {
	providerErr := errors.New("provider unavailable")
	for _, test := range []struct {
		name        string
		user        *identity.User
		banned      bool
		checkErr    error
		want        int
		wantChecks  int
		wantSubject string
	}{
		{name: "missing identity skips check", want: stdhttp.StatusNoContent},
		{name: "empty subject skips check", user: &identity.User{}, want: stdhttp.StatusNoContent},
		{name: "guest skips check", user: &identity.User{Subject: "guest"}, want: stdhttp.StatusNoContent},
		{name: "allowed user", user: &identity.User{Subject: "allowed"}, want: stdhttp.StatusNoContent, wantChecks: 1, wantSubject: "allowed"},
		{name: "banned user", user: &identity.User{Subject: "banned"}, banned: true, want: stdhttp.StatusForbidden, wantChecks: 1, wantSubject: "banned"},
		{name: "provider error wins over banned result", user: &identity.User{Subject: "uncertain"}, banned: true, checkErr: providerErr, want: stdhttp.StatusNoContent, wantChecks: 1, wantSubject: "uncertain"},
		{name: "provider failure fails open", user: &identity.User{Subject: "outage"}, checkErr: providerErr, want: stdhttp.StatusNoContent, wantChecks: 1, wantSubject: "outage"},
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

			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
			if checks != test.wantChecks || subject != test.wantSubject {
				t.Errorf("checks=%d subject=%q, want %d %q", checks, subject, test.wantChecks, test.wantSubject)
			}
			wantDownstream := 1
			if test.want == stdhttp.StatusForbidden {
				wantDownstream = 0
			}
			if downstreamCalls != wantDownstream {
				t.Errorf("downstream calls=%d, want %d", downstreamCalls, wantDownstream)
			}
			if test.banned && response.Body.Len() != 0 {
				t.Errorf("forbidden body=%q, want empty", response.Body.String())
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
