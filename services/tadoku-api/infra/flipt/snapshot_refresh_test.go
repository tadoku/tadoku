package flipt

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	featureflags "github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestClientClearsStaleAfterUnchangedSnapshotRecovers(t *testing.T) {
	var mode atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		switch mode.Load() {
		case 0:
			response.Header().Set("ETag", `"v1"`)
			_, _ = response.Write([]byte(booleanSnapshot(nil)))
		case 1:
			response.WriteHeader(http.StatusNotFound)
		default:
			response.WriteHeader(http.StatusNotModified)
		}
	}))
	t.Cleanup(server.Close)
	observer := &recordingObserver{}
	client, err := New(context.Background(), Config{
		URL:            server.URL,
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Second,
		RequestTimeout: time.Second,
		StartupTimeout: time.Second,
	}, observer)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })

	mode.Store(1)
	require.Eventually(t, func() bool {
		return client.client.Err() != nil
	}, 3*time.Second, 20*time.Millisecond)
	mode.Store(2)
	require.Eventually(t, func() bool {
		_, refreshes, _ := observer.snapshot()
		return refreshes >= 3
	}, 4*time.Second, 20*time.Millisecond)

	result, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
		FlagKey:  "release-log-entry-v2",
		EntityID: "4c47b265-6987-4bc1-8933-006168a31793",
	})
	require.NoError(t, err)
	assert.False(t, result.Stale)
}

func TestClientConfirmsChangedSnapshotAfterRecovery(t *testing.T) {
	var mode atomic.Int32
	changedServed := make(chan struct{})
	var changedOnce sync.Once
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		switch mode.Load() {
		case 0:
			response.Header().Set("ETag", `"v1"`)
			_, _ = response.Write([]byte(booleanSnapshot(nil)))
		case 1:
			response.WriteHeader(http.StatusNotFound)
		case 2:
			response.Header().Set("ETag", `"v2"`)
			changedOnce.Do(func() { close(changedServed) })
			changed := strings.Replace(booleanSnapshot(nil), `"enabled": false`, `"enabled": true`, 1)
			_, _ = response.Write([]byte(changed))
		default:
			response.WriteHeader(http.StatusNotModified)
		}
	}))
	t.Cleanup(server.Close)
	observer := &recordingObserver{}
	client, err := New(context.Background(), Config{
		URL:            server.URL,
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Second,
		RequestTimeout: time.Second,
		StartupTimeout: time.Second,
	}, observer)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })

	mode.Store(1)
	require.Eventually(t, func() bool { return client.client.Err() != nil }, 3*time.Second, 20*time.Millisecond)
	mode.Store(2)
	select {
	case <-changedServed:
	case <-time.After(3 * time.Second):
		require.Fail(t, "changed snapshot was not fetched")
	}
	require.Eventually(t, func() bool { return client.client.Err() == nil }, 3*time.Second, 20*time.Millisecond)
	_, refreshesBeforeEvaluation, _ := observer.snapshot()
	assert.Equal(t, 1, refreshesBeforeEvaluation)

	result, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
		FlagKey:  "release-log-entry-v2",
		EntityID: "4c47b265-6987-4bc1-8933-006168a31793",
	})
	require.NoError(t, err)
	assert.True(t, result.Enabled)
	assert.False(t, result.Stale)
	_, refreshesAfterEvaluation, _ := observer.snapshot()
	assert.Equal(t, 2, refreshesAfterEvaluation)
}

func TestObservedTransportTracksRefreshesAndBoundedFetchErrors(t *testing.T) {
	observer := &recordingObserver{}
	statuses := []int{http.StatusOK, http.StatusNotModified, http.StatusServiceUnavailable}
	transport := &observedTransport{
		state: newProviderFetchState("default", observer),
		base: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			status := statuses[0]
			statuses = statuses[1:]
			body := io.ReadCloser(http.NoBody)
			header := make(http.Header)
			if status == http.StatusOK {
				body = io.NopCloser(strings.NewReader(booleanSnapshot(nil)))
				header.Set("ETag", `"fixture"`)
			}
			return &http.Response{StatusCode: status, Body: body, Header: header}, nil
		}),
	}
	request := httptest.NewRequest(http.MethodGet, "http://flipt/snapshot", nil)

	_, err := transport.RoundTrip(request)
	require.NoError(t, err)
	transport.state.confirmPendingSnapshot()
	request.Header.Set("If-None-Match", `"fixture"`)
	_, err = transport.RoundTrip(request)
	require.NoError(t, err)
	_, err = transport.RoundTrip(request)
	require.NoError(t, err)

	_, refreshes, providerErrors := observer.snapshot()
	assert.Equal(t, 2, refreshes)
	assert.Equal(t, []string{"fetch"}, providerErrors)
}

func TestObservedTransportConfirmsRecoveredSnapshotOnMatchingNotModified(t *testing.T) {
	observer := &recordingObserver{}
	statuses := []int{http.StatusOK, http.StatusNotModified, http.StatusNotModified}
	state := newProviderFetchState("default", observer)
	transport := &observedTransport{
		state: state,
		base: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			status := statuses[0]
			statuses = statuses[1:]
			body := io.ReadCloser(http.NoBody)
			header := make(http.Header)
			if status == http.StatusOK {
				body = io.NopCloser(strings.NewReader(booleanSnapshot(nil)))
				header.Set("ETag", `"recovered"`)
			}
			return &http.Response{StatusCode: status, Body: body, Header: header}, nil
		}),
	}
	request := httptest.NewRequest(http.MethodGet, "http://flipt/snapshot", nil)

	_, err := transport.RoundTrip(request)
	require.NoError(t, err)
	request.Header.Set("If-None-Match", `"recovered"`)
	_, err = transport.RoundTrip(request)
	require.NoError(t, err)
	_, err = transport.RoundTrip(request)
	require.NoError(t, err)

	_, refreshes, providerErrors := observer.snapshot()
	assert.Equal(t, 2, refreshes)
	assert.Empty(t, providerErrors)
	assert.False(t, state.isStale(false))
}

func TestFetchStateRejectsMismatchedPendingETagWithoutReplacingConfirmedSnapshot(t *testing.T) {
	observer := &recordingObserver{}
	state := newProviderFetchState("default", observer)
	state.snapshotReceived(`"v1"`)
	state.confirmPendingSnapshot()
	state.snapshotReceived(`"v2"`)

	state.snapshotNotModified(`"v1"`)

	_, refreshes, providerErrors := observer.snapshot()
	assert.Equal(t, 1, refreshes)
	assert.Equal(t, []string{"fetch"}, providerErrors)
	assert.True(t, state.isStale(false))

	state.snapshotNotModified(`"v1"`)

	_, refreshes, providerErrors = observer.snapshot()
	assert.Equal(t, 2, refreshes)
	assert.Equal(t, []string{"fetch"}, providerErrors)
	assert.False(t, state.isStale(false))

	state.snapshotNotModified("")

	_, refreshes, providerErrors = observer.snapshot()
	assert.Equal(t, 2, refreshes)
	assert.Equal(t, []string{"fetch", "fetch"}, providerErrors)
	assert.True(t, state.isStale(false))
}
