package kratos_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commonkratos "github.com/tadoku/tadoku/services/common/client/kratos"
)

func TestListIdentitiesFetchesOnlyRequestedPage(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests++
		if req.URL.Path != "/admin/identities" || req.URL.Query().Get("page_size") != "600" {
			t.Errorf("unexpected requested page: %s", req.URL.String())
		}
		if req.URL.Query().Has("page") || req.URL.Query().Has("per_page") {
			t.Errorf("request used unsupported offset pagination: %s", req.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		switch req.URL.Query().Get("page_token") {
		case "":
			w.Header().Set("Link", `</admin/identities?page_size=600&page_token=>; rel="first",</admin/identities?page_size=600&page_token=second>; rel="next"`)
			_, _ = w.Write([]byte(`[{"id":"first-page"}]`))
		case "second":
			_, _ = w.Write([]byte(`[{"id":"second-page"}]`))
		default:
			t.Errorf("unexpected page token %q", req.URL.Query().Get("page_token"))
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	client := commonkratos.NewClient(server.URL)
	identities, next, err := client.ListIdentities(context.Background(), 600, "")
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if len(identities) != 1 || identities[0].GetId() != "first-page" || next != "second" || requests != 1 {
		t.Fatalf("first page: identities=%v next=%q requests=%d", identities, next, requests)
	}
	identities, next, err = client.ListIdentities(context.Background(), 600, next)
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(identities) != 1 || identities[0].GetId() != "second-page" || next != "" || requests != 2 {
		t.Fatalf("second page: identities=%v next=%q requests=%d", identities, next, requests)
	}
}

func TestListIdentitiesRejectsMalformedContinuation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Link", `</admin/identities?page_size=500>; rel="next"`)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer server.Close()

	identities, next, err := commonkratos.NewClient(server.URL).ListIdentities(context.Background(), 500, "")
	if err == nil || !strings.Contains(err.Error(), "page_token") {
		t.Fatalf("error = %v, want missing page_token error", err)
	}
	if identities != nil || next != "" {
		t.Fatalf("identities = %v, next = %q, want no result on malformed continuation", identities, next)
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
