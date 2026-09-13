package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

var api *testAPI

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) (code int) {
	var err error
	api, err = newTestAPI(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() {
		if err := api.db.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if code == 0 {
				code = 1
			}
		}
	}()

	return m.Run()
}

// testAPI owns the production handler and an in-process sentinel transport.
type testAPI struct {
	db      *testpostgres.Database
	handler http.Handler
	proxied atomic.Int32
}

func newTestAPI(ctx context.Context) (*testAPI, error) {
	db, err := testpostgres.New(ctx)
	if err != nil {
		return nil, err
	}
	api := &testAPI{db: db}

	repository := content.NewRepository(api.db.Pool)
	service := content.NewService(repository)
	application := app.New(service)
	upstreams := transport.Upstreams{
		Authz:     "http://upstream.test",
		Content:   "http://upstream.test",
		Immersion: "http://upstream.test",
		Profile:   "http://upstream.test",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	api.handler, err = transport.NewHandler(
		application,
		api.db.Pool.Ping,
		upstreams,
		api,
		time.Second,
		prometheus.NewRegistry(),
		logger,
	)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("create API handler: %w", err), db.Close())
	}

	return api, nil
}

func (a *testAPI) RoundTrip(request *http.Request) (*http.Response, error) {
	a.proxied.Add(1)
	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     http.Header{"X-Proxied": {"yes"}},
		Body:       http.NoBody,
		Request:    request,
	}, nil
}

func reset(t *testing.T, seedFiles ...string) {
	t.Helper()
	if err := api.db.Reset(t.Context(), seedFiles...); err != nil {
		t.Fatal(err)
	}
	api.proxied.Store(0)
}
