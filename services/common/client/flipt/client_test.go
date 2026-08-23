package flipt

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
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
					FlagKey: "release.log-entry-v2",
					Reason:  "MATCH_EVALUATION_REASON",
				},
				stateErr: tt.stateErr,
			}
			client := &Client{client: sdk}
			request := featureflags.EvaluationRequest{
				FlagKey:  "release.log-entry-v2",
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
			_, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "release.log-entry-v2"})
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
	sdk := &fakeSDKClient{response: &fliptsdk.BooleanEvaluationResponse{FlagKey: "release.log-entry-v2"}}
	client := &Client{client: sdk}

	_, err := client.EvaluateBoolean(ctx, featureflags.EvaluationRequest{FlagKey: "release.log-entry-v2"})

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
	_, err = client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "release.log-entry-v2", EntityID: "4c47b265-6987-4bc1-8933-006168a31793"})
	assert.ErrorIs(t, err, featureflags.ErrFlagNotFound)
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
			FlagKey: "release.log-entry-v2",
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
	_, err = client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "release.log-entry-v2"})
	assert.Error(t, err)

	close(release)
	require.Eventually(t, func() bool {
		result, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "release.log-entry-v2"})
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
		observer: observer,
		base: roundTripperFunc(func(_ *http.Request) (*http.Response, error) {
			status := statuses[0]
			statuses = statuses[1:]
			return &http.Response{StatusCode: status, Body: http.NoBody}, nil
		}),
	}
	request := httptest.NewRequest(http.MethodGet, "http://flipt/snapshot", nil)

	for range 3 {
		_, err := transport.RoundTrip(request)
		require.NoError(t, err)
	}

	_, refreshes, providerErrors := observer.snapshot()
	assert.Equal(t, 2, refreshes)
	assert.Equal(t, []string{"fetch"}, providerErrors)
}

func TestCloseReleasesSDK(t *testing.T) {
	sdk := &fakeSDKClient{}
	client := &Client{client: sdk}
	require.NoError(t, client.Close(context.Background()))
	assert.True(t, sdk.closed)
}
