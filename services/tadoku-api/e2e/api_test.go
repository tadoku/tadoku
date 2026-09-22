package e2e_test

import (
	"context"
	"errors"
	"flag"
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
	commonroles "github.com/tadoku/tadoku/services/common/authz/roles"
	ketoclient "github.com/tadoku/tadoku/services/common/client/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/app"
	"github.com/tadoku/tadoku/services/tadoku-api/features/announcements"
	featureaudit "github.com/tadoku/tadoku/services/tadoku-api/features/audit"
	featureauthz "github.com/tadoku/tadoku/services/tadoku-api/features/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/features/contests"
	"github.com/tadoku/tadoku/services/tadoku-api/features/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
	"github.com/tadoku/tadoku/services/tadoku-api/features/pages"
	"github.com/tadoku/tadoku/services/tadoku-api/features/posts"
	featureprofile "github.com/tadoku/tadoku/services/tadoku-api/features/profile"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testkratos"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	transport "github.com/tadoku/tadoku/services/tadoku-api/transport/http"
)

var api *suite
var scoringEnabledHandler http.Handler
var legacyContent *legacyContentAPI
var legacyProfile *legacyProfileAPI
var legacyImmersion *legacyImmersionAPI
var legacyAuthentication http.Handler
var legacyBannedUsers http.Handler
var authenticationJWKS *httptest.Server
var keto *testketo.Fixture

const callbackToken = "test-oathkeeper-callback-token"

var unavailableCallback http.Handler

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(runTests(m))
}

func runTests(m *testing.M) (code int) {
	if *updateGoldens && os.Getenv("CI") != "" {
		fmt.Fprintln(os.Stderr, "-update-goldens is disabled in CI")
		return 1
	}

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	legacyAuthentication = newLegacyAuthenticationHandler(authenticationJWKS.URL)
	keto, err = testketo.New(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, keto.Close()) }()

	kratos, err := testkratos.New(ctx, "testdata/kratos.sql")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, kratos.Close()) }()

	api, err = newTestAPI(ctx, keto, kratos)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, api.db.Close()) }()

	scoringEnabledHandler, _, _, err = newTestRouterWithScoringEngine(ctx, api.db.Pool, api.db.Pool, keto, kratos, slog.New(slog.NewTextHandler(io.Discard, nil)), true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	unavailableKeto, err := testketo.New(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	unavailableNative, _, _, err := newTestRouterWithLogger(
		ctx,
		api.db.Pool,
		api.db.Pool,
		unavailableKeto,
		kratos,
		slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	if err != nil {
		_ = unavailableKeto.Close()
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := unavailableKeto.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	unavailableCallback = unavailableNative

	legacyContent, err = newLegacyContentAPI(ctx, api.db.DSN, authenticationJWKS.URL, keto.ReadURL())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, legacyContent.db.Close()) }()
	legacyProfile, err = newLegacyProfileAPI(ctx, api.db.DSN, authenticationJWKS.URL, keto.ReadURL(), kratos.CursorClient())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, legacyProfile.db.Close()) }()
	legacyImmersion, err = newLegacyImmersionAPI(ctx, api.db.DSN, authenticationJWKS.URL, keto.ReadURL(), kratos.Client())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() { cleanupErr = errors.Join(cleanupErr, legacyImmersion.db.Close()) }()
	legacyBannedUsers = newLegacyBannedUsersHandler(authenticationJWKS.URL, keto.ReadURL())

	return m.Run()
}

// suite owns the stores a handler reads. Reset only those stores.
type suite struct {
	db      *testpostgres.Database
	keto    *testketo.Fixture
	kratos  *testkratos.Fixture
	handler *transport.Router
	profile *featureprofile.Service
	roles   *commonroles.KetoService
	proxied atomic.Int32
}

func newTestAPI(ctx context.Context, ketoFixture *testketo.Fixture, kratosFixture *testkratos.Fixture) (_ *suite, err error) {
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

	handler, profileService, roleService, err := newTestRouter(ctx, db.Pool, ketoFixture, kratosFixture)
	if err != nil {
		return nil, err
	}
	api := &suite{
		db:      db,
		keto:    ketoFixture,
		kratos:  kratosFixture,
		handler: handler,
		profile: profileService,
		roles:   roleService,
	}
	if err := registerSentinelProxy(api); err != nil {
		return nil, err
	}

	complete = true
	return api, nil
}

func newTestRouter(
	ctx context.Context,
	pool *pgxpool.Pool,
	ketoFixture *testketo.Fixture,
	kratosFixture *testkratos.Fixture,
) (*transport.Router, *featureprofile.Service, *commonroles.KetoService, error) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return newTestRouterWithLogger(ctx, pool, pool, ketoFixture, kratosFixture, logger)
}

func newTestRouterWithLogger(
	ctx context.Context,
	pool *pgxpool.Pool,
	auditPool *pgxpool.Pool,
	ketoFixture *testketo.Fixture,
	kratosFixture *testkratos.Fixture,
	logger *slog.Logger,
) (*transport.Router, *featureprofile.Service, *commonroles.KetoService, error) {
	return newTestRouterWithScoringEngine(ctx, pool, auditPool, ketoFixture, kratosFixture, logger, false)
}

func newTestRouterWithScoringEngine(
	ctx context.Context,
	pool *pgxpool.Pool,
	auditPool *pgxpool.Pool,
	ketoFixture *testketo.Fixture,
	kratosFixture *testkratos.Fixture,
	logger *slog.Logger,
	scoringEngineEnabled bool,
) (*transport.Router, *featureprofile.Service, *commonroles.KetoService, error) {
	reader := ketoclient.NewReadClient(ketoFixture.ReadURL())
	readWriter := ketoclient.NewClient(ketoFixture.ReadURL(), ketoFixture.WriteURL())
	permissionChecker := permissions.NewKetoChecker(reader)
	roleService := commonroles.NewKetoService(reader, "app", "tadoku")
	identities := kratosFixture.CursorClient()
	authzService := featureauthz.NewService(
		permissionChecker,
		reader,
		identities,
		roleService,
		commonroles.NewKetoManager(readWriter, "app", "tadoku"),
		nil,
	)
	auditService := featureaudit.NewService(featureaudit.NewRepository(auditPool))
	announcementsRepository := announcements.NewAnnouncementsRepository(pool)
	contestsRepository := contests.NewContestsRepository(pool)
	languagesRepository := languages.NewLanguagesRepository(pool)
	logsRepository := logs.NewLogsRepository(pool)
	pagesRepository := pages.NewPagesRepository(pool)
	postsRepository := posts.NewPostsRepository(pool)
	profileRepository := featureprofile.NewRepository(pool)
	announcementsService := announcements.NewService(announcementsRepository)
	contestsService := contests.NewService(contestsRepository, kratosFixture.Client())
	languagesService := languages.NewService(languagesRepository)
	logsService := logs.NewService(logsRepository, scoringEngineEnabled)
	pagesService := pages.NewService(pagesRepository)
	postsService := posts.NewService(postsRepository)
	profileService := featureprofile.NewService(profileRepository, featureprofile.NewUserCache(identities), roleService, identities)
	application := app.New(announcementsService, auditService, authzService, contestsService, languagesService, logsService, pagesService, postsService, profileService, pool, permissionChecker)
	authenticate, err := transport.NewJWTAuthentication(ctx, authenticationJWKS.URL, time.Second, 24*time.Hour, "http://oathkeeper-api/", logger)
	if err != nil {
		return nil, nil, nil, err
	}
	rejectBanned := transport.RejectBannedUsers(func(ctx context.Context, subjectID string) (bool, error) {
		return reader.CheckPermission(ctx, "app", "tadoku", "banned", ketoclient.Subject{ID: subjectID})
	}, logger)
	authenticateCallback, err := transport.NewCallbackAuthentication(callbackToken)
	if err != nil {
		return nil, nil, nil, err
	}
	handler, err := transport.NewHandler(application, pool.Ping, time.Second, prometheus.NewRegistry(), logger, authenticate, rejectBanned, authenticateCallback)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create API handler: %w", err)
	}
	handler.HandleFunc("GET /test/authentication", observeIdentity)
	handler.HandleFunc("GET /test/banned", observeIdentity)
	return handler, profileService, roleService, nil
}

func registerSentinelProxy(s *suite) error {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	upstreams := transport.Upstreams{
		Immersion: "http://upstream.test",
		Profile:   "http://upstream.test",
	}
	if err := transport.RegisterProxyRoutes(s.handler, upstreams, s, time.Second, logger); err != nil {
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
	if s.kratos != nil {
		if err := s.kratos.Err(); err != nil {
			t.Fatal(err)
		}
	}
	requireKnownCaseFiles(t, caseDir)

	var postgresSeeds, ketoSeeds []string
	if caseDir != "" {
		postgresSeeds = fixtureSeeds(caseDir, "setup.sql")
		ketoSeeds = fixtureSeeds(caseDir, "relationships.json")
	}
	if s.db != nil {
		if err := s.db.Reset(t.Context(), postgresSeeds...); err != nil {
			t.Fatal(err)
		}
	}
	s.resetProfileCaches()
	if s.keto != nil {
		if err := s.keto.Reset(t.Context(), ketoSeeds...); err != nil {
			t.Fatal(err)
		}
	}
	s.proxied.Store(0)
}

func (s *suite) resetProfileCaches() {
	if s.profile != nil {
		*s.profile = *featureprofile.NewService(featureprofile.NewRepository(s.db.Pool), featureprofile.NewUserCache(s.kratos.CursorClient()), s.roles, s.kratos.CursorClient())
	}
	if legacyProfile != nil {
		legacyProfile.resetCache()
	}
}

func fixtureSeeds(caseDir, name string) []string {
	caseSeed := filepath.Join(caseDir, name)
	info, err := os.Stat(caseSeed)
	if err == nil && info.Mode().IsRegular() && info.Size() == 0 {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return []string{caseSeed}
	}
	if _, err := os.Lstat(caseSeed); !errors.Is(err, os.ErrNotExist) {
		return []string{caseSeed}
	}
	return []string{filepath.Join(filepath.Dir(caseDir), name)}
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

func TestFixtureSeed(t *testing.T) {
	testdata := t.TempDir()
	operationDir := filepath.Join(testdata, "Operation")
	caseDir := filepath.Join(operationDir, "200_case")
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		t.Fatal(err)
	}

	operationSeed := filepath.Join(operationDir, "setup.sql")
	if err := os.WriteFile(operationSeed, []byte("operation"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := fixtureSeeds(caseDir, "setup.sql"); len(got) != 1 || got[0] != operationSeed {
		t.Errorf("fallback seeds = %v, want [%s]", got, operationSeed)
	}

	caseSeed := filepath.Join(caseDir, "setup.sql")
	if err := os.WriteFile(caseSeed, []byte("case"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := fixtureSeeds(caseDir, "setup.sql"); len(got) != 1 || got[0] != caseSeed {
		t.Errorf("case override = %v, want [%s]", got, caseSeed)
	}
	if err := os.WriteFile(caseSeed, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkKnownCaseFiles(caseDir); err != nil {
		t.Errorf("zero-byte sentinel rejected by case guard: %v", err)
	}
	if got := fixtureSeeds(caseDir, "setup.sql"); len(got) != 0 {
		t.Errorf("zero-byte override = %v, want no seeds", got)
	}

	if err := os.Remove(caseSeed); err != nil {
		t.Fatal(err)
	}
	emptyTarget := filepath.Join(testdata, "empty.sql")
	if err := os.WriteFile(emptyTarget, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(emptyTarget, caseSeed); err != nil {
		t.Fatal(err)
	}
	if got := fixtureSeeds(caseDir, "setup.sql"); len(got) != 0 {
		t.Errorf("symlinked zero-byte override = %v, want no seeds", got)
	}
	if err := os.Remove(emptyTarget); err != nil {
		t.Fatal(err)
	}
	if got := fixtureSeeds(caseDir, "setup.sql"); len(got) != 1 || got[0] != caseSeed {
		t.Errorf("dangling case override = %v, want unreadable case seed [%s]", got, caseSeed)
	}
	if err := os.Remove(caseSeed); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(operationSeed); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(testdata, "setup.sql"), []byte("global"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := fixtureSeeds(caseDir, "setup.sql"); len(got) != 1 || got[0] != operationSeed {
		t.Errorf("bounded fallback = %v, want missing operation seed [%s]", got, operationSeed)
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
