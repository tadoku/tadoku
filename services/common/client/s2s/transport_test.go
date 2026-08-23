package s2s

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
