package e2e_test

import (
	"bytes"
	"io"
	"maps"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

const primaryRole = "primary"

// fixtureTenant maps a test tenant that checked-in fixtures name, in tokens, relationships or cookies, to the
// scenario role that stands in for it.
type fixtureTenant struct {
	fixture string
	role    string
}

var alphaAsOther = fixtureTenant{fixture: "e2e/alpha-0000000a", role: "other"}

// fixtureScenario runs checked-in golden cases and journeys in a scenario's private tenants instead of a
// whole-database reset: primaryRole stands in for tadoku/prod and each fixtureTenant's role for its fixture tenant.
type fixtureScenario struct {
	*scenario
	roles map[string]string
}

func newFixtureScenario(t *testing.T, others ...fixtureTenant) *fixtureScenario {
	t.Helper()
	roles := map[string]string{tenant.Production().String(): primaryRole}
	names := []string{primaryRole}
	for _, other := range others {
		if _, err := tenant.ParseTestTenant(other.fixture); err != nil {
			t.Fatalf("fixture tenant %q: %v", other.fixture, err)
		}
		if _, duplicate := roles[other.fixture]; duplicate {
			t.Fatalf("fixture tenant %q is listed twice", other.fixture)
		}
		roles[other.fixture] = other.role
		names = append(names, other.role)
	}
	return &fixtureScenario{scenario: newScenario(t, names...), roles: roles}
}

// runScenarioCase is runCase in the case's own fixtureScenario, in parallel with other scenario cases. Call it from
// the case's subtest. A parent test whose cases all run as scenarios calls t.Parallel too; otherwise its scenario
// cases start once its serial cases have returned.
func runScenarioCase(t *testing.T, name string, want int, impl implementation, others ...fixtureTenant) {
	t.Helper()
	runScenarioCaseAt(t, name, want, fixtureInstant, impl, others...)
}

func runScenarioCaseAt(
	t *testing.T,
	name string,
	want int,
	businessTime time.Time,
	impl implementation,
	others ...fixtureTenant,
) {
	t.Helper()
	t.Parallel()

	dir := filepath.Join("testdata", name)
	t.Run(impl.name, func(t *testing.T) {
		newFixtureScenario(t, others...).runCase(t, dir, want, businessTime, impl)
	})
}

// runCase seeds the case into the scenario's tenants and checks its golden at businessTime with the implementation's
// UUID stream.
func (s *fixtureScenario) runCase(
	t *testing.T,
	directory string,
	want int,
	businessTime time.Time,
	impl implementation,
) {
	t.Helper()
	requireKnownCaseFiles(t, directory)
	setup := map[string][]string{primaryRole: fixtureSeeds(directory, "setup.sql")}
	for role, file := range roleSetups(t, directory) {
		if role == primaryRole || !slices.Contains(slices.Collect(maps.Values(s.roles)), role) {
			t.Fatalf("%s seeds role %q, which the case does not map with a fixtureTenant", file, role)
		}
		setup[role] = []string{file}
	}
	s.seedFixtures(t,
		setup,
		fixtureSeeds(directory, "finalize.sql"),
		fixtureSeeds(directory, "relationships.json"),
	)
	s.clock.Set(businessTime)
	s.useIDs(impl.uuidReader())
	s.checkCase(t, directory, want, impl)
}

// checkCase sends the case's request to the implementation and checks its golden, without seeding.
func (s *fixtureScenario) checkCase(t *testing.T, directory string, want int, impl implementation) {
	t.Helper()
	request := readHTTPRequest(t, directory)
	defer request.Body.Close()
	recorder := s.serveFixture(t, impl.handler, request, "")
	checkRecordedHTTPGolden(t, request, recorder, directory, want, *updateGoldens, s.normalize)
}

// runScenarioJourney is runJourneyWithSetup in the journey's own fixtureScenario, in parallel with other scenario
// tests. Call it from the journey's top-level test, which must not call t.Parallel itself. ids supplies the journey's
// generated UUIDs; nil uses the serial runner's fixed seed.
func runScenarioJourney(
	t *testing.T,
	handler http.Handler,
	name string,
	ids io.Reader,
	steps []step,
	others ...fixtureTenant,
) {
	t.Helper()
	runScenarioJourneyWithSetup(t, handler, name, nil, ids, steps, others...)
}

// runScenarioJourneyWithSetup is runScenarioJourney that runs afterSeed with the journey's scenario once its fixtures
// are seeded, before the first step. afterSeed may only touch that scenario's tenants.
func runScenarioJourneyWithSetup(
	t *testing.T,
	handler http.Handler,
	name string,
	afterSeed func(*testing.T, *fixtureScenario),
	ids io.Reader,
	steps []step,
	others ...fixtureTenant,
) {
	t.Helper()
	t.Parallel()

	directory, names, tokens := prepareJourney(t, name, steps)
	s := newFixtureScenario(t, others...)
	s.seedFixtures(t,
		map[string][]string{
			primaryRole: {filepath.Join(journeysDir, "setup.sql"), filepath.Join(directory, "setup.sql")},
		},
		nil,
		[]string{filepath.Join(journeysDir, "relationships.json"), filepath.Join(directory, "relationships.json")},
	)
	if afterSeed != nil {
		afterSeed(t, s)
	}
	if ids == nil {
		ids = rand.New(rand.NewSource(1))
	}
	s.useIDs(ids)

	runJourneySteps(t, s, handler, directory, names, steps, tokens)
}

// seedFixtures seeds each role's setup, keyed by role, then the account cast into every tenant, and finalize into the
// primary tenant alone. It gives each relationship row to the tenant standing in for the row's fixture tenant.
func (s *fixtureScenario) seedFixtures(t *testing.T, setup map[string][]string, finalize, relationshipFiles []string) {
	t.Helper()
	relationships := map[string][]fixtureRelationship{}
	for _, file := range relationshipFiles {
		rows, err := readRelationships(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range rows {
			role := s.roleFor(t, row.Tenant)
			row.Tenant = ""
			relationships[role] = append(relationships[role], row)
		}
	}

	s.seedTenant(t, primaryRole, setup[primaryRole], finalize, relationships[primaryRole])
	for _, role := range slices.Sorted(maps.Values(s.roles)) {
		if role != primaryRole {
			s.seedTenant(t, role, setup[role], nil, relationships[role])
		}
	}
}

// roleSetups finds the setup.<role>.sql files in the case's operation directory. Only scenario cases may seed them.
func roleSetups(t *testing.T, caseDir string) map[string]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(filepath.Dir(caseDir), "setup.*.sql"))
	if err != nil {
		t.Fatal(err)
	}

	setups := map[string]string{}
	for _, file := range files {
		role := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(file), "setup."), ".sql")
		setups[role] = file
	}
	return setups
}

func (s *fixtureScenario) roleFor(t testing.TB, fixture string) string {
	t.Helper()
	role, ok := s.roles[fixture]
	if !ok {
		t.Fatalf("fixture tenant %q has no scenario role; map it with a fixtureTenant", fixture)
	}
	return role
}

func (s *fixtureScenario) businessClock() *timex.Clock {
	return s.clock
}

// serveFixture sends the request with token as its bearer credential, or with its own Authorization headers when
// token is empty. Each fixture token in them that verifies is re-signed for the tenant standing in for its tenant, in
// place, and each fixture session cookie is renamed the same way. It refuses a request that names a fixture tenant
// anywhere else.
func (s *fixtureScenario) serveFixture(
	t *testing.T,
	handler http.Handler,
	request *http.Request,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()
	s.refuseFixtureTenants(t, request)
	s.renameSessionCookies(t, request)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	s.resignFixtureTokens(t, request)
	return s.serveRequest(handler, request)
}

func (s *fixtureScenario) resignFixtureTokens(t *testing.T, request *http.Request) {
	t.Helper()
	stand := func(fixture string) string {
		return s.tenant(t, s.roleFor(t, fixture)).String()
	}

	values := request.Header.Values("Authorization")
	if len(values) == 0 {
		return
	}
	resigned := make([]string, 0, len(values))
	for _, value := range values {
		scheme, token, found := strings.Cut(value, " ")
		if found {
			signed, err := resignFixtureToken(token, stand)
			if err != nil {
				t.Fatal(err)
			}
			value = scheme + " " + signed
		}
		resigned = append(resigned, value)
	}
	request.Header["Authorization"] = resigned
}

func (s *fixtureScenario) refuseFixtureTenants(t *testing.T, request *http.Request) {
	t.Helper()
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	request.Body = io.NopCloser(bytes.NewReader(body))

	text := request.RequestURI + "\n" + string(body)
	for name, values := range request.Header {
		if name != "Authorization" && name != "Cookie" {
			text += "\n" + strings.Join(values, "\n")
		}
	}
	for fixture := range s.roles {
		_, route, _ := strings.Cut(fixture, "/")
		if strings.Contains(text, fixture) || (fixture != tenant.Production().String() && strings.Contains(text, route)) {
			t.Fatalf("request names fixture tenant %q outside its token and session cookie; keep the case serial", fixture)
		}
	}
}

func (s *fixtureScenario) renameSessionCookies(t *testing.T, request *http.Request) {
	t.Helper()
	cookies := request.Cookies()
	if len(cookies) == 0 {
		return
	}

	names := map[string]string{}
	for fixture, role := range s.roles {
		names[sessionCookieName(fixture)] = sessionCookieName(s.tenant(t, role).String())
	}
	request.Header.Del("Cookie")
	for _, cookie := range cookies {
		name, ok := names[cookie.Name]
		if !ok && strings.HasPrefix(cookie.Name, "tadoku_session") {
			t.Fatalf("session cookie %q belongs to no mapped fixture tenant", cookie.Name)
		}
		if ok {
			cookie.Name = name
		}
		request.AddCookie(cookie)
	}
}

// normalize replaces each scenario tenant's generated route with {role}, wherever the tenant's key, session cookie
// name or recovery host carries it. tadoku/prod and every other text stay as they are.
func (s *fixtureScenario) normalize(text string) string {
	for role, key := range s.tenants {
		_, route, _ := strings.Cut(key.String(), "/")
		text = strings.ReplaceAll(text, route, "{"+role+"}")
	}
	return text
}

func (s *fixtureScenario) verifyQuery(t *testing.T, query string) string {
	t.Helper()
	return s.verifyRows(t, primaryRole, query)
}

func (s *fixtureScenario) cacheKeys(t *testing.T) []string {
	t.Helper()
	if err := leaderboardValkey.ownsLease(t.Context()); err != nil {
		t.Fatal(err)
	}

	var keys []string
	for _, key := range s.tenants {
		matched, err := leaderboardValkey.keysMatching(t.Context(), "tenant:"+key.String()+":*")
		if err != nil {
			t.Fatal(err)
		}
		keys = append(keys, matched...)
	}
	return keys
}

// workerStep completes the jobs of every scenario tenant, or with forTenant of the tenant standing in for it, each
// with a worker scoped to that tenant alone.
func (s *fixtureScenario) workerStep(t *testing.T, forTenant string) {
	t.Helper()
	roles := slices.Sorted(maps.Values(s.roles))
	if forTenant != "" {
		roles = []string{s.roleFor(t, forTenant)}
	}

	queued := 0
	for _, role := range roles {
		queued += scenarioJobs(t, s.tenant(t, role)).total
	}
	if queued == 0 {
		t.Fatal("API write did not enqueue jobs")
	}
	for _, role := range roles {
		s.runWorker(t, role)
	}
}

func sessionCookieName(key string) string {
	if key == tenant.Production().String() {
		return "tadoku_session"
	}
	return "tadoku_session_" + strings.ReplaceAll(key, "/", "_")
}
