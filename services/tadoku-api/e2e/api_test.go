package e2e_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
var legacyContent *legacyContentAPI
var legacyAuthentication http.Handler
var authenticationJWKS *httptest.Server

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) (code int) {
	jwks, err := os.ReadFile("testdata/authentication.jwks.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	authenticationJWKS = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwks)
	}))
	defer authenticationJWKS.Close()
	legacyAuthentication = newLegacyAuthenticationHandler(authenticationJWKS.URL)

	api, err = newTestAPI(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() {
		var legacyErr error
		if legacyContent != nil {
			legacyErr = legacyContent.db.Close()
		}
		if err := errors.Join(legacyErr, api.db.Close()); err != nil {
			fmt.Fprintln(os.Stderr, err)
			if code == 0 {
				code = 1
			}
		}
	}()

	legacyContent, err = newLegacyContentAPI(context.Background(), api.db.DSN)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return m.Run()
}

// testAPI owns the production handler and an in-process sentinel transport.
type testAPI struct {
	db                   *testpostgres.Database
	handler              *http.ServeMux
	authenticatedHandler *http.ServeMux
	authentication       http.Handler
	proxied              atomic.Int32
}

func newTestAPI(ctx context.Context) (*testAPI, error) {
	db, err := testpostgres.New(ctx)
	if err != nil {
		return nil, err
	}
	api := &testAPI{db: db}

	repository := content.NewAnnouncementsRepository(api.db.Pool)
	service := content.NewService(repository)
	application := app.New(service)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	api.handler, err = transport.NewHandler(application, api.db.Pool.Ping, time.Second, logger, withoutAuthentication)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("create API handler: %w", err), db.Close())
	}
	authenticate, err := transport.NewJWTAuthentication(authenticationJWKS.URL, time.Second)
	if err != nil {
		return nil, errors.Join(err, db.Close())
	}
	api.authenticatedHandler, err = transport.NewHandler(application, api.db.Pool.Ping, time.Second, logger, authenticate)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("create authenticated API handler: %w", err), db.Close())
	}
	api.authentication = authenticate(http.HandlerFunc(authenticationSuccess))

	upstreams := transport.Upstreams{
		Authz:     "http://upstream.test",
		Content:   "http://upstream.test",
		Immersion: "http://upstream.test",
		Profile:   "http://upstream.test",
	}
	for _, handler := range []*http.ServeMux{api.handler, api.authenticatedHandler} {
		err = transport.RegisterProxyRoutes(
			handler,
			upstreams,
			api,
			time.Second,
			prometheus.NewRegistry(),
			logger,
		)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("register proxy routes: %w", err), db.Close())
		}
	}

	return api, nil
}

// Endpoint contract scenarios deliberately exclude shared authentication.
// Production construction always supplies real authentication middleware.
func withoutAuthentication(next http.Handler) http.Handler { return next }

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
