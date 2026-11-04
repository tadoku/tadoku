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
	s.seedFixtures(t,
		fixtureSeeds(directory, "setup.sql"),
		fixtureSeeds(directory, "finalize.sql"),
		fixtureSeeds(directory, "relationships.json"),
	)
	s.clock.Set(businessTime)
	s.useIDs(impl.uuidReader())

	request := readHTTPRequest(t, directory)
	defer request.Body.Close()
	token, found := strings.CutPrefix(request.Header.Get("Authorization"), "Bearer ")
	if !found && request.Header.Get("Authorization") != "" {
		t.Fatalf("%s/request.http carries a non-bearer Authorization header", directory)
	}
	recorder := s.serveFixture(t, impl.handler, request, token)
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
	t.Parallel()

	directory, names, tokens := prepareJourney(t, name, steps)
	s := newFixtureScenario(t, others...)
	s.seedFixtures(t,
		[]string{filepath.Join(journeysDir, "setup.sql"), filepath.Join(directory, "setup.sql")},
		nil,
		[]string{filepath.Join(journeysDir, "relationships.json"), filepath.Join(directory, "relationships.json")},
	)
	if ids == nil {
		ids = rand.New(rand.NewSource(1))
	}
	s.useIDs(ids)

	runJourneySteps(t, s, handler, directory, names, steps, tokens)
}

// seedFixtures seeds setup, the account cast and finalize into the primary tenant and the cast alone into every other
// tenant, then gives each relationship row to the tenant standing in for the row's fixture tenant.
func (s *fixtureScenario) seedFixtures(t *testing.T, setup, finalize, relationshipFiles []string) {
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

	s.seedTenant(t, primaryRole, setup, finalize, relationships[primaryRole])
	for _, role := range slices.Sorted(maps.Values(s.roles)) {
		if role != primaryRole {
			s.seedTenant(t, role, nil, nil, relationships[role])
		}
	}
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

// serveFixture re-signs the fixture token's cast claims for the tenant standing in for the token's tenant and renames
// each fixture session cookie the same way. It refuses a request that names a fixture tenant anywhere else.
func (s *fixtureScenario) serveFixture(
	t *testing.T,
	handler http.Handler,
	request *http.Request,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()
	s.refuseFixtureTenants(t, request)
	s.renameSessionCookies(t, request)

	role, as := primaryRole, none
	if token != "" {
		fixture, member, err := parseCastToken(token)
		if err != nil {
			t.Fatal(err)
		}
		role, as = s.roleFor(t, fixture), member
	}
	return s.serveWith(t, handler, role, as, request)
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
