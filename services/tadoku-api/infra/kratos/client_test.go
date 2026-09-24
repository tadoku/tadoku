package kratos_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	kratosclient "github.com/tadoku/tadoku/services/tadoku-api/infra/kratos"
)

func TestClientUsesConfiguredHTTPClientForAllRequests(t *testing.T) {
	id := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	// Only the supplied HTTP client trusts this scoped provider's certificate.
	provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/admin/identities/" + id.String():
			_, _ = w.Write([]byte(identityJSON(id, "active")))
		case "/admin/identities":
			if r.URL.Query().Get("page_size") != "10" || r.URL.Query().Get("page_token") != "next" {
				t.Errorf("unexpected pagination query: %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte("[" + identityJSON(id, "active") + "]"))
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.String())
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(provider.Close)
	httpClient := provider.Client()
	httpClient.Timeout = time.Second
	t.Cleanup(httpClient.CloseIdleConnections)
	client := kratosclient.NewClient(provider.URL, kratosclient.WithHTTPClient(httpClient))

	identity, err := client.FetchIdentity(t.Context(), id)
	if err != nil {
		t.Fatalf("fetch identity using supplied HTTP client: %v", err)
	}
	if identity == nil || identity.GetId() != id.String() {
		t.Errorf("identity=%v, want %s", identity, id)
	}

	identities, next, err := client.ListIdentities(t.Context(), 10, "next")
	if err != nil {
		t.Fatalf("list identities using supplied HTTP client: %v", err)
	}
	if len(identities) != 1 || identities[0].GetId() != id.String() || next != "" {
		t.Errorf("identities=%v next=%q", identities, next)
	}
}

func TestListIdentitiesUsesConfiguredHTTPTimeout(t *testing.T) {
	canceled := make(chan struct{})
	provider := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		close(canceled)
	}))
	t.Cleanup(provider.Close)
	httpClient := provider.Client()
	httpClient.Timeout = 100 * time.Millisecond
	t.Cleanup(httpClient.CloseIdleConnections)
	client := kratosclient.NewClient(provider.URL, kratosclient.WithHTTPClient(httpClient))

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	t.Cleanup(cancel)
	_, _, err := client.ListIdentities(ctx, 10, "")
	if !errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
		t.Errorf("list error=%v, caller error=%v; want configured HTTP timeout", err, ctx.Err())
	}
	select {
	case <-canceled:
	case <-time.After(time.Second):
		t.Error("provider did not observe timeout cancellation")
	}
}

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

	client := kratosclient.NewClient(server.URL)
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

	identities, next, err := kratosclient.NewClient(server.URL).ListIdentities(context.Background(), 500, "")
	if err == nil || !strings.Contains(err.Error(), "page_token") {
		t.Fatalf("error = %v, want missing page_token error", err)
	}
	if identities != nil || next != "" {
		t.Fatalf("identities = %v, next = %q, want no result on malformed continuation", identities, next)
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
