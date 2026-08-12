package observability

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerStartsAndShutsDownGracefully(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	handler := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-releaseRequest
		response.WriteHeader(http.StatusNoContent)
	})
	server := NewServer("127.0.0.1:0", handler)
	require.NoError(t, server.Start())

	requestDone := make(chan error, 1)
	go func() {
		response, err := http.Get("http://" + server.Addr())
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			err = response.Body.Close()
		}
		requestDone <- err
	}()
	<-requestStarted

	shutdownDone := make(chan error, 1)
	shutdownContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() {
		shutdownDone <- server.Shutdown(shutdownContext)
	}()

	select {
	case err := <-shutdownDone:
		require.Failf(t, "shutdown returned before active request completed", "error: %v", err)
	case <-time.After(25 * time.Millisecond):
	}
	close(releaseRequest)
	require.NoError(t, <-requestDone)
	require.NoError(t, <-shutdownDone)
	_, err := http.Get("http://" + server.Addr())
	assert.Error(t, err)
}
