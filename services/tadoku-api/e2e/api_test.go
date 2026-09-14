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
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

var api *testAPI
var legacyContent *legacyContentAPI
var legacyAuthentication http.Handler
var legacyBannedUsers http.Handler
var authenticationJWKS *httptest.Server
var keto *testketo.Fixture

func TestMain(m *testing.M) {
	os.Exit(runTests(m))
}

func runTests(m *testing.M) (code int) {
	var cleanupErr error
	defer func() {
		if cleanupErr != nil {
			fmt.Fprintln(os.Stderr, cleanupErr)
			if code == 0 {
				code = 1
			}
		}
	}()

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
	keto, err = testketo.New(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, keto.Close()) }()

	api, err = newTestAPI(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, api.db.Close()) }()

	legacyContent, err = newLegacyContentAPI(context.Background(), api.db.DSN)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, legacyContent.db.Close()) }()
	legacyBannedUsers = newLegacyBannedUsersHandler(authenticationJWKS.URL, keto.ReadURL())

	return m.Run()
}

// testAPI owns the production handler and an in-process sentinel transport.
type testAPI struct {
	db                   *testpostgres.Database
	handler              *transport.Router
	authenticatedHandler *transport.Router
	proxied              atomic.Int32
}

func newTestAPI(ctx context.Context) (_ *testAPI, err error) {
	db, err := testpostgres.New(ctx)
	if err != nil {
		return nil, err
	}
	complete := false
	defer func() {
		if !complete {
			err = errors.Join(err, db.Close())
		}
	}()
	api := &testAPI{db: db}

	repository := content.NewAnnouncementsRepository(api.db.Pool)
	service := content.NewService(repository)
	application := app.New(service)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	api.handler, err = transport.NewHandler(application, api.db.Pool.Ping, time.Second, logger, skipAuthentication, skipBanCheck)
	if err != nil {
		return nil, fmt.Errorf("create API handler: %w", err)
	}
	authenticate, err := transport.NewJWTAuthentication(authenticationJWKS.URL, time.Second)
	if err != nil {
		return nil, err
	}
	reader := ketoclient.NewReadClient(keto.ReadURL())
	rejectBanned := transport.RejectBannedUsers(func(ctx context.Context, subjectID string) (bool, error) {
		return reader.CheckPermission(ctx, "app", "tadoku", "banned", ketoclient.Subject{ID: subjectID})
	}, logger)
	api.authenticatedHandler, err = transport.NewHandler(application, api.db.Pool.Ping, time.Second, logger, authenticate, rejectBanned)
	if err != nil {
		return nil, fmt.Errorf("create authenticated API handler: %w", err)
	}
	api.authenticatedHandler.HandleFunc("GET /test/authentication", authenticationSuccess)
	api.authenticatedHandler.HandleFunc("GET /test/banned", bannedUsersSuccess)

	upstreams := transport.Upstreams{
		Authz:     "http://upstream.test",
		Content:   "http://upstream.test",
		Immersion: "http://upstream.test",
		Profile:   "http://upstream.test",
	}
	for _, handler := range []*transport.Router{api.handler, api.authenticatedHandler} {
		err = transport.RegisterProxyRoutes(
			handler,
			upstreams,
			api,
			time.Second,
			prometheus.NewRegistry(),
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("register proxy routes: %w", err)
		}
	}

	complete = true
	return api, nil
}

// Endpoint-contract scenarios deliberately skip both shared gates. Production
// construction has no opt-out and always supplies the real middleware.
func skipAuthentication(next http.Handler) http.Handler { return next }
func skipBanCheck(next http.Handler) http.Handler       { return next }

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
	var postgresSeeds, ketoSeeds []string
	for _, seedFile := range seedFiles {
		switch filepath.Ext(seedFile) {
		case ".sql":
			postgresSeeds = append(postgresSeeds, seedFile)
		case ".json":
			ketoSeeds = append(ketoSeeds, seedFile)
		default:
			t.Fatalf("unsupported seed file: %s", seedFile)
		}
	}
	if err := api.db.Reset(t.Context(), postgresSeeds...); err != nil {
		t.Fatal(err)
	}
	if err := keto.Reset(t.Context(), ketoSeeds...); err != nil {
		t.Fatal(err)
	}
	api.proxied.Store(0)
}

func resetCase(t *testing.T, directory string) {
	t.Helper()
	reset(t,
		filepath.Join(directory, "setup.sql"),
		filepath.Join(directory, "relationships.json"),
	)
}
