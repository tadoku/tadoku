package e2e_test

import (
	"net/http"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	fliptclient "github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestFliptTenantIsolation(t *testing.T) {
	api.reset(t, "")
	if err := api.db.Reset(t.Context(), "testdata/FliptTenantIsolation/setup.sql"); err != nil {
		t.Fatal(err)
	}
	if err := resetRelationships(
		t.Context(),
		api.db.Pool,
		"testdata/FliptTenantIsolation/relationships.json",
	); err != nil {
		t.Fatal(err)
	}

	targets, err := fliptclient.NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	management := fliptmanagement.NewClient(fliptmanagement.Config{URL: flipt.URL(), Targets: targets})
	testTenant, err := tenant.Parse("e2e/flipt-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	segment := fliptmanagement.Segment{
		Key: "release-webv3-access", Name: "Webv3 access", Description: "Synthetic fixture account IDs.",
	}
	subject := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	cases := []struct {
		name    string
		change  bool
		key     tenant.Key
		enabled bool
	}{
		{name: "production_initially_enabled"},
		{name: "test_initially_disabled"},
		{name: "test_enabled", change: true, key: testTenant, enabled: true},
		{name: "production_disabled", change: true, key: tenant.Production()},
		{name: "test_still_enabled"},
		{name: "test_disabled", change: true, key: testTenant},
		{name: "production_still_disabled"},
	}
	atFixtureInstant(func() {
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				if test.change {
					ctx := tenant.WithKey(t.Context(), test.key)
					if _, err := management.SetNamedUserAccess(ctx, segment, subject, test.enabled); err != nil {
						t.Fatal(err)
					}
				}
				directory := filepath.Join("testdata", "FliptTenantIsolation", test.name)
				checkHTTPGolden(t, api.handler, directory, http.StatusOK, *updateGoldens)
			})
		}
	})
}
