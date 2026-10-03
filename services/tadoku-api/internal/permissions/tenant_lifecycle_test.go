package permissions

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	ketoapi "github.com/ory/keto-client-go"
	ketoclient "github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
)

func TestTenantManagerLifecycle(t *testing.T) {
	fixture, err := testketo.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.Close(); err != nil {
			t.Error(err)
		}
	})
	client := ketoclient.NewClient(fixture.ReadURL(), fixture.WriteURL())
	manager := NewTenantManager(client)
	key := lifecycleTestKey(t, "e2e/alpha-0000000a")
	other := lifecycleTestKey(t, "e2e/bravo-0000000b")
	admin, banned, tester, addedTester, otherTester := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for relation, subject := range map[string]uuid.UUID{"admins": admin, "banned": banned} {
		err := client.AddRelation(t.Context(), "app", "tadoku", relation, ketoclient.Subject{ID: subject.String()})
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Provision(t.Context(), other, []uuid.UUID{otherTester}); err != nil {
		t.Fatal(err)
	}
	canonicalBefore := lifecycleRelationships(t, fixture.ReadURL(), "tadoku")
	otherBefore := lifecycleRelationships(t, fixture.ReadURL(), other.String())

	for range 2 {
		if err := manager.Provision(t.Context(), key, []uuid.UUID{tester, tester}); err != nil {
			t.Fatal(err)
		}
	}
	provisioned := lifecycleRelationships(t, fixture.ReadURL(), key.String())
	if got := len(provisioned); got != 2 {
		t.Fatalf("provision twice produced %d tuples, want one parent and one tester: %v", got, provisioned)
	}
	for _, test := range []struct {
		name     string
		subject  uuid.UUID
		relation string
		allowed  bool
	}{
		{"inherited administrator", admin, "admin", true},
		{"inherited administrator access", admin, "access", true},
		{"inherited ban", banned, "is_banned", true},
		{"tester access", tester, "access", true},
		{"tester has no administrator grant", tester, "admin", false},
		{"other tenant tester has no access", otherTester, "access", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			allowed, err := client.CheckPermission(
				t.Context(), "app", key.String(), test.relation, ketoclient.Subject{ID: test.subject.String()},
			)
			if err != nil || allowed != test.allowed {
				t.Errorf("allowed=%t error=%v, want %t", allowed, err, test.allowed)
			}
		})
	}

	for _, relation := range []string{"admins", "banned"} {
		err := client.AddRelation(t.Context(), "app", key.String(), relation, ketoclient.Subject{ID: tester.String()})
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := manager.Provision(t.Context(), key, []uuid.UUID{tester, addedTester}); err != nil {
		t.Fatal(err)
	}
	if got := len(lifecycleRelationships(t, fixture.ReadURL(), key.String())); got != 5 {
		t.Fatalf("rerun produced %d tuples, want existing grants plus the added tester", got)
	}

	for range 2 {
		if err := manager.Delete(t.Context(), key); err != nil {
			t.Fatal(err)
		}
		if got := len(lifecycleRelationships(t, fixture.ReadURL(), key.String())); got != 0 {
			t.Errorf("delete left %d tenant tuples", got)
		}
	}
	if after := lifecycleRelationships(t, fixture.ReadURL(), "tadoku"); !reflect.DeepEqual(after, canonicalBefore) {
		t.Errorf("canonical tuples changed: before=%v after=%v", canonicalBefore, after)
	}
	if after := lifecycleRelationships(t, fixture.ReadURL(), other.String()); !reflect.DeepEqual(after, otherBefore) {
		t.Errorf("other tenant tuples changed: before=%v after=%v", otherBefore, after)
	}
	allowed, err := client.CheckPermission(
		t.Context(), "app", other.String(), "access", ketoclient.Subject{ID: otherTester.String()},
	)
	if err != nil || !allowed {
		t.Errorf("other tenant lost access: allowed=%t error=%v", allowed, err)
	}
	t.Log("real Keto lifecycle: inherited grants, idempotent provision, scoped deletion and preserved neighbours")
}

func TestTenantManagerRejectsZeroBeforeProvider(t *testing.T) {
	var requests atomic.Int32
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(provider.Close)
	manager := NewTenantManager(ketoclient.NewClient(provider.URL, provider.URL))

	if err := manager.Provision(t.Context(), tenant.TestKey{}, []uuid.UUID{uuid.New()}); err == nil {
		t.Error("zero test key accepted for provision")
	}
	if err := manager.Delete(t.Context(), tenant.TestKey{}); err == nil {
		t.Error("zero test key accepted for deletion")
	}
	if got := requests.Load(); got != 0 {
		t.Errorf("invalid key sent %d provider requests", got)
	}
}

func TestTenantManagerPreservesProviderFailure(t *testing.T) {
	for _, operation := range []string{"provision", "delete"} {
		for _, status := range []int{http.StatusTooManyRequests, http.StatusServiceUnavailable} {
			t.Run(operation+"/"+http.StatusText(status), func(t *testing.T) {
				body := fmt.Sprintf(`{"error":{"code":%d,"message":"provider detail"}}`, status)
				provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					_, _ = io.Copy(io.Discard, r.Body)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write([]byte(body))
				}))
				t.Cleanup(provider.Close)
				manager := NewTenantManager(ketoclient.NewClient(provider.URL, provider.URL))
				key := lifecycleTestKey(t, "e2e/failure-0000000f")

				var err error
				if operation == "provision" {
					err = manager.Provision(t.Context(), key, []uuid.UUID{uuid.New()})
				} else {
					err = manager.Delete(t.Context(), key)
				}
				var providerErr *ketoapi.GenericOpenAPIError
				if !errors.As(err, &providerErr) {
					t.Fatalf("error=%v, want SDK provider error", err)
				}
				if string(providerErr.Body()) != body {
					t.Errorf("provider error body=%s, want %s", providerErr.Body(), body)
				}
			})
		}
	}
}

func lifecycleTestKey(t *testing.T, raw string) tenant.TestKey {
	t.Helper()
	key, err := tenant.ParseTestTenant(raw)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func lifecycleRelationships(t *testing.T, readURL, object string) []string {
	t.Helper()
	query := url.Values{"namespace": {"app"}, "object": {object}, "page_size": {"500"}}
	endpoint := readURL + "/relation-tuples?" + query.Encode()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := (&http.Client{Timeout: 2 * time.Second}).Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("read relationships: status %d", response.StatusCode)
	}

	var result struct {
		Tuples    []json.RawMessage `json:"relation_tuples"`
		NextToken string            `json:"next_page_token"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.NextToken != "" {
		t.Fatal("fixture relationships exceed one page")
	}
	tuples := make([]string, len(result.Tuples))
	for index, tuple := range result.Tuples {
		tuples[index] = string(tuple)
	}
	sort.Strings(tuples)
	return tuples
}
