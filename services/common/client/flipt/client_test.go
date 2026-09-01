package flipt

import (
	"context"
	"errors"
	"fmt"
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
	featureflags "github.com/tadoku/tadoku/services/common/featureflags"
	fliptsdk "go.flipt.io/flipt-client"
)

type fakeSDKClient struct {
	response *fliptsdk.BooleanEvaluationResponse
	err      error
	stateErr error
	request  *fliptsdk.EvaluationRequest
	closed   bool
}

func (c *fakeSDKClient) EvaluateBoolean(_ context.Context, request *fliptsdk.EvaluationRequest) (*fliptsdk.BooleanEvaluationResponse, error) {
	c.request = request
	return c.response, c.err
}

func (c *fakeSDKClient) Err() error { return c.stateErr }

func (c *fakeSDKClient) Close(_ context.Context) error {
	c.closed = true
	return nil
}

type recordingObserver struct {
	mu        sync.Mutex
	statuses  []featureflags.InitializationStatus
	refreshes int
	errors    []string
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func (o *recordingObserver) ObserveInitialization(status featureflags.InitializationStatus) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.statuses = append(o.statuses, status)
}

func (o *recordingObserver) ObserveConfigRefresh() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.refreshes++
}
func (o *recordingObserver) ObserveProviderError(kind string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.errors = append(o.errors, kind)
}

func (o *recordingObserver) snapshot() ([]featureflags.InitializationStatus, int, []string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]featureflags.InitializationStatus(nil), o.statuses...), o.refreshes, append([]string(nil), o.errors...)
}

func TestClientMapsSuccessfulAndStaleBooleanDecisions(t *testing.T) {
	tests := []struct {
		name      string
		stateErr  error
		wantStale bool
	}{
		{name: "fresh"},
		{name: "last known good after provider outage", stateErr: errors.New("provider unavailable"), wantStale: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sdk := &fakeSDKClient{
				response: &fliptsdk.BooleanEvaluationResponse{
					Enabled: true,
					FlagKey: "release-log-entry-v2",
					Reason:  "MATCH_EVALUATION_REASON",
				},
				stateErr: tt.stateErr,
			}
			client := &Client{client: sdk}
			request := featureflags.EvaluationRequest{
				FlagKey:  "release-log-entry-v2",
				EntityID: "4c47b265-6987-4bc1-8933-006168a31793",
				Context:  map[string]string{"authenticated": "true"},
			}

			result, err := client.EvaluateBoolean(context.Background(), request)

			require.NoError(t, err)
			assert.True(t, result.Enabled)
			assert.Equal(t, "MATCH_EVALUATION_REASON", result.Reason)
			assert.Equal(t, tt.wantStale, result.Stale)
			require.NotNil(t, sdk.request)
			assert.Equal(t, request.FlagKey, sdk.request.FlagKey)
			assert.Equal(t, request.EntityID, sdk.request.EntityID)
			assert.Equal(t, request.Context, sdk.request.Context)
		})
	}
}

func TestClientRejectsMalformedAndMissingResponses(t *testing.T) {
	tests := []struct {
		name      string
		sdk       *fakeSDKClient
		wantError error
	}{
		{name: "nil response", sdk: &fakeSDKClient{}, wantError: featureflags.ErrInvalidResponse},
		{name: "wrong flag", sdk: &fakeSDKClient{response: &fliptsdk.BooleanEvaluationResponse{FlagKey: "different"}}, wantError: featureflags.ErrInvalidResponse},
		{name: "missing flag", sdk: &fakeSDKClient{err: errors.New("flag not found")}, wantError: featureflags.ErrFlagNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{client: tt.sdk}
			_, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "release-log-entry-v2"})
			if tt.wantError != nil {
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestClientHonorsCancellationWithoutCallingSDK(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sdk := &fakeSDKClient{response: &fliptsdk.BooleanEvaluationResponse{FlagKey: "release-log-entry-v2"}}
	client := &Client{client: sdk}

	_, err := client.EvaluateBoolean(ctx, featureflags.EvaluationRequest{FlagKey: "release-log-entry-v2"})

	assert.ErrorIs(t, err, context.Canceled)
	assert.Nil(t, sdk.request)
}

func TestNewUsesFallbackWhenFliptIsEmptyOrUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/internal/v1/evaluation/snapshot/namespace/default", request.URL.Path)
		assert.Equal(t, "local", request.Header.Get("x-flipt-environment"))
		http.NotFound(response, request)
	}))
	t.Cleanup(server.Close)
	observer := &recordingObserver{}

	client, err := New(context.Background(), Config{
		URL:            server.URL,
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Minute,
		RequestTimeout: time.Second,
		StartupTimeout: time.Second,
	}, observer)

	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })
	statuses, _, providerErrors := observer.snapshot()
	assert.Equal(t, []featureflags.InitializationStatus{featureflags.InitializationStatusFallback}, statuses)
	assert.Contains(t, providerErrors, "fetch")
	_, err = client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "release-log-entry-v2", EntityID: "4c47b265-6987-4bc1-8933-006168a31793"})
	assert.ErrorIs(t, err, featureflags.ErrFlagNotFound)
}

func TestNewDoesNotMarkMalformedSnapshotAsFresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte(`{"version":`))
	}))
	t.Cleanup(server.Close)
	observer := &recordingObserver{}

	client, err := New(context.Background(), Config{
		URL:            server.URL,
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Minute,
		RequestTimeout: time.Second,
		StartupTimeout: time.Second,
	}, observer)
	if client != nil {
		t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })
	}

	assert.Error(t, err)
	_, refreshes, providerErrors := observer.snapshot()
	assert.Zero(t, refreshes)
	assert.Contains(t, providerErrors, "fetch")
}

func TestNewDoesNotMarkSemanticallyInvalidSnapshotAsFresh(t *testing.T) {
	invalidSnapshot := strings.Replace(booleanSnapshot(nil), `"rollouts": []`, `"rollouts": {}`, 1)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
		_, _ = response.Write([]byte(invalidSnapshot))
	}))
	t.Cleanup(server.Close)
	observer := &recordingObserver{}

	client, err := New(context.Background(), Config{
		URL:            server.URL,
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Minute,
		RequestTimeout: time.Second,
		StartupTimeout: time.Second,
	}, observer)
	if client != nil {
		t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })
	}

	assert.Error(t, err)
	statuses, refreshes, providerErrors := observer.snapshot()
	assert.Zero(t, refreshes)
	assert.Equal(t, []featureflags.InitializationStatus{featureflags.InitializationStatusError}, statuses)
	assert.Contains(t, providerErrors, "initialization")
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

func TestPinnedSDKSupportsNamedAndStickyPercentageTargeting(t *testing.T) {
	t.Run("named UUID", func(t *testing.T) {
		targetedID := "4c47b265-6987-4bc1-8933-006168a31793"
		client := newSnapshotClient(t, booleanSnapshot([]string{fmt.Sprintf(`{
                    "type": "SEGMENT_ROLLOUT_TYPE",
                    "rank": 1,
                    "segment": {
                        "value": true,
                        "segmentOperator": "OR_SEGMENT_OPERATOR",
                        "segments": [{
                                "key": "named-maintainer",
                                "matchType": "ANY_SEGMENT_MATCH_TYPE",
                                "constraints": [{
                                    "type": "ENTITY_ID_CONSTRAINT_COMPARISON_TYPE",
                                    "property": "entity",
                                    "operator": "eq",
                                    "value": %q
                                }]
                            }]
                    }
                }`, targetedID)}))

		targeted, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
			FlagKey: "release-log-entry-v2", EntityID: targetedID,
		})
		require.NoError(t, err)
		other, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
			FlagKey: "release-log-entry-v2", EntityID: "978dc423-2868-481f-b30d-4a88cf791903",
		})
		require.NoError(t, err)
		assert.True(t, targeted.Enabled)
		assert.False(t, other.Enabled)
	})

	t.Run("sticky percentage", func(t *testing.T) {
		client := newSnapshotClient(t, booleanSnapshot([]string{`{
                    "type": "THRESHOLD_ROLLOUT_TYPE",
                    "rank": 1,
                    "threshold": {"percentage": 50.0, "value": true}
                }`}))
		seen := map[bool]bool{}
		for i := range 100 {
			entityID := fmt.Sprintf("00000000-0000-4000-8000-%012d", i)
			first, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
				FlagKey: "release-log-entry-v2", EntityID: entityID,
			})
			require.NoError(t, err)
			seen[first.Enabled] = true
			for range 3 {
				repeated, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
					FlagKey: "release-log-entry-v2", EntityID: entityID,
				})
				require.NoError(t, err)
				assert.Equal(t, first.Enabled, repeated.Enabled)
			}
		}
		assert.Equal(t, map[bool]bool{false: true, true: true}, seen)
	})
}

func newSnapshotClient(t *testing.T, snapshot string) *Client {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("ETag", `"fixture"`)
		_, _ = response.Write([]byte(snapshot))
	}))
	t.Cleanup(server.Close)
	client, err := New(context.Background(), Config{
		URL:            server.URL,
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Minute,
		RequestTimeout: time.Second,
		StartupTimeout: time.Second,
	}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })
	return client
}

func booleanSnapshot(rollouts []string) string {
	if rollouts == nil {
		rollouts = []string{}
	}
	return fmt.Sprintf(`{
        "namespace": {"key": "default"},
        "flags": [{
			"key": "release-log-entry-v2",
			"name": "",
			"description": "",
			"enabled": false,
			"type": "BOOLEAN_FLAG_TYPE",
			"createdAt": "2026-08-24T06:01:42Z",
			"updatedAt": "2026-08-24T06:01:42Z",
			"rules": [],
			"rollouts": [%s]
		}],
		"digest": "f86fe0149bc542c8b350e994b869ee5370411c12"
    }`, strings.Join(rollouts, ","))
}

func TestConfigRequiresExplicitSafeBounds(t *testing.T) {
	valid := Config{
		URL:            "http://oathkeeper-proxy.default:4455/flipt",
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: 30 * time.Second,
		RequestTimeout: 5 * time.Second,
		StartupTimeout: 3 * time.Second,
	}
	require.NoError(t, valid.validate())

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "relative URL", mutate: func(cfg *Config) { cfg.URL = "/flipt" }},
		{name: "environment", mutate: func(cfg *Config) { cfg.Environment = "" }},
		{name: "namespace", mutate: func(cfg *Config) { cfg.Namespace = "" }},
		{name: "update interval", mutate: func(cfg *Config) { cfg.UpdateInterval = time.Millisecond }},
		{name: "request timeout", mutate: func(cfg *Config) { cfg.RequestTimeout = time.Millisecond }},
		{name: "startup timeout", mutate: func(cfg *Config) { cfg.StartupTimeout = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.mutate(&cfg)
			assert.Error(t, cfg.validate())
		})
	}
}

func TestNewReturnsAtStartupBudgetAndInstallsClientAfterRecovery(t *testing.T) {
	release := make(chan struct{})
	sdk := &fakeSDKClient{
		response: &fliptsdk.BooleanEvaluationResponse{
			Enabled: true,
			FlagKey: "release-log-entry-v2",
			Reason:  "MATCH_EVALUATION_REASON",
		},
	}
	observer := &recordingObserver{}
	started := time.Now()
	client, err := newClient(context.Background(), Config{
		URL:            "http://flipt.test/flipt",
		Environment:    "local",
		Namespace:      "default",
		UpdateInterval: time.Minute,
		RequestTimeout: time.Second,
		StartupTimeout: 10 * time.Millisecond,
	}, observer, func(context.Context, ...fliptsdk.Option) (sdkClient, error) {
		<-release
		return sdk, nil
	})

	require.NoError(t, err)
	assert.Less(t, time.Since(started), time.Second)
	_, err = client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "release-log-entry-v2"})
	assert.Error(t, err)

	close(release)
	require.Eventually(t, func() bool {
		result, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "release-log-entry-v2"})
		return err == nil && result.Enabled
	}, time.Second, 10*time.Millisecond)
	statuses, _, _ := observer.snapshot()
	assert.Equal(t, []featureflags.InitializationStatus{
		featureflags.InitializationStatusFallback,
		featureflags.InitializationStatusReady,
	}, statuses)
	require.NoError(t, client.Close(context.Background()))
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
	statuses := []int{http.StatusOK, http.StatusNotModified}
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

	_, refreshes, providerErrors := observer.snapshot()
	assert.Equal(t, 1, refreshes)
	assert.Empty(t, providerErrors)
	assert.False(t, state.isStale(false))
	assert.True(t, state.canAcceptNotModified(`"recovered"`))
}

func TestFetchStateOnlyAcceptsNotModifiedForConfirmedETag(t *testing.T) {
	state := newProviderFetchState("default", nil)
	state.snapshotReceived(`"v1"`)
	state.confirmPendingSnapshot()
	state.failed()

	assert.False(t, state.canAcceptNotModified(`"uninstalled-v2"`))
	assert.True(t, state.canAcceptNotModified(`"v1"`))
}

func TestCloseReleasesSDK(t *testing.T) {
	sdk := &fakeSDKClient{}
	client := &Client{client: sdk}
	require.NoError(t, client.Close(context.Background()))
	assert.True(t, sdk.closed)
}
