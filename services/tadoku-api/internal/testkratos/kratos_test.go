package testkratos

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	kratosapi "github.com/ory/kratos-client-go"
)

const readerID = "11111111-1111-4111-8111-111111111111"

func TestFixedIdentitiesAndFullReset(t *testing.T) {
	t.Setenv("DSN", "postgres://127.0.0.1:1/unsafe")
	t.Setenv("SERVE_ADMIN_HOST", "127.0.0.1")
	t.Setenv("SERVE_ADMIN_PORT", "1")
	fixture, err := New(t.Context(), "testdata/identities.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})
	client := fixture.Client()
	initial, response, err := client.IdentityApi.GetIdentity(t.Context(), readerID).Execute()
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || initial.Id != readerID || initial.GetState() != "active" {
		t.Fatalf("seeded identity = %+v, status = %d", initial, response.StatusCode)
	}
	wantTime := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	if !initial.GetCreatedAt().Equal(wantTime) || !initial.GetUpdatedAt().Equal(wantTime) {
		t.Errorf("seed timestamps = %v, %v", initial.CreatedAt, initial.UpdatedAt)
	}
	wantTraits := map[string]any{"email": "reader@example.test", "display_name": "Reader One"}
	if !reflect.DeepEqual(initial.Traits, wantTraits) {
		t.Errorf("traits = %#v", initial.Traits)
	}

	body := kratosapi.UpdateIdentityBody{
		SchemaId: "user",
		State:    "inactive",
		Traits:   map[string]any{"email": "changed@example.test", "display_name": "Changed"},
	}
	if _, _, err := client.IdentityApi.UpdateIdentity(t.Context(), readerID).UpdateIdentityBody(body).Execute(); err != nil {
		t.Fatal(err)
	}
	extra, _, err := client.IdentityApi.CreateIdentity(t.Context()).CreateIdentityBody(kratosapi.CreateIdentityBody{
		SchemaId: "user",
		Traits:   map[string]any{"email": "extra@example.test", "display_name": "Extra"},
	}).Execute()
	if err != nil {
		t.Fatal(err)
	}

	firstProcess := fixture.command.Process
	if err := fixture.Reset(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := firstProcess.Signal(syscall.Signal(0)); err == nil {
		t.Error("reset left the old process alive")
	}
	restored, _, err := client.IdentityApi.GetIdentity(t.Context(), readerID).Execute()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, initial) {
		t.Errorf("reset identity = %+v, want %+v", restored, initial)
	}
	if _, response, err := client.IdentityApi.GetIdentity(t.Context(), extra.Id).Execute(); err == nil || response == nil || response.StatusCode != http.StatusNotFound {
		t.Fatalf("extra identity survived reset: response = %v, error = %v", response, err)
	}

	if _, err := client.IdentityApi.DeleteIdentity(t.Context(), readerID).Execute(); err != nil {
		t.Fatal(err)
	}
	if err := fixture.Reset(t.Context()); err != nil {
		t.Fatal(err)
	}
	if _, _, err := client.IdentityApi.GetIdentity(t.Context(), readerID).Execute(); err != nil {
		t.Fatalf("reset did not restore deleted identity: %v", err)
	}
	lastProcess := fixture.command.Process
	if err := fixture.Close(); err != nil {
		t.Fatal(err)
	}
	if err := lastProcess.Signal(syscall.Signal(0)); err == nil {
		t.Error("close left the process alive")
	}
	if _, err := os.Stat(fixture.dir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("fixture directory remains: %v", err)
	}
	if err := fixture.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

func TestFixturesAreIsolated(t *testing.T) {
	first, err := New(t.Context(), "testdata/identities.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := first.Close(); err != nil {
			t.Error(err)
		}
	})
	second, err := New(t.Context(), "testdata/identities.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := second.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := first.Client().IdentityApi.DeleteIdentity(t.Context(), readerID).Execute(); err != nil {
		t.Fatal(err)
	}
	if _, _, err := second.Client().IdentityApi.GetIdentity(t.Context(), readerID).Execute(); err != nil {
		t.Fatalf("one fixture changed another: %v", err)
	}
}

func TestFailedStartupCleansOwnedDirectory(t *testing.T) {
	badSeed := filepath.Join(t.TempDir(), "bad.sql")
	if err := os.WriteFile(badSeed, []byte("insert into absent_table values (1);"), 0o600); err != nil {
		t.Fatal(err)
	}
	fixture, err := New(t.Context(), badSeed)
	if fixture != nil || err == nil || !strings.Contains(err.Error(), "seed Kratos identities") {
		if fixture != nil {
			_ = fixture.Close()
		}
		t.Fatalf("bad seed: fixture = %v, error = %v", fixture, err)
	}
	matches, err := filepath.Glob(fmt.Sprintf("/dev/shm/tadoku-testkratos-%d-*", os.Getpid()))
	if err != nil || len(matches) != 0 {
		t.Errorf("failed constructor left directories: %v, %v", matches, err)
	}
}

func TestCancellationAndUnusableReset(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if fixture, err := New(ctx, "testdata/identities.sql"); fixture != nil || !errors.Is(err, context.Canceled) {
		if fixture != nil {
			_ = fixture.Close()
		}
		t.Fatalf("cancelled startup: fixture = %v, error = %v", fixture, err)
	}
	fixture, err := New(t.Context(), "testdata/identities.sql")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := fixture.Reset(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("cancelled reset = %v", err)
	}
	if !errors.Is(fixture.Err(), context.Canceled) {
		t.Errorf("failed reset did not poison fixture: %v", fixture.Err())
	}
}

func TestUnexpectedExit(t *testing.T) {
	fixture, err := New(t.Context(), "testdata/identities.sql")
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.command.Process.Kill(); err != nil {
		_ = fixture.Close()
		t.Fatal(err)
	}
	<-fixture.wait
	if err := fixture.Err(); err == nil {
		t.Error("unexpected exit was not detected")
	}
	if err := fixture.Close(); err == nil {
		t.Error("Close did not report the unexpected exit")
	}
	if _, err := os.Stat(fixture.dir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("directory remains after unexpected exit: %v", err)
	}
}
