package s2s

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

func TestAuthTransportCancelsTokenExchangeWithRequest(t *testing.T) {
	exchangeStarted := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		close(exchangeStarted)
		<-request.Context().Done()
	}))
	t.Cleanup(server.Close)

	tokenPath := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenPath, []byte("projected-token"), 0o600))
	client := &Client{
		oathkeeperURL: server.URL,
		k8sTokenPath:  tokenPath,
		httpClient:    &http.Client{Timeout: time.Second},
		clock:         commondomain.NewMockClock(time.Time{}),
		tokenCache:    make(map[string]*cachedToken),
	}
	transport := NewAuthTransport(client, "flipt-evaluation/immersion-api", http.DefaultTransport)
	ctx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, "http://flipt.test/snapshot", nil).WithContext(ctx)

	result := make(chan error, 1)
	go func() {
		_, err := transport.RoundTrip(request)
		result <- err
	}()
	<-exchangeStarted
	cancel()

	select {
	case err := <-result:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(200 * time.Millisecond):
		assert.Fail(t, "token exchange did not honor request cancellation")
	}
}

func TestClientUsesInjectedClockForTokenCacheExpiry(t *testing.T) {
	var exchanges atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		exchanges.Add(1)
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"access_token":"token","token_type":"bearer","expires_in":600}`))
	}))
	t.Cleanup(server.Close)
	tokenPath := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenPath, []byte("projected-token"), 0o600))
	clock := commondomain.NewMockClock(time.Date(2026, 8, 24, 6, 0, 0, 0, time.UTC))
	client := &Client{
		oathkeeperURL: server.URL,
		k8sTokenPath:  tokenPath,
		httpClient:    &http.Client{Timeout: time.Second},
		clock:         clock,
		tokenCache:    make(map[string]*cachedToken),
	}

	_, err := client.GetTokenContext(context.Background(), "flipt-evaluation/immersion-api")
	require.NoError(t, err)
	_, err = client.GetTokenContext(context.Background(), "flipt-evaluation/immersion-api")
	require.NoError(t, err)
	assert.EqualValues(t, 1, exchanges.Load())

	clock.SetTime(clock.Now().Add(301 * time.Second))
	_, err = client.GetTokenContext(context.Background(), "flipt-evaluation/immersion-api")
	require.NoError(t, err)
	assert.EqualValues(t, 2, exchanges.Load())
}

func TestClientRejectsInvalidSuccessfulTokenExchange(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty access token", body: `{"access_token":"","token_type":"bearer","expires_in":600}`},
		{name: "wrong token type", body: `{"access_token":"token","token_type":"mac","expires_in":600}`},
		{name: "missing expiry", body: `{"access_token":"token","token_type":"bearer"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				response.Header().Set("Content-Type", "application/json")
				_, _ = response.Write([]byte(tt.body))
			}))
			t.Cleanup(server.Close)
			tokenPath := filepath.Join(t.TempDir(), "token")
			if err := os.WriteFile(tokenPath, []byte("projected-token"), 0o600); err != nil {
				t.Fatal(err)
			}
			client := &Client{
				oathkeeperURL: server.URL,
				k8sTokenPath:  tokenPath,
				httpClient:    &http.Client{Timeout: time.Second},
				clock:         commondomain.NewMockClock(time.Time{}),
				tokenCache:    make(map[string]*cachedToken),
			}

			token, err := client.GetTokenContext(context.Background(), "flipt-evaluation/tadoku-api")

			if err == nil || token != "" {
				t.Errorf("token=%q error=%v", token, err)
			}
		})
	}
}

func TestAuthTransportUsesTadokuFliptAudience(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenPath, []byte("projected-token"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, target := range []string{"flipt-evaluation/tadoku-api", "flipt-management/tadoku-api"} {
		t.Run(target, func(t *testing.T) {
			exchange := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if request.URL.Path != "/token-exchange/"+target {
					t.Errorf("exchange path=%q", request.URL.Path)
				}
				if got := request.Header.Get("Authorization"); got != "Bearer projected-token" {
					t.Errorf("exchange authorization=%q", got)
				}
				_, _ = response.Write([]byte(`{"access_token":"exchanged-token","token_type":"bearer","expires_in":600}`))
			}))
			t.Cleanup(exchange.Close)
			downstream := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if got := request.Header.Get("Authorization"); got != "Bearer exchanged-token" {
					t.Errorf("downstream authorization=%q", got)
				}
				response.WriteHeader(http.StatusNoContent)
			}))
			t.Cleanup(downstream.Close)

			client := &Client{
				oathkeeperURL: exchange.URL,
				k8sTokenPath:  tokenPath,
				httpClient:    &http.Client{Timeout: time.Second},
				clock:         commondomain.NewMockClock(time.Time{}),
				tokenCache:    make(map[string]*cachedToken),
			}
			httpClient := &http.Client{Transport: NewAuthTransport(client, target, http.DefaultTransport)}

			response, err := httpClient.Get(downstream.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != http.StatusNoContent {
				t.Errorf("status=%d", response.StatusCode)
			}
		})
	}
}
