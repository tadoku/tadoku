package testketo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResetSeedsAndClearsEveryNamespace(t *testing.T) {
	t.Setenv("DSN", "postgres://127.0.0.1:1/unsafe")
	t.Setenv("SERVE_READ_PORT", "1")

	fixture, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	if err := fixture.Reset(t.Context(), "missing.json", filepath.Join("testdata", "relationships.json")); err != nil {
		t.Fatal(err)
	}
	configPath := fixture.command.Args[len(fixture.command.Args)-1]
	configFiles, err := os.ReadDir(filepath.Dir(configPath))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range configFiles {
		if file.Name() != filepath.Base(configPath) {
			t.Errorf("runtime file %s shares Keto's watched config directory", file.Name())
		}
	}

	for _, namespace := range []string{"app", "User"} {
		if count := relationshipCount(t, fixture, namespace); count != 1 {
			t.Errorf("%s relationship count=%d, want 1", namespace, count)
		}
	}
	if fixture.ReadURL() == "http://127.0.0.1:1" {
		t.Error("Keto inherited the parent read-port override")
	}
	additional := []byte(`{"namespace":"app","object":"tadoku","relation":"admins","subject_id":"written-after-seed"}`)
	if err := fixture.do(t.Context(), http.MethodPut, fixture.WriteURL()+"/admin/relation-tuples", additional); err != nil {
		t.Fatal(err)
	}
	if count := relationshipCount(t, fixture, "app"); count != 2 {
		t.Errorf("app relationship count after write=%d, want 2", count)
	}

	if err := fixture.Reset(t.Context()); err != nil {
		t.Fatal(err)
	}
	for _, namespace := range []string{"app", "User"} {
		if count := relationshipCount(t, fixture, namespace); count != 0 {
			t.Errorf("%s relationship count after reset=%d, want 0", namespace, count)
		}
	}
}

func TestFixturesAreIsolated(t *testing.T) {
	first, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := first.Close(); err != nil {
			t.Error(err)
		}
	})
	second, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := second.Close(); err != nil {
			t.Error(err)
		}
	})

	if err := first.Reset(t.Context(), filepath.Join("testdata", "relationships.json")); err != nil {
		t.Fatal(err)
	}
	if got := relationshipCount(t, second, "app"); got != 0 {
		t.Errorf("second fixture saw %d relationships from first", got)
	}
}

func TestResetRejectsBadSeeds(t *testing.T) {
	fixture, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})

	if err := fixture.Reset(t.Context(), filepath.Join("testdata", "bad.json")); err == nil || !strings.Contains(err.Error(), "decode Keto seed") {
		t.Errorf("bad JSON error=%v", err)
	}
	if err := fixture.Reset(t.Context(), filepath.Join("testdata", "relationships.json"), filepath.Join("testdata", "missing", "seed.json")); err != nil {
		t.Errorf("missing optional seed: %v", err)
	}
	if err := fixture.Reset(t.Context(), "testdata"); err == nil || !strings.Contains(err.Error(), "read Keto seed") {
		t.Errorf("unreadable seed error=%v", err)
	}
	if err := fixture.Reset(t.Context(), filepath.Join("testdata", "invalid.json")); err == nil || !strings.Contains(err.Error(), "seed Keto relationship") {
		t.Errorf("invalid tuple error=%v", err)
	}
}

func TestCloseReportsUnexpectedExit(t *testing.T) {
	fixture, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.command.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	<-fixture.wait
	if err := fixture.Close(); err == nil || !strings.Contains(err.Error(), "exited before close") {
		t.Errorf("unexpected exit error=%v", err)
	}
}

func relationshipCount(t *testing.T, fixture *Fixture, namespace string) int {
	t.Helper()
	endpoint := fixture.ReadURL() + "/relation-tuples?namespace=" + url.QueryEscape(namespace)
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{Timeout: requestTimeout}).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET %s status=%d", endpoint, response.StatusCode)
	}
	var relationships struct {
		RelationTuples []json.RawMessage `json:"relation_tuples"`
	}
	if err := json.NewDecoder(response.Body).Decode(&relationships); err != nil {
		t.Fatal(err)
	}
	return len(relationships.RelationTuples)
}

func TestCanceledStartup(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fixture, err := New(ctx)
	if fixture != nil || err == nil {
		if fixture != nil {
			_ = fixture.Close()
		}
		t.Fatalf("New with canceled context=(%v, %v)", fixture, err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("startup error=%v", err)
	}
}
