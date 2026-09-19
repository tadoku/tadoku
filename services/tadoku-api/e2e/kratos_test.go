package e2e_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	kratosapi "github.com/ory/kratos-client-go"
)

// resetKratosAfter marks a test that may mutate the shared provider. Register it
// before mutation, on each implementation subtest or the enclosing journey test.
func resetKratosAfter(t *testing.T, s *suite) {
	t.Helper()
	if s.kratos == nil {
		t.Fatal("Kratos reset requested without a fixture")
	}
	if err := s.kratos.Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// testing cancels t.Context before running cleanup callbacks.
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		if err := s.kratos.Reset(ctx); err != nil {
			t.Errorf("restore shared Kratos seed: %v", err)
		}
	})
}

func TestKratosCastMatchesSignedSubjects(t *testing.T) {
	createdAt := map[string]time.Time{
		"11111111-1111-4111-8111-111111111111": time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		"22222222-2222-4222-8222-222222222222": fixtureInstant,
		"33333333-3333-4333-8333-333333333333": fixtureInstant,
		"44444444-4444-4444-8444-444444444444": time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC),
	}
	cast, err := loadCast(nil)
	if err != nil {
		t.Fatal(err)
	}
	for who, encoded := range cast {
		if who == guest {
			continue
		}
		// These checked-in tokens are verified by the real JWT middleware in
		// HTTP cases. Here we check that their subjects/traits match the provider.
		claims := jwt.MapClaims{}
		if _, _, err := new(jwt.Parser).ParseUnverified(encoded, claims); err != nil {
			t.Fatal(err)
		}
		subject, ok := claims["sub"].(string)
		if !ok {
			t.Fatalf("%s has no subject", who)
		}
		identity, _, err := api.kratos.Client().IdentityApi.GetIdentity(t.Context(), subject).Execute()
		if err != nil {
			t.Fatalf("%s: %v", who, err)
		}
		session := claims["session"].(map[string]any)
		traits := session["identity"].(map[string]any)["traits"]
		if !reflect.DeepEqual(identity.Traits, traits) {
			t.Errorf("%s provider traits = %#v, token traits = %#v", who, identity.Traits, traits)
		}
		if identity.Id != subject || identity.SchemaId != "user" || identity.GetState() != "active" || !identity.GetCreatedAt().Equal(createdAt[subject]) {
			t.Errorf("%s provider identity = %+v", who, identity)
		}
	}
}

func TestKratosResetIsExplicit(t *testing.T) {
	var extraID string
	t.Run("marked", func(t *testing.T) {
		resetKratosAfter(t, api)
		extra, _, err := api.kratos.Client().IdentityApi.CreateIdentity(t.Context()).CreateIdentityBody(kratosapi.CreateIdentityBody{
			SchemaId: "user",
			Traits:   map[string]any{"email": "cleanup@example.test", "display_name": "Cleanup"},
		}).Execute()
		if err != nil {
			t.Fatal(err)
		}
		extraID = extra.Id

		api.reset(t, "")
		if _, _, err := api.kratos.Client().IdentityApi.GetIdentity(t.Context(), extraID).Execute(); err != nil {
			t.Fatalf("ordinary scenario reset changed Kratos: %v", err)
		}
	})
	if extraID == "" {
		t.Fatal("marked test did not create its identity")
	}
	if _, response, err := api.kratos.Client().IdentityApi.GetIdentity(t.Context(), extraID).Execute(); err == nil || response == nil || response.StatusCode != http.StatusNotFound {
		t.Fatalf("marked test did not clean up: response = %v, error = %v", response, err)
	}
}

// Use a child test process to exercise t.Fatal without failing this test run.
// Its post-subtest read verifies cleanup ran after t.Context was cancelled.
func TestKratosCleanupAfterFailure(t *testing.T) {
	const childEnv = "TADOKU_KRATOS_CLEANUP_FAILURE_TEST"
	if os.Getenv(childEnv) == "1" {
		var extraID string
		t.Run("intentional_failure", func(t *testing.T) {
			resetKratosAfter(t, api)
			extra, _, err := api.kratos.Client().IdentityApi.CreateIdentity(t.Context()).CreateIdentityBody(kratosapi.CreateIdentityBody{
				SchemaId: "user",
				Traits:   map[string]any{"email": "failure@example.test", "display_name": "Failure"},
			}).Execute()
			if err != nil {
				t.Fatal(err)
			}
			extraID = extra.Id
			t.Fatal("intentional failure after mutation")
		})
		if extraID == "" || api.kratos.Err() != nil {
			t.Fatal("cleanup did not leave a usable fixture")
		}
		if _, response, err := api.kratos.Client().IdentityApi.GetIdentity(t.Context(), extraID).Execute(); err == nil || response == nil || response.StatusCode != http.StatusNotFound {
			t.Fatal("failed test leaked its identity")
		}
		fmt.Println("KRATOS_FAILURE_CLEANUP_VERIFIED")
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestKratosCleanupAfterFailure$", "-test.v")
	command.Env = append(os.Environ(), childEnv+"=1")
	output, err := command.CombinedOutput()
	if err == nil || ctx.Err() != nil || !strings.Contains(string(output), "KRATOS_FAILURE_CLEANUP_VERIFIED") {
		t.Fatalf("failure cleanup probe: %v\n%s", err, output)
	}
}
