package e2e_test

import (
	"bytes"
	"context"
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

type member string

const (
	none   member = "none"
	guest  member = "guest"
	user   member = "user"
	user2  member = "user2"
	admin  member = "admin"
	banned member = "banned"
)

type cast map[member]int

type step struct {
	request string
	verify  string
	job     string

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
	runJourneyWithSetup(t, s, handler, name, nil, steps)
}

func runJourneyWithSetup(t *testing.T, s *suite, handler http.Handler, name string, afterReset func(*testing.T), steps []step) {
	t.Helper()
	if s.kratos != nil {
		if err := s.kratos.Err(); err != nil {
			t.Fatal(err)
		}
	}

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
	if s.outbox != nil {
		workerContext, cancel := context.WithCancel(t.Context())
		s.outboxContext = workerContext
		defer func() {
			cancel()
			if s.outboxDone != nil {
				<-s.outboxDone
			}
		}()
	}
	if afterReset != nil {
		afterReset(t)
	}

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
				if current.job != "" {
					runJobStep(t, s, current.job)
					return
				}
				runRequestStep(t, s, handler, stepDir, current, tokens)
			})
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
	for _, value := range []string{current.request, current.verify, current.job} {
		if value != "" {
			kinds++
		}
	}
	if kinds != 1 {
		return "", fmt.Errorf("step %d must set exactly one request, verify or job", position)
	}

	name := current.request
	if current.verify != "" {
		name = current.verify
		if current.as != "" || current.want != 0 || current.others != nil {
			return "", fmt.Errorf("verify step %d must not set as, want or others", position)
		}
	} else if current.job != "" {
		name = current.job
		if current.as != "" || current.want != 0 || current.others != nil {
			return "", fmt.Errorf("job step %d must not set as, want or others", position)
		}
	} else if current.as == "" || current.want == 0 {
		return "", fmt.Errorf("request step %d must set as and want", position)
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
		}
		if err := checkStepFiles(filepath.Join(directory, name), want); err != nil {
			return err
		}
	}
	return nil
}

func runJobStep(t *testing.T, s *suite, job string) {
	t.Helper()
	switch job {
	case "run_leaderboard_outbox":
		if s.outbox == nil || s.outboxReady == nil || s.outboxContext == nil {
			t.Fatal("leaderboard outbox worker is not configured")
		}
		if s.outboxDone != nil {
			t.Fatal("leaderboard outbox worker already started")
		}
		done := make(chan struct{})
		s.outboxDone = done
		go func() {
			defer close(done)
			s.outbox.Run(s.outboxContext)
		}()
		select {
		case <-s.outboxReady:
		case <-time.After(3 * time.Second):
			t.Fatal("leaderboard outbox did not become ready")
		}
	default:
		t.Fatalf("unknown journey job %q", job)
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
	if err := s.db.Reset(ctx, filepath.Join(journeysDir, "setup.sql"), filepath.Join(directory, "setup.sql")); err != nil {
		t.Fatal(err)
	}
	if err := s.keto.Reset(ctx, filepath.Join(journeysDir, "relationships.json"), filepath.Join(directory, "relationships.json")); err != nil {
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
	s.resetProfileCaches()
}

func runRequestStep(t *testing.T, s *suite, handler http.Handler, directory string, current step, tokens map[member]string) {
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
		handler.ServeHTTP(recorder, replay)
		if recorder.Code != current.others[other] {
			t.Errorf("%s: status=%d, want %d", other, recorder.Code, current.others[other])
		}
	}
	if t.Failed() {
		return
	}

	request := readHTTPRequest(t, directory)
	authorize(request, current.as, tokens)
	checkHTTPResponseGolden(t, handler, request, directory, current.want, *updateGoldens)
}

func authorize(request *http.Request, who member, tokens map[member]string) {
	if who == none {
		return
	}
	request.Header.Set("Authorization", "Bearer "+tokens[who])
}

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
