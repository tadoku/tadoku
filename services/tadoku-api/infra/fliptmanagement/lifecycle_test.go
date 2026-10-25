package fliptmanagement_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testflipt"
)

func TestTestNamespaceLifecyclePreservesExistingResources(t *testing.T) {
	fixture := testflipt.New()
	t.Cleanup(fixture.Close)
	targets, err := flipt.NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	client := fliptmanagement.NewClient(fliptmanagement.Config{URL: fixture.URL(), Targets: targets})
	key, err := tenant.ParseTestTenant("e2e/lifecycle-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	resources, err := fliptmanagement.ParseFeatures(strings.NewReader(lifecycleSeed))
	if err != nil {
		t.Fatal(err)
	}

	if err := client.ProvisionTestNamespace(t.Context(), key, resources); err != nil {
		t.Fatal(err)
	}
	user := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	spec := fliptmanagement.Segment{
		Key:         "release-log-entry-v2-access",
		Name:        "Release log entry v2 access",
		Description: "Kratos UUIDs explicitly granted access to release log entry v2.",
	}
	_, err = client.SetNamedUserAccess(tenant.WithKey(t.Context(), key.Key()), spec, user, true)
	if err != nil {
		t.Fatal(err)
	}
	resources, err = fliptmanagement.ParseFeatures(strings.NewReader(lifecycleSeed + `
flags:
- key: added-flag
  type: BOOLEAN_FLAG_TYPE
  enabled: false
`))
	if err != nil {
		t.Fatal(err)
	}
	if err := client.ProvisionTestNamespace(t.Context(), key, resources); err != nil {
		t.Fatal(err)
	}
	state, err := client.GetNamedUserAccess(tenant.WithKey(t.Context(), key.Key()), spec, user)
	if err != nil || !state.Enabled {
		t.Fatalf("existing grant overwritten: state=%+v error=%v", state, err)
	}
	counts := fixture.CreateCounts(key.Key())
	if counts.Namespace != 1 || counts.Resources != 2 || fixture.ResourceCount(key.Key()) != 2 {
		t.Fatalf("idempotent creates: %+v resources=%d", counts, fixture.ResourceCount(key.Key()))
	}
	canonical := tenant.Production()
	if !fixture.Exists(canonical) {
		t.Fatal("canonical namespace removed during provisioning")
	}

	fixture.SetUnavailable(true)
	if err := client.DeleteTestNamespace(t.Context(), key); !errors.Is(err, fliptmanagement.ErrUnavailable) {
		t.Fatalf("provider failure not propagated: %v", err)
	}
	if !fixture.Exists(key.Key()) {
		t.Fatal("failed delete removed fixture state")
	}
	fixture.SetUnavailable(false)
	if err := client.DeleteTestNamespace(t.Context(), key); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteTestNamespace(t.Context(), key); err != nil {
		t.Fatal(err)
	}
	if fixture.Exists(key.Key()) || !fixture.Exists(canonical) {
		t.Fatal("teardown did not isolate its namespace")
	}
}

func TestTestNamespaceLifecycleRejectsZeroKeyWithoutRequests(t *testing.T) {
	fixture := testflipt.New()
	t.Cleanup(fixture.Close)
	targets, _ := flipt.NewTargets("local", "default", "test")
	client := fliptmanagement.NewClient(fliptmanagement.Config{URL: fixture.URL(), Targets: targets})

	if err := client.ProvisionTestNamespace(t.Context(), tenant.TestKey{}, nil); err == nil {
		t.Fatal("zero test key provisioned")
	}
	if err := client.DeleteTestNamespace(t.Context(), tenant.TestKey{}); err == nil {
		t.Fatal("zero test key deleted")
	}
	if fixture.RequestCount() != 0 {
		t.Fatalf("invalid key reached provider %d times", fixture.RequestCount())
	}
}

const lifecycleSeed = `version: "1.6"
namespace: {key: default}
segments:
- key: release-log-entry-v2-access
  name: Release log entry v2 access
  description: Kratos UUIDs explicitly granted access to release log entry v2.
  match_type: ALL_MATCH_TYPE
  constraints:
  - type: ENTITY_ID_COMPARISON_TYPE
    property: entityId
    operator: isoneof
    value: []
`
