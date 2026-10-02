package e2e_test

import (
	"net/http"
	"path/filepath"
	"testing"
)

func TestFliptTenantIsolation(t *testing.T) {
	api.reset(t, "")
	if err := api.db.Reset(t.Context(), "testdata/FliptTenantIsolation/setup.sql"); err != nil {
		t.Fatal(err)
	}
	if err := api.keto.Reset(t.Context(), "testdata/FliptTenantIsolation/relationships.json"); err != nil {
		t.Fatal(err)
	}

	atFixtureInstant(func() {
		for _, name := range []string{
			"production_initially_enabled",
			"test_initially_disabled",
			"grant_test_user",
			"test_enabled",
			"revoke_production_user",
			"production_disabled",
			"test_still_enabled",
			"revoke_test_user",
			"test_disabled",
			"production_still_disabled",
		} {
			t.Run(name, func(t *testing.T) {
				directory := filepath.Join("testdata", "FliptTenantIsolation", name)
				checkHTTPGolden(t, api.handler, directory, http.StatusOK, *updateGoldens)
			})
		}
	})
}
