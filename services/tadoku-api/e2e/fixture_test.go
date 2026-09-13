package e2e_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

type fixture struct {
	db      *testpostgres.Database
	handler http.Handler
	proxied atomic.Int32
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	f := &fixture{
		db: testpostgres.New(t),
	}

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.proxied.Add(1)
		w.Header().Set("X-Proxied", "yes")
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	repository := content.NewRepository(f.db.Pool)
	service := content.NewService(repository)
	application := app.New(service)
	upstreams := transport.Upstreams{
		Authz:     upstream.URL,
		Content:   upstream.URL,
		Immersion: upstream.URL,
		Profile:   upstream.URL,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	var err error
	f.handler, err = transport.NewHandler(
		application,
		f.db.Pool.Ping,
		upstreams,
		http.DefaultTransport,
		time.Second,
		prometheus.NewRegistry(),
		logger,
	)
	if err != nil {
		t.Fatalf("create API handler: %v", err)
	}

	return f
}
