package keto

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	ketoapi "github.com/ory/keto-client-go"
)

func TestRelationWriteProviderResponses(t *testing.T) {
	for _, operation := range []string{"add", "delete"} {
		for _, status := range []int{http.StatusConflict, http.StatusNotFound, http.StatusTooManyRequests, http.StatusServiceUnavailable} {
			t.Run(operation+"/"+http.StatusText(status), func(t *testing.T) {
				body := fmt.Sprintf(`{"error":{"code":%d,"message":"provider detail"}}`, status)
				var requests atomic.Int32
				provider := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					_, _ = io.Copy(io.Discard, r.Body)
					method := http.MethodPut
					if operation == "delete" {
						method = http.MethodDelete
					}
					if r.Method != method || r.URL.Path != "/provider/admin/relation-tuples" {
						t.Errorf("request=%s %s", r.Method, r.URL)
					}
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("Retry-After", "0")
					w.WriteHeader(status)
					_, _ = w.Write([]byte(body))
				}))
				t.Cleanup(provider.Close)
				httpClient := provider.Client()
				httpClient.Timeout = time.Second
				t.Cleanup(httpClient.CloseIdleConnections)
				client := NewClient("http://unused-read.test", provider.URL+"/provider", WithHTTPClient(httpClient))
				write := client.AddRelation
				if operation == "delete" {
					write = client.DeleteRelation
				}

				err := write(t.Context(), "app", "synthetic", "admins", Subject{ID: "synthetic"})
				idempotent := operation == "add" && status == http.StatusConflict || operation == "delete" && status == http.StatusNotFound
				if idempotent {
					if err != nil {
						t.Errorf("idempotent write: %v", err)
					}
				} else {
					var providerErr *ketoapi.GenericOpenAPIError
					if !errors.As(err, &providerErr) {
						t.Fatalf("error=%v, want SDK provider error", err)
					}
					if string(providerErr.Body()) != body {
						t.Errorf("provider error body=%s, want %s", providerErr.Body(), body)
					}
				}
				if got := requests.Load(); got != 1 {
					t.Errorf("write sent %d requests, want 1", got)
				}
			})
		}
	}
}

func TestRelationWritesAreCancellableAndBounded(t *testing.T) {
	for _, operation := range []string{"add", "delete"} {
		for _, stage := range []string{"headers", "body"} {
			for _, cause := range []string{"caller cancel", "caller deadline", "client timeout"} {
				t.Run(operation+"/"+stage+"/"+cause, func(t *testing.T) {
					started := make(chan struct{}, 1)
					canceled := make(chan struct{}, 1)
					var requests atomic.Int32
					provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests.Add(1)
						_, _ = io.Copy(io.Discard, r.Body)
						if stage == "body" {
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusServiceUnavailable)
							_, _ = w.Write([]byte(`{"error":`))
							w.(http.Flusher).Flush()
						}
						started <- struct{}{}
						<-r.Context().Done()
						canceled <- struct{}{}
					}))
					t.Cleanup(provider.Close)
					httpClient := &http.Client{
						Transport: http.DefaultTransport.(*http.Transport).Clone(),
						Timeout:   3 * time.Second,
					}
					t.Cleanup(httpClient.CloseIdleConnections)
					client := NewClient("http://unused-read.test", provider.URL, WithHTTPClient(httpClient))
					write := client.AddRelation
					if operation == "delete" {
						write = client.DeleteRelation
					}
					timeout := 5 * time.Second
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
						result <- write(ctx, "app", "synthetic", "admins", Subject{ID: "synthetic"})
					}()

					select {
					case <-started:
					case <-time.After(2 * time.Second):
						t.Fatal("provider did not receive request")
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
						if cause == "client timeout" && ctx.Err() != nil {
							t.Errorf("client timeout did not expire before caller: %v", ctx.Err())
						}
					case <-time.After(2 * time.Second):
						t.Fatal("write did not terminate within bound")
					}
					select {
					case <-canceled:
					case <-time.After(time.Second):
						t.Error("provider did not observe cancellation")
					}
					if got := requests.Load(); got != 1 {
						t.Errorf("write sent %d requests, want 1", got)
					}
				})
			}
		}
	}
}
