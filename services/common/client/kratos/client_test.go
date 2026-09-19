package kratos_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commonkratos "github.com/tadoku/tadoku/services/common/client/kratos"
)

func TestListAllIdentitiesBeyondOffsetLimit(t *testing.T) {
	const identityCount = 1205
	requests := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/admin/identities" {
			t.Errorf("path = %q, want /admin/identities", req.URL.Path)
		}
		if req.URL.Query().Get("page_size") != "500" {
			t.Errorf("page_size = %q, want 500", req.URL.Query().Get("page_size"))
		}
		if req.URL.Query().Has("page") || req.URL.Query().Has("per_page") {
			t.Errorf("request used unsupported offset pagination: %s", req.URL.RawQuery)
		}

		pageToken := req.URL.Query().Get("page_token")
		requests[pageToken]++
		var start, end int
		var next string
		switch pageToken {
		case "":
			start, end, next = 0, 500, "after-500"
		case "after-500":
			start, end, next = 500, 1000, "after-1000"
		case "after-1000":
			start, end = 1000, identityCount
		default:
			t.Errorf("unexpected page token %q", pageToken)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if next != "" {
			w.Header().Set("Link", fmt.Sprintf("</admin/identities?page_size=500&page_token=>; rel=\"first\",</admin/identities?page_size=500&page_token=%s>; rel=\"next\"", next))
		}
		identities := make([]map[string]string, 0, end-start)
		for id := start; id < end; id++ {
			identities = append(identities, map[string]string{"id": fmt.Sprintf("identity-%d", id)})
		}
		if err := json.NewEncoder(w).Encode(identities); err != nil {
			t.Fatalf("encode identities: %v", err)
		}
	}))
	defer server.Close()

	client := commonkratos.NewClient(server.URL)
	identities, err := client.ListIdentities(context.Background())
	if err != nil {
		t.Fatalf("list identities: %v", err)
	}
	if len(identities) != identityCount {
		t.Fatalf("got %d identities, want %d", len(identities), identityCount)
	}

	seen := make(map[string]int, identityCount)
	for _, identity := range identities {
		seen[identity.GetId()]++
	}
	for id := 0; id < identityCount; id++ {
		identityID := fmt.Sprintf("identity-%d", id)
		if seen[identityID] != 1 {
			t.Errorf("identity %q returned %d times, want once", identityID, seen[identityID])
		}
	}
	for _, pageToken := range []string{"", "after-500", "after-1000"} {
		if requests[pageToken] != 1 {
			t.Errorf("page token %q requested %d times, want once", pageToken, requests[pageToken])
		}
	}
}

func TestListIdentitiesRejectsMalformedContinuation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", `</admin/identities?page_size=500>; rel="next"`)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	identities, err := commonkratos.NewClient(server.URL).ListIdentities(context.Background())
	if err == nil || !strings.Contains(err.Error(), "page_token") {
		t.Fatalf("error = %v, want missing page_token error", err)
	}
	if identities != nil {
		t.Fatalf("identities = %v, want nil on malformed continuation", identities)
	}
}

func TestListIdentitiesRejectsRepeatedContinuation(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", `</admin/identities?page_size=500&page_token=repeated>; rel="next"`)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	identities, err := commonkratos.NewClient(server.URL).ListIdentities(context.Background())
	if err == nil || !strings.Contains(err.Error(), "repeated next page token") {
		t.Fatalf("error = %v, want repeated page token error", err)
	}
	if identities != nil {
		t.Fatalf("identities = %v, want nil on repeated continuation", identities)
	}
	if requests != 2 {
		t.Fatalf("requests = %d, want 2", requests)
	}
}

func TestListIdentitiesReturnsLaterPageErrorWithoutPartialResults(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		if requests == 1 {
			w.Header().Set("Link", `</admin/identities?page_size=500&page_token=next>; rel="next"`)
			_, _ = w.Write([]byte(`[{"id":"first-page"}]`))
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	identities, err := commonkratos.NewClient(server.URL).ListIdentities(context.Background())
	if err == nil || !strings.Contains(err.Error(), "503 Service Unavailable") {
		t.Fatalf("error = %v, want dependency status", err)
	}
	if identities != nil {
		t.Fatalf("identities = %v, want nil after later-page failure", identities)
	}
}

func TestDeactivateIdentityOnlyPatchesState(t *testing.T) {
	identityID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/admin/identities/"+identityID.String(), req.URL.Path)
		assert.Equal(t, http.MethodPatch, req.Method)
		w.Header().Set("Content-Type", "application/json")
		var body []map[string]interface{}
		if !assert.NoError(t, json.NewDecoder(req.Body).Decode(&body)) {
			return
		}
		assert.Equal(t, []map[string]interface{}{{
			"op":    "replace",
			"path":  "/state",
			"value": "inactive",
		}}, body)
		_, _ = w.Write([]byte(identityJSON(identityID, "inactive")))
	}))
	defer server.Close()

	err := commonkratos.NewClient(server.URL).DeactivateIdentity(context.Background(), identityID)

	require.NoError(t, err)
}

func TestDeactivateIdentityIsIdempotent(t *testing.T) {
	for name, status := range map[string]int{
		"already inactive": http.StatusOK,
		"missing identity": http.StatusNotFound,
	} {
		t.Run(name, func(t *testing.T) {
			identityID := uuid.New()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if status == http.StatusOK {
					_, _ = w.Write([]byte(identityJSON(identityID, "inactive")))
				}
			}))
			defer server.Close()

			err := commonkratos.NewClient(server.URL).DeactivateIdentity(context.Background(), identityID)

			require.NoError(t, err)
		})
	}
}

func TestDeactivateIdentityReturnsDependencyErrors(t *testing.T) {
	identityID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	err := commonkratos.NewClient(server.URL).DeactivateIdentity(context.Background(), identityID)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deactivate identity")
}

func TestDeleteIdentitySessions(t *testing.T) {
	testIdempotentDelete(t, "/admin/identities/%s/sessions", "delete identity sessions", func(client *commonkratos.Client, identityID uuid.UUID) error {
		return client.DeleteIdentitySessions(context.Background(), identityID)
	})
}

func TestDeleteIdentity(t *testing.T) {
	testIdempotentDelete(t, "/admin/identities/%s", "delete identity", func(client *commonkratos.Client, identityID uuid.UUID) error {
		return client.DeleteIdentity(context.Background(), identityID)
	})
}

func testIdempotentDelete(
	t *testing.T,
	pathFormat string,
	wantError string,
	execute func(*commonkratos.Client, uuid.UUID) error,
) {
	t.Helper()
	for name, status := range map[string]int{
		"success":            http.StatusNoContent,
		"already missing":    http.StatusNotFound,
		"dependency failure": http.StatusInternalServerError,
	} {
		t.Run(name, func(t *testing.T) {
			identityID := uuid.New()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, http.MethodDelete, req.Method)
				assert.Equal(t, strings.Replace(pathFormat, "%s", identityID.String(), 1), req.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
			}))
			defer server.Close()

			err := execute(commonkratos.NewClient(server.URL), identityID)
			if status == http.StatusInternalServerError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), wantError)
				return
			}
			require.NoError(t, err)
		})
	}
}

func identityJSON(identityID uuid.UUID, state string) string {
	return `{
		"id":"` + identityID.String() + `",
		"schema_id":"user",
		"schema_url":"http://kratos/schemas/user",
		"state":"` + state + `",
		"traits":{"display_name":"Keep Me","email":"keep@example.com"},
		"metadata_admin":{"support":"keep"},
		"metadata_public":{"theme":"keep"}
	}`
}
