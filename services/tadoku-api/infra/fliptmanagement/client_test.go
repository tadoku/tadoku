package fliptmanagement

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestNewClient(t *testing.T) {
	t.Parallel()
	targets, err := flipt.NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	httpClient := &http.Client{Timeout: time.Second}

	if _, err := NewClient(Config{URL: "relative/path", Targets: targets, HTTPClient: httpClient}); err == nil {
		t.Fatal("relative URL accepted")
	}
	if _, err := NewClient(Config{URL: "http://flipt.example", Targets: targets}); err == nil {
		t.Fatal("nil HTTP client accepted")
	}
	client, err := NewClient(Config{URL: "http://flipt.example", Targets: targets, HTTPClient: httpClient})
	if err != nil || client == nil {
		t.Fatalf("valid config rejected: client=%v error=%v", client, err)
	}
}

func TestManagementRejectsAnotherNamespaceAndMissingTenant(t *testing.T) {
	targets, err := flipt.NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	key, err := tenant.Parse("e2e/flipt-0123abcd")
	if err != nil {
		t.Fatal(err)
	}
	spec := Segment{
		Key:         "fixture-flag-access",
		Name:        "Fixture flag access",
		Description: "Synthetic fixture account IDs.",
	}
	user := uuid.MustParse("11111111-1111-4111-8111-111111111111")

	for _, namespace := range []string{"default", "e2e_flipt-0123abcd"} {
		t.Run(namespace, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				calls++
				if request.URL.Path != "/api/v2/environments/test/namespaces/e2e_flipt-0123abcd/resources/"+
					"flipt.core.Segment/fixture-flag-access" {
					t.Errorf("management target=%q", request.URL.Path)
				}
				response.Header().Set("Content-Type", "application/json")
				body := `{"resource":{"namespaceKey":"NAMESPACE","key":"fixture-flag-access",` +
					`"payload":{"@type":"flipt.core.Segment","key":"fixture-flag-access",` +
					`"name":"Fixture flag access",` +
					`"description":"Synthetic fixture account IDs.",` +
					`"matchType":"ALL_MATCH_TYPE","constraints":[{"type":"ENTITY_ID_COMPARISON_TYPE",` +
					`"property":"entityId","operator":"isoneof","value":"[]","description":""}]}},` +
					`"revision":"0000000000000000000000000000000000000001"}`
				_, _ = response.Write([]byte(strings.Replace(body, "NAMESPACE", namespace, 1)))
			}))
			t.Cleanup(server.Close)
			client, err := NewClient(Config{
				URL: server.URL, Targets: targets, HTTPClient: &http.Client{Timeout: time.Second},
			})
			if err != nil {
				t.Fatal(err)
			}

			if _, err := client.SetNamedUserAccess(context.Background(), spec, user, true); err == nil || calls != 0 {
				t.Fatalf("missing tenant error=%v provider calls=%d", err, calls)
			}
			state, err := client.GetNamedUserAccess(tenant.WithKey(t.Context(), key), spec, user)
			if namespace == "default" {
				if !errors.Is(err, ErrUnavailable) {
					t.Fatalf("cross-namespace provider response accepted: state=%+v error=%v", state, err)
				}
			} else if err != nil || state.Environment != "test" || state.Enabled {
				t.Fatalf("resolved management state=%+v error=%v", state, err)
			}
		})
	}
}
