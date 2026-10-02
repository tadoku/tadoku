package e2e_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestFliptTenantIsolation(t *testing.T) {
	api.reset(t, "testdata/FliptTenantIsolation")
	if err := api.keto.Reset(t.Context(), "testdata/ImmersionFeatureAccessGrant/relationships.json"); err != nil {
		t.Fatal(err)
	}

	contents, err := os.ReadFile("testdata/FliptTenantIsolation/tokens.json")
	if err != nil {
		t.Fatal(err)
	}
	var tokens map[string]string
	if err := json.Unmarshal(contents, &tokens); err != nil {
		t.Fatal(err)
	}

	path := "/immersion/admin/feature-flags/release-log-entry-v2/users/11111111-1111-4111-8111-111111111111"
	for _, who := range []string{"test_user", "production_user"} {
		want := who == "production_user"
		assertFliptDecision(t, tokens[who], want)
	}

	body := isolationRequest(t, tokens["test_admin"], http.MethodPut, path, "", http.StatusOK)
	var state struct {
		Changed     bool   `json:"changed"`
		Enabled     bool   `json:"enabled"`
		Environment string `json:"environment"`
	}
	if err := json.Unmarshal(body, &state); err != nil || !state.Changed || !state.Enabled || state.Environment != "test" {
		t.Fatalf("test grant state=%+v error=%v", state, err)
	}
	assertFliptDecision(t, tokens["test_user"], true)

	isolationRequest(t, tokens["production_admin"], http.MethodDelete, path, "", http.StatusOK)
	assertFliptDecision(t, tokens["production_user"], false)
	assertFliptDecision(t, tokens["test_user"], true)

	isolationRequest(t, tokens["test_admin"], http.MethodDelete, path, "", http.StatusOK)
	assertFliptDecision(t, tokens["test_user"], false)
	assertFliptDecision(t, tokens["production_user"], false)
}

func assertFliptDecision(t *testing.T, token string, want bool) {
	t.Helper()
	body := isolationRequest(t, token, http.MethodGet, "/immersion/feature-flags", "", http.StatusOK)
	var response struct {
		Decisions map[string]bool `json:"decisions"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	got, ok := response.Decisions["release-log-entry-v2"]
	if !ok || got != want {
		t.Fatalf("release-log-entry-v2 decision=%t present=%t want=%t", got, ok, want)
	}
}
