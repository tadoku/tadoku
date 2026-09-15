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
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

var api *testAPI
var legacyContent *legacyContentAPI
var legacyAuthentication http.Handler
var legacyBannedUsers http.Handler
var legacyPermissions http.Handler
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

	api, err = newTestAPI(context.Background(), keto.ReadURL())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, api.db.Close()) }()

	legacyContent, err = newLegacyContentAPI(context.Background(), api.db.DSN, authenticationJWKS.URL, keto.ReadURL())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, legacyContent.db.Close()) }()
	legacyBannedUsers = newLegacyBannedUsersHandler(authenticationJWKS.URL, keto.ReadURL())
	legacyPermissions = newLegacyPermissionsHandler(authenticationJWKS.URL, keto.ReadURL())

	return m.Run()
}

// testAPI owns the production handler and an in-process sentinel transport.
type testAPI struct {
	db      *testpostgres.Database
	handler *transport.Router
	proxied atomic.Int32
}

func newTestAPI(ctx context.Context, ketoReadURL string) (_ *testAPI, err error) {
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
	api.handler, err = newTestRouter(db, ketoReadURL)
	if err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	upstreams := transport.Upstreams{
		Authz:     "http://upstream.test",
		Content:   "http://upstream.test",
		Immersion: "http://upstream.test",
		Profile:   "http://upstream.test",
	}
	err = transport.RegisterProxyRoutes(api.handler, upstreams, api, time.Second, prometheus.NewRegistry(), logger)
	if err != nil {
		return nil, fmt.Errorf("register proxy routes: %w", err)
	}

	complete = true
	return api, nil
}

func newTestRouter(db *testpostgres.Database, ketoReadURL string) (*transport.Router, error) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reader := ketoclient.NewReadClient(ketoReadURL)
	permissionChecker := permissions.NewKetoChecker(reader)
	repository := content.NewAnnouncementsRepository(db.Pool)
	service := content.NewService(repository)
	application := app.New(service, db.Pool, permissionChecker)
	authenticate, err := transport.NewJWTAuthentication(authenticationJWKS.URL, time.Second)
	if err != nil {
		return nil, err
	}
	rejectBanned := transport.RejectBannedUsers(func(ctx context.Context, subjectID string) (bool, error) {
		return reader.CheckPermission(ctx, "app", "tadoku", "banned", ketoclient.Subject{ID: subjectID})
	}, logger)
	handler, err := transport.NewHandler(application, db.Pool.Ping, time.Second, logger, authenticate, rejectBanned)
	if err != nil {
		return nil, fmt.Errorf("create API handler: %w", err)
	}
	handler.HandleFunc("GET /test/authentication", authenticationSuccess)
	handler.HandleFunc("GET /test/banned", bannedUsersSuccess)
	handler.HandleFunc("GET /test/permissions/admin", requireAdmin(permissionChecker))
	handler.HandleFunc("GET /test/permissions/check", checkAdmin(permissionChecker))
	return handler, nil
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
