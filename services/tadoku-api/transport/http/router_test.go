package http

import (
	"context"
	"io"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
)

func TestReadinessUsesIndependentDeadline(t *testing.T) {
	passthrough := func(next stdhttp.Handler) stdhttp.Handler { return next }
	deadline := make(chan time.Duration, 1)
	ready := func(ctx context.Context) error {
		until, ok := ctx.Deadline()
		if !ok {
			deadline <- 0
		} else {
			deadline <- time.Until(until)
		}
		return context.DeadlineExceeded
	}
	router, err := NewHandler(
		app.New(nil, nil, nil),
		ready,
		30*time.Second,
		prometheus.NewRegistry(),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		passthrough,
		passthrough,
	)
	if err != nil {
		t.Fatal(err)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/readyz", nil))

	if response.Code != stdhttp.StatusServiceUnavailable {
		t.Errorf("status=%d, want %d", response.Code, stdhttp.StatusServiceUnavailable)
	}
	if got := <-deadline; got < time.Second || got > 3*time.Second {
		t.Errorf("readiness deadline=%v, want about 2s", got)
	}
}
