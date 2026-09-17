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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/content"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

var api *suite
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

	api, err = newTestAPI(context.Background(), keto)
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

	return m.Run()
}

// suite owns the stores a handler reads. Reset only those stores.
type suite struct {
	db      *testpostgres.Database
	keto    *testketo.Fixture
	handler *transport.Router
	proxied atomic.Int32
}

func newTestAPI(ctx context.Context, ketoFixture *testketo.Fixture) (_ *suite, err error) {
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

	handler, err := newTestRouter(db.Pool, ketoFixture.ReadURL())
	if err != nil {
		return nil, err
	}
	api := &suite{db: db, keto: ketoFixture, handler: handler}
	if err := registerSentinelProxy(api); err != nil {
		return nil, err
	}

	complete = true
	return api, nil
}

func newTestRouter(pool *pgxpool.Pool, ketoReadURL string) (*transport.Router, error) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	reader := ketoclient.NewReadClient(ketoReadURL)
	permissionChecker := permissions.NewKetoChecker(reader)
	repository := content.NewAnnouncementsRepository(pool)
	service := content.NewService(repository)
	application := app.New(service, pool, permissionChecker)
	authenticate, err := transport.NewJWTAuthentication(authenticationJWKS.URL, time.Second)
	if err != nil {
		return nil, err
	}
	rejectBanned := transport.RejectBannedUsers(func(ctx context.Context, subjectID string) (bool, error) {
		return reader.CheckPermission(ctx, "app", "tadoku", "banned", ketoclient.Subject{ID: subjectID})
	}, logger)
	handler, err := transport.NewHandler(application, pool.Ping, time.Second, logger, authenticate, rejectBanned)
	if err != nil {
		return nil, fmt.Errorf("create API handler: %w", err)
	}
	handler.HandleFunc("GET /test/authentication", observeIdentity)
	handler.HandleFunc("GET /test/banned", observeIdentity)
	return handler, nil
}

func registerSentinelProxy(s *suite) error {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	upstreams := transport.Upstreams{
		Authz:     "http://upstream.test",
		Content:   "http://upstream.test",
		Immersion: "http://upstream.test",
		Profile:   "http://upstream.test",
	}
	if err := transport.RegisterProxyRoutes(s.handler, upstreams, s, time.Second, prometheus.NewRegistry(), logger); err != nil {
		return fmt.Errorf("register proxy routes: %w", err)
	}
	return nil
}

func (s *suite) RoundTrip(request *http.Request) (*http.Response, error) {
	s.proxied.Add(1)
	return &http.Response{
		StatusCode: http.StatusNoContent,
		Header:     http.Header{"X-Proxied": {"yes"}},
		Body:       http.NoBody,
		Request:    request,
	}, nil
}

func (s *suite) reset(t *testing.T, caseDir string) {
	t.Helper()
	requireKnownCaseFiles(t, caseDir)

	var postgresSeeds, ketoSeeds []string
	if caseDir != "" {
		postgresSeeds = []string{filepath.Join(caseDir, "setup.sql")}
		ketoSeeds = []string{filepath.Join(caseDir, "relationships.json")}
	}
	if s.db != nil {
		if err := s.db.Reset(t.Context(), postgresSeeds...); err != nil {
			t.Fatal(err)
		}
	}
	if s.keto != nil {
		if err := s.keto.Reset(t.Context(), ketoSeeds...); err != nil {
			t.Fatal(err)
		}
	}
	s.proxied.Store(0)
}

func (s *suite) resetProxyCount() {
	s.proxied.Store(0)
}

func requireKnownCaseFiles(t *testing.T, caseDir string) {
	t.Helper()
	if caseDir == "" {
		return
	}

	if err := checkKnownCaseFiles(caseDir); err != nil {
		t.Fatal(err)
	}
}

func checkKnownCaseFiles(caseDir string) error {
	entries, err := os.ReadDir(caseDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		switch entry.Name() {
		case "request.http", "golden.http", "setup.sql", "relationships.json":
		default:
			return fmt.Errorf("unknown entry %q in case directory %s", entry.Name(), caseDir)
		}
	}

	return nil
}

func TestRequireKnownCaseFilesRejectsUnknownFilename(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "relationship.json"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	err := checkKnownCaseFiles(dir)
	if err == nil {
		t.Fatal("expected unknown filename to fail")
	}

	want := fmt.Sprintf("unknown entry %q in case directory %s", "relationship.json", dir)
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err, want)
	}
}

func TestRequireKnownCaseFilesRejectsUnexpectedDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "extra"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := checkKnownCaseFiles(dir)
	if err == nil {
		t.Fatal("expected unexpected directory to fail")
	}

	want := fmt.Sprintf("unknown entry %q in case directory %s", "extra", dir)
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err, want)
	}
}

func openClosedPool(t *testing.T, dsn string) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(t.Context(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	pool.Close()
	return pool
}
