package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

type member string

const (
	none   member = "none"
	guest  member = "guest"
	user   member = "user"
	user2  member = "user2"
	admin  member = "admin"
	banned member = "banned"

	alphaAdmin  member = "alpha_admin"
	alphaTester member = "alpha_tester"
	alphaUser   member = "alpha_user"
	alphaBanned member = "alpha_banned"
	alphaGuest  member = "alpha_guest"
)

type cast map[member]int

type step struct {
	request string
	verify  string
	job     string
	valkey  string

	forTenant string

	as     member
	want   int
	others cast

	at time.Time
}

const journeysDir = "testdata/journeys"

var stepNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

func runJourney(t *testing.T, s *suite, name string, steps []step) {
	runJourneyWithHandler(t, s, s.handler, name, steps)
}

func runJourneyWithHandler(t *testing.T, s *suite, handler http.Handler, name string, steps []step) {
	runJourneyWithSetup(t, s, handler, name, nil, nil, steps)
}

// runJourneyWithSetup runs afterReset after the journey's reset. ids supplies the bytes of the journey's generated
// UUIDs in creation order; nil uses a fixed seed.
func runJourneyWithSetup(
	t *testing.T,
	s *suite,
	handler http.Handler,
	name string,
	afterReset func(*testing.T),
	ids io.Reader,
	steps []step,
) {
	t.Helper()
	directory, names, tokens := prepareJourney(t, name, steps)

	resetJourney(t, s, directory)
	if afterReset != nil {
		afterReset(t)
	}

	if ids == nil {
		ids = rand.New(rand.NewSource(1))
	}
	clock := timex.NewClock(fixtureInstant)
	fixture := suiteJourney{suite: s, clock: clock, ctx: scenarioContext(t, clock, ids)}
	runJourneySteps(t, fixture, handler, directory, names, steps, tokens)
}

// journeyFixture is what a journey's steps run against: the suite after a whole-database reset, or a
// fixtureScenario's private tenants.
type journeyFixture interface {
	businessClock() *timex.Clock
	// serveFixture sends a checked-in request with a fixture token; an empty token sends no credentials.
	serveFixture(t *testing.T, handler http.Handler, request *http.Request, token string) *httptest.ResponseRecorder
	// normalize maps generated tenant keys in a response, verify result or cache key to stable aliases.
	normalize(text string) string
	verifyQuery(t *testing.T, query string) string
	cacheKeys(t *testing.T) []string
	workerStep(t *testing.T, forTenant string)
}

// suiteJourney runs a journey against the suite: requests carry their fixture tokens unchanged and share one context
// with the journey's clock and UUID stream.
type suiteJourney struct {
	suite *suite
	clock *timex.Clock
	ctx   context.Context
}

func (j suiteJourney) businessClock() *timex.Clock {
	return j.clock
}

func (j suiteJourney) serveFixture(
	t *testing.T,
	handler http.Handler,
	request *http.Request,
	token string,
) *httptest.ResponseRecorder {
	t.Helper()
	request = request.WithContext(j.ctx)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func (suiteJourney) normalize(text string) string {
	return text
}

func (j suiteJourney) verifyQuery(t *testing.T, query string) string {
	t.Helper()
	var rows string
	if err := j.suite.db.Pool.QueryRow(t.Context(), wrapVerifyQuery(query)).Scan(&rows); err != nil {
		t.Fatalf("run verify query: %v", err)
	}
	return rows
}

func (suiteJourney) cacheKeys(t *testing.T) []string {
	t.Helper()
	if err := leaderboardValkey.ownsLease(t.Context()); err != nil {
		t.Fatal(err)
	}
	keys, err := leaderboardValkey.cacheKeys(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	return keys
}

func (j suiteJourney) workerStep(t *testing.T, forTenant string) {
	t.Helper()
	runWorkerStep(t, j.ctx, j.suite, forTenant)
}

// prepareJourney checks the journey's step table, fixture files and cast before anything is seeded.
func prepareJourney(t *testing.T, name string, steps []step) (string, []string, map[member]string) {
	t.Helper()
	directory := filepath.Join(journeysDir, name)
	names, err := stepDirNames(steps)
	if err != nil {
		t.Fatal(err)
	}
	if err := checkJourneyFiles(directory, steps, names); err != nil {
		t.Fatal(err)
	}
	tokens, err := loadCast(steps)
	if err != nil {
		t.Fatal(err)
	}
	return directory, names, tokens
}

// runJourneySteps runs the steps in order, each at its own business time, and stops at the first failing step.
func runJourneySteps(
	t *testing.T,
	fixture journeyFixture,
	handler http.Handler,
	directory string,
	names []string,
	steps []step,
	tokens map[member]string,
) {
	t.Helper()
	for index, current := range steps {
		stepDir := filepath.Join(directory, names[index])
		at := current.at
		if at.IsZero() {
			at = fixtureInstant
		}
		fixture.businessClock().Set(at)

		passed := t.Run(names[index], func(t *testing.T) {
			switch {
			case current.verify != "":
				query := readVerifyQuery(t, stepDir)
				checkJSONGolden(t, stepDir, []byte(fixture.normalize(fixture.verifyQuery(t, query))))
			case current.valkey != "":
				checkValkeyGolden(t, fixture, stepDir)
			case current.job != "":
				runJobStep(t, fixture, current)
			default:
				runRequestStep(t, fixture, handler, stepDir, current, tokens)
			}
		})
		if !passed {
			return
		}
	}
}

func stepDirNames(steps []step) ([]string, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("journey has no steps")
	}

	names := make([]string, 0, len(steps))
	for index, current := range steps {
		name, err := current.dirName(index + 1)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (current step) dirName(position int) (string, error) {
	kinds := 0
	for _, value := range []string{current.request, current.verify, current.job, current.valkey} {
		if value != "" {
			kinds++
		}
	}
	if kinds != 1 {
		return "", fmt.Errorf("step %d must set exactly one request, verify, job or valkey", position)
	}

	name := current.request
	if current.request == "" && (current.as != "" || current.want != 0 || current.others != nil) {
		return "", fmt.Errorf("non-request step %d must not set as, want or others", position)
	}

	if current.verify != "" {
		name = current.verify
	} else if current.valkey != "" {
		name = current.valkey
	} else if current.job != "" {
		name = current.job
	} else if current.as == "" || current.want == 0 {
		return "", fmt.Errorf("request step %d must set as and want", position)
	}

	if current.job == "run_branch_worker" {
		key, err := tenant.Parse(current.forTenant)
		if err != nil || key == tenant.Production() {
			return "", fmt.Errorf("branch worker step %d requires a non-production tenant", position)
		}
	} else if current.forTenant != "" {
		return "", fmt.Errorf("step %d sets forTenant without a branch worker", position)
	}

	if !stepNamePattern.MatchString(name) {
		return "", fmt.Errorf("step %d name %q must match %s", position, name, stepNamePattern)
	}

	return fmt.Sprintf("%02d_%s", position, name), nil
}

func checkJourneyFiles(directory string, steps []step, names []string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	expected := make(map[string]bool, len(names))
	for index, name := range names {
		if steps[index].job == "" {
			expected[name] = true
		}
	}
	for _, entry := range entries {
		switch {
		case entry.IsDir() && expected[entry.Name()]:
		case !entry.IsDir() && (entry.Name() == "setup.sql" || entry.Name() == "relationships.json"):
		default:
			return fmt.Errorf("unknown entry %q in journey directory %s", entry.Name(), directory)
		}
	}

	for index, name := range names {
		if steps[index].job != "" {
			continue
		}
		want := []string{"golden.http", "request.http"}
		if steps[index].verify != "" {
			want = []string{"verify.json", "verify.sql"}
		} else if steps[index].valkey != "" {
			want = []string{"verify.json"}
		}
		if err := checkStepFiles(filepath.Join(directory, name), want); err != nil {
			return err
		}
	}
	return nil
}

func runJobStep(t *testing.T, fixture journeyFixture, current step) {
	t.Helper()
	switch current.job {
	case "run_worker", "run_branch_worker":
		fixture.workerStep(t, current.forTenant)
	default:
		t.Fatalf("unknown journey job %q", current.job)
	}
}

func checkStepFiles(directory string, want []string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		got = append(got, entry.Name())
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("step directory %s holds %v, want exactly %v", directory, got, want)
	}
	return nil
}

func loadCast(steps []step) (map[member]string, error) {
	contents, err := os.ReadFile(filepath.Join(journeysDir, "cast.json"))
	if err != nil {
		return nil, fmt.Errorf("read cast: %w", err)
	}
	var tokens map[member]string
	if err := json.Unmarshal(contents, &tokens); err != nil {
		return nil, fmt.Errorf("decode cast: %w", err)
	}
	if _, ok := tokens[none]; ok {
		return nil, fmt.Errorf("cast must not define %q", none)
	}

	for index, current := range steps {
		members := []member{current.as}
		for other := range current.others {
			members = append(members, other)
		}
		for _, who := range members {
			if who == "" || who == none {
				continue
			}
			if tokens[who] == "" {
				return nil, fmt.Errorf("step %d references cast member %q without a token", index+1, who)
			}
		}
	}
	return tokens, nil
}

func resetJourney(t *testing.T, s *suite, directory string) {
	t.Helper()

	ctx := t.Context()
	if err := s.db.Reset(
		ctx,
		filepath.Join(journeysDir, "setup.sql"),
		filepath.Join(directory, "setup.sql"),
		"testdata/accounts.sql",
	); err != nil {
		t.Fatal(err)
	}
	err := resetRelationships(
		ctx,
		s.db.Pool,
		filepath.Join(journeysDir, "relationships.json"),
		filepath.Join(directory, "relationships.json"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if s.flipt != nil {
		s.flipt.Reset()
	}
	if leaderboardValkey != nil {
		if err := leaderboardValkey.reset(ctx); err != nil {
			t.Fatal(err)
		}
	}
	s.resetProfileService()
}

func runRequestStep(
	t *testing.T,
	fixture journeyFixture,
	handler http.Handler,
	directory string,
	current step,
	tokens map[member]string,
) {
	t.Helper()

	if readHTTPRequest(t, directory).Header.Get("Authorization") != "" {
		t.Fatalf("%s/request.http carries Authorization; the journey table names the cast member instead", directory)
	}

	others := make([]member, 0, len(current.others))
	for other := range current.others {
		others = append(others, other)
	}
	slices.Sort(others)
	for _, other := range others {
		recorder := fixture.serveFixture(t, handler, readHTTPRequest(t, directory), tokens[other])
		if recorder.Code != current.others[other] {
			t.Errorf("%s: status=%d, want %d", other, recorder.Code, current.others[other])
		}
	}
	if t.Failed() {
		return
	}

	request := readHTTPRequest(t, directory)
	defer request.Body.Close()
	recorder := fixture.serveFixture(t, handler, request, tokens[current.as])
	checkRecordedHTTPGolden(t, request, recorder, directory, current.want, *updateGoldens, fixture.normalize)
}

func authorize(request *http.Request, who member, tokens map[member]string) {
	if who == none {
		return
	}
	request.Header.Set("Authorization", "Bearer "+tokens[who])
}

func checkVerifyGolden(t *testing.T, s *suite, directory string) {
	t.Helper()
	fixture := suiteJourney{suite: s}
	checkJSONGolden(t, directory, []byte(fixture.verifyQuery(t, readVerifyQuery(t, directory))))
}

func readVerifyQuery(t testing.TB, directory string) string {
	t.Helper()
	query, err := os.ReadFile(filepath.Join(directory, "verify.sql"))
	if err != nil {
		t.Fatalf("read verify query: %v", err)
	}
	return string(query)
}

func checkValkeyGolden(t *testing.T, fixture journeyFixture, directory string) {
	t.Helper()
	keys := fixture.cacheKeys(t)
	for index, key := range keys {
		keys[index] = fixture.normalize(key)
	}
	slices.Sort(keys)
	encoded, err := json.Marshal(slices.Compact(keys))
	if err != nil {
		t.Fatal(err)
	}
	checkJSONGolden(t, directory, encoded)
}

func checkJSONGolden(t *testing.T, directory string, value []byte) {
	t.Helper()
	var got bytes.Buffer
	if err := json.Indent(&got, value, "", "  "); err != nil {
		t.Fatalf("indent verify result: %v", err)
	}
	got.WriteByte('\n')

	goldenPath, err := goldenFilePath(directory, "verify.json", *updateGoldens)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := reconcileGolden(goldenPath, got.String(), *updateGoldens)
	if err != nil {
		t.Error(err)
		return
	}
	if updated {
		fmt.Fprintf(os.Stderr, "rewrote verify golden %s\n", goldenPath)
	}
}

func wrapVerifyQuery(query string) string {
	query = strings.TrimSuffix(strings.TrimSpace(query), ";")
	return "select coalesce(json_agg(q), '[]'::json)::text from (" + query + ") as q"
}

func TestStepDirNames(t *testing.T) {
	tests := []struct {
		name  string
		steps []step
		want  []string
	}{
		{
			name: "request then verify",
			steps: []step{
				{request: "create", as: admin, want: http.StatusCreated},
				{verify: "stored"},
			},
			want: []string{"01_create", "02_stored"},
		},
		{name: "no steps"},
		{name: "both kinds", steps: []step{{request: "a", verify: "b", as: admin, want: http.StatusOK}}},
		{name: "neither kind", steps: []step{{as: admin, want: http.StatusOK}}},
		{name: "request without cast member", steps: []step{{request: "a", want: http.StatusOK}}},
		{name: "request without status", steps: []step{{request: "a", as: guest}}},
		{name: "verify with others", steps: []step{{verify: "a", others: cast{guest: http.StatusOK}}}},
		{name: "uppercase name", steps: []step{{request: "Create", as: admin, want: http.StatusCreated}}},
		{
			name:  "cache snapshot",
			steps: []step{{valkey: "separate_keys"}},
			want:  []string{"01_separate_keys"},
		},
		{
			name:  "scoped worker",
			steps: []step{{job: "run_branch_worker", forTenant: "e2e/alpha-0000000a"}},
			want:  []string{"01_run_branch_worker"},
		},
		{name: "branch worker missing tenant", steps: []step{{job: "run_branch_worker"}}},
		{
			name:  "branch worker invalid tenant",
			steps: []step{{job: "run_branch_worker", forTenant: "invalid"}},
		},
		{
			name:  "branch worker canonical tenant",
			steps: []step{{job: "run_branch_worker", forTenant: "tadoku/prod"}},
		},
		{
			name:  "base worker tenant",
			steps: []step{{job: "run_worker", forTenant: "e2e/alpha-0000000a"}},
		},
		{
			name:  "request with worker tenant",
			steps: []step{{request: "read", as: admin, want: http.StatusOK, forTenant: "e2e/alpha-0000000a"}},
		},
		{name: "cache snapshot with request", steps: []step{{valkey: "keys", request: "read"}}},
		{name: "cache snapshot with actor", steps: []step{{valkey: "keys", as: admin}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := stepDirNames(test.steps)
			if test.want == nil {
				if err == nil {
					t.Fatalf("names=%v, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, test.want) {
				t.Errorf("names=%v, want %v", got, test.want)
			}
		})
	}
}

func TestCheckJourneyFiles(t *testing.T) {
	steps := []step{
		{request: "create", as: admin, want: http.StatusCreated},
		{verify: "stored"},
	}
	names := []string{"01_create", "02_stored"}
	write := func(t *testing.T, path string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	remove := func(t *testing.T, path string) {
		t.Helper()
		if err := os.RemoveAll(path); err != nil {
			t.Fatal(err)
		}
	}
	journey := func(t *testing.T) string {
		directory := t.TempDir()
		write(t, filepath.Join(directory, "setup.sql"))
		write(t, filepath.Join(directory, "01_create", "request.http"))
		write(t, filepath.Join(directory, "01_create", "golden.http"))
		write(t, filepath.Join(directory, "02_stored", "verify.sql"))
		write(t, filepath.Join(directory, "02_stored", "verify.json"))
		return directory
	}

	t.Run("complete journey", func(t *testing.T) {
		if err := checkJourneyFiles(journey(t), steps, names); err != nil {
			t.Fatal(err)
		}
	})

	corruptions := []struct {
		name    string
		corrupt func(t *testing.T, directory string)
	}{
		{name: "unknown file", corrupt: func(t *testing.T, directory string) { write(t, filepath.Join(directory, "notes.md")) }},
		{name: "unknown directory", corrupt: func(t *testing.T, directory string) { write(t, filepath.Join(directory, "03_extra", "request.http")) }},
		{name: "extra step file", corrupt: func(t *testing.T, directory string) { write(t, filepath.Join(directory, "01_create", "setup.sql")) }},
		{name: "missing step file", corrupt: func(t *testing.T, directory string) { remove(t, filepath.Join(directory, "02_stored", "verify.json")) }},
		{name: "missing step", corrupt: func(t *testing.T, directory string) { remove(t, filepath.Join(directory, "01_create")) }},
	}
	for _, test := range corruptions {
		t.Run(test.name, func(t *testing.T) {
			directory := journey(t)
			test.corrupt(t, directory)
			if err := checkJourneyFiles(directory, steps, names); err == nil {
				t.Fatal("corrupt journey directory accepted")
			}
		})
	}
}

func TestWrapVerifyQuery(t *testing.T) {
	got := wrapVerifyQuery("select 1 as one;\n")
	want := "select coalesce(json_agg(q), '[]'::json)::text from (select 1 as one) as q"
	if got != want {
		t.Errorf("query=%q, want %q", got, want)
	}
}
