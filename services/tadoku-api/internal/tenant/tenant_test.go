package tenant_test

import (
	"context"
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestParse(t *testing.T) {
	for _, raw := range []string{
		"tadoku/prod",
		"tadoku/dev",
		"e2e/run-1a2b3c4d",
		"a/0",
		strings.Repeat("a", 56) + "/" + strings.Repeat("b", 56),
	} {
		t.Run(raw, func(t *testing.T) {
			key, err := tenant.Parse(raw)
			if err != nil || key.String() != raw {
				t.Fatalf("Parse(%q): key=%q error=%v", raw, key.String(), err)
			}
		})
	}
	for _, raw := range []string{
		"",
		"tadoku",
		"TADOKU/prod",
		"tadoku/PROD",
		"-x/id",
		"x/-id",
		"a_b/id",
		"x/a_b",
		"tadoku/prod ",
		"e2e/",
		"/run",
		"e2e/run/id",
		strings.Repeat("a", 57) + "/id",
		"x/" + strings.Repeat("a", 57),
		"é/id",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := tenant.Parse(raw); err == nil {
				t.Fatalf("accepted invalid key %q", raw)
			}
		})
	}
}

func TestDeployment(t *testing.T) {
	other, err := tenant.Parse("e2e/other")
	if err != nil {
		t.Fatal(err)
	}
	base, err := tenant.ParseDeployment("")
	if err != nil || !base.Serves(tenant.Production()) || !base.Serves(other) {
		t.Fatalf("base deployment does not serve valid tenants: %v", err)
	}
	branch, err := tenant.ParseDeployment("e2e/other")
	if err != nil || !branch.Serves(other) || branch.Serves(tenant.Production()) {
		t.Fatalf("branch deployment does not isolate its tenant: %v", err)
	}
	for _, raw := range []string{"tadoku/prod", "tadoku", "e2e/", "invalid_key/id"} {
		if _, err := tenant.ParseDeployment(raw); err == nil {
			t.Errorf("accepted branch deployment %q", raw)
		}
	}
	if base.Serves(tenant.Key{}) || branch.Serves(tenant.Key{}) {
		t.Error("deployment served a zero key")
	}
}

func TestContext(t *testing.T) {
	if _, ok := tenant.FromContext(context.Background()); ok {
		t.Fatal("bare context has a tenant")
	}
	if _, ok := tenant.FromContext(tenant.WithKey(context.Background(), tenant.Key{})); ok {
		t.Fatal("zero key became a valid context tenant")
	}
	key, err := tenant.Parse("e2e/context")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(tenant.WithKey(context.Background(), key))
	cancel()
	got, ok := tenant.FromContext(ctx)
	if !ok || got != key || ctx.Err() != context.Canceled {
		t.Fatalf("tenant context: key=%q ok=%v error=%v", got.String(), ok, ctx.Err())
	}
}
