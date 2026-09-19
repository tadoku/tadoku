package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

// member names a cast member in testdata/journeys/cast.json.
type member string

const (
	// none sends the request without credentials.
	none    member = "none"
	guest   member = "guest"
	reader  member = "reader"
	reader2 member = "reader2"
	admin   member = "admin"
	banned  member = "banned"
)

// cast maps cast members to the status each must receive when replaying a
// step's request. Replays run before the primary request and must not mutate
// state.
type cast map[member]int

// step is one entry of a journey. A request step sends request.http as the
// named cast member and compares the complete response with golden.http. A
// verify step runs verify.sql and compares its JSON result with verify.json.
type step struct {
	request string
	verify  string

	as     member
	want   int
	others cast

	// at is the business instant of the step; zero means fixtureInstant.
	at time.Time
}

const journeysDir = "testdata/journeys"

var stepNamePattern = regexp.MustCompile(`^[a-z0-9_]+$`)

// runJourney resets the stores once, then runs the steps in order against the
// production router, so state after the first step comes only from the API.
// It stops at the first failing step because later steps depend on it.
func runJourney(t *testing.T, s *suite, name string, steps []step) {
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

	resetJourney(t, s, directory)

	previous := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time { return fixtureInstant }
	defer func() { jwt.TimeFunc = previous }()

	for index, current := range steps {
		stepDir := filepath.Join(directory, names[index])
		at := current.at
		if at.IsZero() {
			at = fixtureInstant
		}

		passed := t.Run(names[index], func(t *testing.T) {
			timex.TheWorld(at, func() {
				if current.verify != "" {
					checkVerifyGolden(t, s, stepDir)
					return
				}
				runRequestStep(t, s, stepDir, current, tokens)
			})
		})
		if !passed {
			return
		}
	}

	if s.proxied.Load() != 0 {
		t.Error("handler contacted an upstream")
	}
}

// stepDirNames validates every step and returns its numbered directory name.
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
	switch {
	case current.request != "" && current.verify != "":
		return "", fmt.Errorf("step %d sets both request and verify", position)
	case current.request == "" && current.verify == "":
		return "", fmt.Errorf("step %d sets neither request nor verify", position)
	}

	name := current.request
	if current.verify != "" {
		name = current.verify
		if current.as != "" || current.want != 0 || current.others != nil {
			return "", fmt.Errorf("verify step %d must not set as, want or others", position)
		}
	} else if current.as == "" || current.want == 0 {
		return "", fmt.Errorf("request step %d must set as and want", position)
	}
	if !stepNamePattern.MatchString(name) {
		return "", fmt.Errorf("step %d name %q must match %s", position, name, stepNamePattern)
	}

	return fmt.Sprintf("%02d_%s", position, name), nil
}

// checkJourneyFiles rejects unknown entries so stale fixtures cannot linger,
// and requires every step directory to hold exactly its two files.
func checkJourneyFiles(directory string, steps []step, names []string) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	expected := make(map[string]bool, len(names))
	for _, name := range names {
		expected[name] = true
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
		want := []string{"golden.http", "request.http"}
		if steps[index].verify != "" {
			want = []string{"verify.json", "verify.sql"}
		}
		if err := checkStepFiles(filepath.Join(directory, name), want); err != nil {
			return err
		}
	}
	return nil
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

// loadCast reads the cast tokens and confirms every referenced member has one.
// none is implicit and never has a token.
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

// resetJourney clears the stores once, seeds the shared cast files, then the
// journey's own files. Steps never reset again.
func resetJourney(t *testing.T, s *suite, directory string) {
	t.Helper()

	ctx := t.Context()
	if err := s.db.Reset(ctx, filepath.Join(journeysDir, "setup.sql"), filepath.Join(directory, "setup.sql")); err != nil {
		t.Fatal(err)
	}
	if err := s.keto.Reset(ctx, filepath.Join(journeysDir, "relationships.json"), filepath.Join(directory, "relationships.json")); err != nil {
		t.Fatal(err)
	}
	s.proxied.Store(0)
}

// runRequestStep replays the request as every member in others, checking
// status only, then sends it as the step's member and compares the golden.
func runRequestStep(t *testing.T, s *suite, directory string, current step, tokens map[member]string) {
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
		replay := readHTTPRequest(t, directory)
		authorize(replay, other, tokens)

		recorder := httptest.NewRecorder()
		s.handler.ServeHTTP(recorder, replay)
		if recorder.Code != current.others[other] {
			t.Errorf("%s: status=%d, want %d", other, recorder.Code, current.others[other])
		}
	}
	if t.Failed() {
		return
	}

	request := readHTTPRequest(t, directory)
	authorize(request, current.as, tokens)
	checkHTTPResponseGolden(t, s.handler, request, directory, current.want, *updateGoldens)
}

func authorize(request *http.Request, who member, tokens map[member]string) {
	if who == none {
		return
	}
	request.Header.Set("Authorization", "Bearer "+tokens[who])
}

// checkVerifyGolden runs verify.sql, aggregates its rows into one JSON array
// in query order and compares the indented result with verify.json.
func checkVerifyGolden(t *testing.T, s *suite, directory string) {
	t.Helper()

	query, err := os.ReadFile(filepath.Join(directory, "verify.sql"))
	if err != nil {
		t.Fatalf("read verify query: %v", err)
	}

	var rows string
	if err := s.db.Pool.QueryRow(t.Context(), wrapVerifyQuery(string(query))).Scan(&rows); err != nil {
		t.Fatalf("run verify query: %v", err)
	}
	var got bytes.Buffer
	if err := json.Indent(&got, []byte(rows), "", "  "); err != nil {
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
