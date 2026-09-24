package e2e_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
)

// Exercise each shared dependency failure once. Do not add endpoint-specific cases
// when the same dependency and failure behavior are already covered here.
func TestDependencyFailures(t *testing.T) {
	closedPool := openClosedPool(t, api.db.DSN)
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	handler, _, _, err := newTestRouterWithLogger(t.Context(), closedPool, closedPool, keto, api.kratos, logger)
	if err != nil {
		t.Fatal(err)
	}
	poolClosed := &suite{keto: keto, handler: handler}

	closedValkey, err := newClosedLeaderboardValkeyClient()
	if err != nil {
		t.Fatal(err)
	}
	cacheUnavailable, _, _, err := newTestRouterWithLeaderboard(t.Context(), api.db.Pool, api.db.Pool, keto, api.kratos, logger, false, closedValkey, 25*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	closedKeto, err := testketo.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	ketoUnavailable, _, _, err := newTestRouterWithLogger(t.Context(), api.db.Pool, api.db.Pool, closedKeto, api.kratos, logger)
	if err != nil {
		_ = closedKeto.Close()
		t.Fatal(err)
	}
	if err := closedKeto.Close(); err != nil {
		t.Fatal(err)
	}

	canceled := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		cancel()
		api.handler.ServeHTTP(w, r.WithContext(ctx))
	})

	tests := []struct {
		subtest     string
		operation   string
		description []string
		want        int
		suite       *suite
		handler     http.Handler
		logMessage  string
		logLevel    slog.Level
	}{
		{
			operation:   "FindAnnouncementByID",
			description: []string{"non", "admin"},
			want:        http.StatusForbidden,
			suite:       poolClosed,
			handler:     poolClosed.handler,
			logMessage:  "find announcement by ID rejected",
			logLevel:    slog.LevelDebug,
		},
		{
			operation:   "FindAnnouncementByID",
			description: []string{"read", "failure"},
			want:        http.StatusInternalServerError,
			suite:       poolClosed,
			handler:     poolClosed.handler,
			logMessage:  "find announcement by ID failed",
			logLevel:    slog.LevelError,
		},
		{
			operation:   "ListAnnouncements",
			description: []string{"non", "admin"},
			want:        http.StatusForbidden,
			suite:       poolClosed,
			handler:     poolClosed.handler,
			logMessage:  "list announcements rejected",
			logLevel:    slog.LevelDebug,
		},
		{
			operation:   "ListActiveAnnouncements",
			description: []string{"read", "failure"},
			want:        http.StatusInternalServerError,
			suite:       poolClosed,
			handler:     poolClosed.handler,
			logMessage:  "list active announcements failed",
			logLevel:    slog.LevelError,
		},
		{
			operation:   "ListAnnouncements",
			description: []string{"read", "failure"},
			want:        http.StatusInternalServerError,
			suite:       poolClosed,
			handler:     poolClosed.handler,
			logMessage:  "list announcements failed",
			logLevel:    slog.LevelError,
		},
		{
			subtest:     "canceled read",
			operation:   "ListActiveAnnouncements",
			description: []string{"read", "failure"},
			want:        499,
			suite:       api,
			handler:     canceled,
		},
		{
			operation:   "ImmersionFetchLeaderboardGlobal",
			description: []string{"cache", "unavailable"},
			want:        http.StatusOK,
			suite:       api,
			handler:     cacheUnavailable,
		},
		{
			operation:   "ImmersionFeatureAccessGet",
			description: []string{"provider", "unavailable"},
			want:        http.StatusServiceUnavailable,
			suite:       api,
			handler:     withFliptUnavailable(api.handler),
		},
		{
			operation:   "AuthzProxyProxyAdminCheck",
			description: []string{"provider", "unavailable"},
			want:        http.StatusServiceUnavailable,
			suite:       api,
			handler:     ketoUnavailable,
		},
	}

	for _, test := range tests {
		name := APITestName(test.operation, test.want, test.description...)
		subtest := test.subtest
		if subtest == "" {
			subtest = name
		}
		t.Run(subtest, func(t *testing.T) {
			logs.Reset()
			runCase(t, test.suite, name, test.want,
				implementation{name: "tadoku-api", handler: test.handler},
			)

			if test.logMessage == "" {
				return
			}
			wantLog := "level=" + test.logLevel.String() + " msg=\"" + test.logMessage + "\""
			if !strings.Contains(logs.String(), wantLog) {
				t.Errorf("logs do not contain %q:\n%s", wantLog, logs.String())
			}
			if test.logLevel < slog.LevelError && strings.Contains(logs.String(), "level=ERROR") {
				t.Errorf("client failure produced ERROR log:\n%s", logs.String())
			}
		})
	}

	t.Run("liveness survives but readiness fails", func(t *testing.T) {
		for _, probe := range []struct {
			path   string
			status int
		}{
			{path: "/livez", status: http.StatusOK},
			{path: "/readyz", status: http.StatusServiceUnavailable},
		} {
			response := httptest.NewRecorder()
			poolClosed.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, probe.path, nil))
			if response.Code != probe.status {
				t.Errorf("%s status=%d, want %d", probe.path, response.Code, probe.status)
			}
		}
	})
}
