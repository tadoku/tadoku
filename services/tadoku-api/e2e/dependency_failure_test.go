package e2e_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Exercise each shared dependency failure once. Do not add endpoint-specific cases
// when the same dependency and failure behavior are already covered here.
func TestDependencyFailures(t *testing.T) {
	closedPool := openClosedPool(t, api.db.DSN)
	handler, err := newTestRouter(closedPool, keto.ReadURL())
	if err != nil {
		t.Fatal(err)
	}
	poolClosed := &suite{keto: keto, handler: handler}
	if err := registerSentinelProxy(poolClosed); err != nil {
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
	}{
		{
			operation:   "FindAnnouncementByID",
			description: []string{"non", "admin"},
			want:        http.StatusForbidden,
			suite:       poolClosed,
			handler:     poolClosed.handler,
		},
		{
			operation:   "FindAnnouncementByID",
			description: []string{"read", "failure"},
			want:        http.StatusInternalServerError,
			suite:       poolClosed,
			handler:     poolClosed.handler,
		},
		{
			operation:   "ListAnnouncements",
			description: []string{"non", "admin"},
			want:        http.StatusForbidden,
			suite:       poolClosed,
			handler:     poolClosed.handler,
		},
		{
			operation:   "ListActiveAnnouncements",
			description: []string{"read", "failure"},
			want:        http.StatusInternalServerError,
			suite:       poolClosed,
			handler:     poolClosed.handler,
		},
		{
			operation:   "ListAnnouncements",
			description: []string{"read", "failure"},
			want:        http.StatusInternalServerError,
			suite:       poolClosed,
			handler:     poolClosed.handler,
		},
		{
			subtest:     "canceled read",
			operation:   "ListActiveAnnouncements",
			description: []string{"read", "failure"},
			want:        http.StatusInternalServerError,
			suite:       api,
			handler:     canceled,
		},
	}

	for _, test := range tests {
		name := APITestName(test.operation, test.want, test.description...)
		subtest := test.subtest
		if subtest == "" {
			subtest = name
		}
		t.Run(subtest, func(t *testing.T) {
			runCase(t, test.suite, name, test.want,
				implementation{name: "tadoku-api", handler: test.handler},
			)
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
