package flipt

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	featureflags "github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	fliptsdk "go.flipt.io/flipt-client"
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
		FlagKey:  "release-webv3",
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
		FlagKey:  "release-webv3",
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

func TestClientNeverReportsMissingFlagsFromTheStartupFallbackEngine(t *testing.T) {
	rejected := strings.Replace(booleanSnapshot(nil), `"rollouts": []`, `"rollouts": {}`, 1)
	var served atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		if served.Add(1) == 1 {
			response.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = response.Write([]byte(rejected))
	}))
	t.Cleanup(server.Close)
	client, err := New(context.Background(), Config{
		URL:            server.URL,
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Second,
		RequestTimeout: time.Second,
		StartupTimeout: time.Second,
	}, &recordingObserver{})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })

	require.Eventually(t, func() bool { return served.Load() >= 4 }, 6*time.Second, 20*time.Millisecond)
	_, err = client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
		FlagKey:  "other-flag",
		EntityID: "tadoku/prod",
	})
	require.Error(t, err)
	assert.NotErrorIs(t, err, featureflags.ErrFlagNotFound)
}

func TestClientNeverInstallsAFailedStartAfterARejectedUpdateClearsItsError(t *testing.T) {
	rejected := strings.Replace(booleanSnapshot(nil), `"rollouts": []`, `"rollouts": {}`, 1)
	enabled := strings.Replace(booleanSnapshot(nil), `"enabled": false`, `"enabled": true`, 1)
	const (
		outage int32 = iota
		rejecting
		serving
	)
	var mode, rejections atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		switch mode.Load() {
		case outage:
			response.WriteHeader(http.StatusServiceUnavailable)
		case rejecting:
			rejections.Add(1)
			_, _ = response.Write([]byte(rejected))
		default:
			_, _ = response.Write([]byte(enabled))
		}
	}))
	t.Cleanup(server.Close)

	var opens atomic.Int32
	held := make(chan *fliptsdk.Client, 1)
	released := make(chan struct{})
	release := sync.OnceFunc(func() { close(released) })
	t.Cleanup(release)
	observer := &recordingObserver{}
	client, err := newClient(context.Background(), Config{
		URL:            server.URL,
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Second,
		RequestTimeout: time.Second,
		StartupTimeout: 100 * time.Millisecond,
	}, observer, func(ctx context.Context, options ...fliptsdk.Option) (sdkClient, error) {
		sdk, err := fliptsdk.NewClient(ctx, options...)
		if opens.Add(1) == 1 && err == nil {
			held <- sdk
			<-released
		}
		return sdk, err
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })

	var first *fliptsdk.Client
	select {
	case first = <-held:
	case <-time.After(5 * time.Second):
		require.Fail(t, "the first SDK client was never constructed")
	}
	require.Error(t, first.Err(), "the first SDK client must start from a failed fetch")
	mode.Store(rejecting)
	require.Eventually(t, func() bool { return rejections.Load() > 0 && first.Err() == nil }, 5*time.Second, 10*time.Millisecond)
	release()
	require.Eventually(t, func() bool {
		statuses, _, _ := observer.snapshot()
		return opens.Load() > 1 || slices.Contains(statuses, featureflags.InitializationStatusReady)
	}, 5*time.Second, 10*time.Millisecond)
	requireNoMissingFlag := func(when string) {
		t.Helper()
		_, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
			FlagKey:  "other-flag",
			EntityID: "tadoku/prod",
		})
		require.Error(t, err, when)
		require.NotErrorIs(t, err, featureflags.ErrFlagNotFound, when)
	}
	requireNoMissingFlag("after the failed start cleared its error")

	attempts := opens.Load()
	require.Eventually(t, func() bool { return opens.Load() > attempts+1 }, 5*time.Second, 10*time.Millisecond)
	requireNoMissingFlag("after an attempt fetched a snapshot its engine rejected")
	mode.Store(outage)
	attempts = opens.Load()
	require.Eventually(t, func() bool { return opens.Load() > attempts+1 }, 5*time.Second, 10*time.Millisecond)
	requireNoMissingFlag("after a later attempt failed its initial fetch")

	mode.Store(serving)
	require.Eventually(t, func() bool {
		result, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
			FlagKey:  "release-webv3",
			EntityID: "4c47b265-6987-4bc1-8933-006168a31793",
		})
		return err == nil && result.Enabled
	}, 5*time.Second, 10*time.Millisecond)
	_, err = client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
		FlagKey:  "other-flag",
		EntityID: "tadoku/prod",
	})
	assert.ErrorIs(t, err, featureflags.ErrFlagNotFound, "a fetched snapshot without the flag reports it missing")
}

func TestClientKeepsItsDecisionWhenAnUpdateFailsValidation(t *testing.T) {
	enabled := strings.NewReplacer(`"release-webv3"`, `"other-flag"`, `"enabled": false`, `"enabled": true`).
		Replace(booleanSnapshot(nil))
	for _, test := range []struct{ name, body string }{
		{"another namespace", `{"namespace":{"key":"other"},"flags":[],"digest":"rev-2"}`},
		{"no digest", `{"namespace":{"key":"default"},"flags":[]}`},
		{
			"namespace key in another case",
			`{"namespace":{"key":"other"},"NAMESPACE":{"key":"default"},"flags":[],"digest":"rev-2"}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var replaced atomic.Bool
			var replacements atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
				if replaced.Load() {
					replacements.Add(1)
					_, _ = response.Write([]byte(test.body))
					return
				}
				_, _ = response.Write([]byte(enabled))
			}))
			t.Cleanup(server.Close)
			client, err := New(context.Background(), Config{
				URL:            server.URL,
				Environment:    "local",
				Namespace:      "default",
				UpdateInterval: time.Second,
				RequestTimeout: time.Second,
				StartupTimeout: time.Second,
			}, &recordingObserver{})
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })
			evaluate := func() (featureflags.ProviderResult, error) {
				return client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
					FlagKey:  "other-flag",
					EntityID: "tadoku/prod",
				})
			}
			result, err := evaluate()
			require.NoError(t, err)
			require.True(t, result.Enabled)

			replaced.Store(true)
			require.Eventually(t, func() bool { return replacements.Load() >= 4 }, 10*time.Second, 20*time.Millisecond)
			result, err = evaluate()
			require.NoError(t, err, "a rejected update replaced the evaluated snapshot")
			assert.True(t, result.Enabled)
			assert.True(t, result.Stale)

			replaced.Store(false)
			require.Eventually(t, func() bool {
				result, err := evaluate()
				return err == nil && result.Enabled && !result.Stale
			}, 5*time.Second, 20*time.Millisecond)
		})
	}
}
