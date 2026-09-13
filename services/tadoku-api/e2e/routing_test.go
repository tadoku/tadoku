package e2e_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReadFailureDoesNotFallBack(t *testing.T) {
	// Closing the pool must not destroy the shared suite's database dependency.
	api, err := newTestAPI(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := api.db.Close(); err != nil {
			t.Error(err)
		}
	})
	api.db.Pool.Close()

	request := httptest.NewRequest(http.MethodGet, "/content/announcements/main/active", nil)
	response := httptest.NewRecorder()
	api.handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError || response.Body.Len() != 0 {
		t.Errorf("database failure: status=%d body=%s", response.Code, response.Body)
	}
	if api.proxied.Load() != 0 {
		t.Error("failed native read fell back to the proxy")
	}

	for path, status := range map[string]int{
		"/livez":  http.StatusOK,
		"/readyz": http.StatusServiceUnavailable,
	} {
		response := httptest.NewRecorder()
		api.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != status {
			t.Errorf("%s status=%d, want %d", path, response.Code, status)
		}
	}
}

func TestUnclaimedMethodsRemainProxied(t *testing.T) {
	for _, method := range []string{http.MethodHead, http.MethodOptions, http.MethodPost, http.MethodDelete, http.MethodPatch} {
		t.Run(method, func(t *testing.T) {
			reset(t)
			request := httptest.NewRequest(method, "/content/announcements/main/active", nil)
			response := httptest.NewRecorder()
			api.handler.ServeHTTP(response, request)

			if response.Header().Get("X-Proxied") != "yes" {
				t.Errorf("%s was not proxied", method)
			}
		})
	}
}

func TestCanceledNativeReadDoesNotFallBack(t *testing.T) {
	reset(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	request := httptest.NewRequest(http.MethodGet, "/content/announcements/main/active", nil).WithContext(ctx)
	response := httptest.NewRecorder()
	api.handler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Errorf("canceled read: status=%d", response.Code)
	}
	if api.proxied.Load() != 0 {
		t.Error("canceled native read fell back to the proxy")
	}
}
