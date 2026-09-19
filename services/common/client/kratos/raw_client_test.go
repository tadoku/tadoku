package kratos_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	kratosapi "github.com/ory/kratos-client-go"
	commonkratos "github.com/tadoku/tadoku/services/common/client/kratos"
)

func TestRawAPIClientPreservesProviderResponses(t *testing.T) {
	id := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	for _, status := range []int{http.StatusOK, http.StatusNotFound, http.StatusTooManyRequests, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			body := fmt.Sprintf(`{"error":{"code":%d,"reason":"provider detail","message":"provider failure"}}`, status)
			if status == http.StatusOK {
				body = identityJSON(id, "active")
			}
			// TLS requires the supplied HTTP client; the process default cannot
			// trust this scoped fixture's certificate.
			provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/provider/admin/identities/"+id.String() {
					t.Errorf("request=%s %s", r.Method, r.URL.String())
				}
				if r.Header.Get("Accept") != "application/json" {
					t.Errorf("Accept=%q", r.Header.Get("Accept"))
				}
				if r.URL.Query().Get("include_credential") != "oidc" {
					t.Errorf("query=%s", r.URL.RawQuery)
				}
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-Id", "provider-request")
				w.Header().Set("Retry-After", "7")
				w.WriteHeader(status)
				_, _ = w.Write([]byte(body))
			}))
			t.Cleanup(provider.Close)
			httpClient := provider.Client()
			httpClient.Timeout = time.Second
			t.Cleanup(httpClient.CloseIdleConnections)
			client := commonkratos.NewAPIClient(provider.URL+"/provider", httpClient)

			identity, response, err := client.IdentityApi.GetIdentity(t.Context(), id.String()).IncludeCredential([]string{"oidc"}).Execute()
			if response == nil {
				t.Fatalf("response missing: %v", err)
			}
			t.Cleanup(func() { _ = response.Body.Close() })
			if response.StatusCode != status || response.Header.Get("X-Request-Id") != "provider-request" || response.Header.Get("Retry-After") != "7" {
				t.Errorf("provider response changed: %v", response)
			}
			if status != http.StatusOK {
				var providerErr *kratosapi.GenericOpenAPIError
				if !errors.As(err, &providerErr) {
					t.Fatalf("error=%v, want SDK provider error", err)
				}
				if string(providerErr.Body()) != body {
					t.Errorf("error body=%s, want %s", providerErr.Body(), body)
				}
				model, ok := providerErr.Model().(kratosapi.ErrorGeneric)
				if !ok || model.Error.GetReason() != "provider detail" || model.Error.GetCode() != int64(status) {
					t.Errorf("provider error model changed: %v", providerErr.Model())
				}
				if errors.Is(err, commonkratos.ErrNotFound) {
					t.Error("raw client translated the provider error to a legacy sentinel")
				}
				return
			}
			if err != nil {
				t.Fatalf("get identity: %v", err)
			}
			if identity == nil || identity.GetId() != id.String() || identity.GetState() != kratosapi.IDENTITYSTATE_ACTIVE {
				t.Fatalf("identity=%v", identity)
			}
			traits, ok := identity.GetTraits().(map[string]interface{})
			if !ok || traits["email"] != "keep@example.com" || traits["display_name"] != "Keep Me" {
				t.Errorf("provider traits changed: %v", identity.GetTraits())
			}
			admin, ok := identity.GetMetadataAdmin().(map[string]interface{})
			if !ok || admin["support"] != "keep" {
				t.Errorf("provider admin metadata changed: %v", identity.GetMetadataAdmin())
			}
		})
	}
}

func TestRawAPIClientPreservesCancellation(t *testing.T) {
	for _, stage := range []string{"headers", "body"} {
		for _, cause := range []string{"caller cancel", "caller deadline", "client timeout"} {
			t.Run(stage+"/"+cause, func(t *testing.T) {
				started := make(chan struct{})
				canceled := make(chan struct{})
				provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if stage == "body" {
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{"id":`))
						w.(http.Flusher).Flush()
					}
					close(started)
					<-r.Context().Done()
					close(canceled)
				}))
				t.Cleanup(provider.Close)
				httpClient := &http.Client{
					Transport: http.DefaultTransport.(*http.Transport).Clone(),
					Timeout:   time.Second,
				}
				t.Cleanup(httpClient.CloseIdleConnections)
				client := commonkratos.NewAPIClient(provider.URL, httpClient)
				timeout := 3 * time.Second
				if cause == "caller deadline" {
					timeout = 100 * time.Millisecond
				}
				ctx, cancel := context.WithTimeout(t.Context(), timeout)
				t.Cleanup(cancel)
				if cause == "client timeout" {
					httpClient.Timeout = 100 * time.Millisecond
				}

				result := make(chan error, 1)
				go func() {
					_, _, err := client.IdentityApi.GetIdentity(ctx, "synthetic").Execute()
					result <- err
				}()
				select {
				case <-started:
				case <-time.After(2 * time.Second):
					t.Fatal("provider did not receive the request")
				}
				want := context.DeadlineExceeded
				if cause == "caller cancel" {
					want = context.Canceled
					cancel()
				}
				select {
				case err := <-result:
					if !errors.Is(err, want) {
						t.Errorf("error=%v, want %v", err, want)
					}
				case <-time.After(2 * time.Second):
					t.Fatal("SDK request did not terminate within the bound")
				}
				select {
				case <-canceled:
				case <-time.After(time.Second):
					t.Error("provider did not observe request cancellation")
				}
			})
		}
	}
}
