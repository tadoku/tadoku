package flipt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	featureflags "github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	fliptsdk "go.flipt.io/flipt-client"
)

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
					FlagKey: "fixture-flag",
					Reason:  "MATCH_EVALUATION_REASON",
				},
				stateErr: tt.stateErr,
			}
			client := &Client{client: sdk}
			request := featureflags.EvaluationRequest{
				FlagKey:  "fixture-flag",
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
			_, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "fixture-flag"})
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
	sdk := &fakeSDKClient{response: &fliptsdk.BooleanEvaluationResponse{FlagKey: "fixture-flag"}}
	client := &Client{client: sdk}

	_, err := client.EvaluateBoolean(ctx, featureflags.EvaluationRequest{FlagKey: "fixture-flag"})

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
	require.Eventually(t, func() bool {
		_, _, providerErrors := observer.snapshot()
		return slices.Contains(providerErrors, "fetch")
	}, 10*time.Second, 10*time.Millisecond)
	statuses, _, _ := observer.snapshot()
	require.NotEmpty(t, statuses)
	for _, status := range statuses {
		assert.Equal(t, featureflags.InitializationStatusFallback, status)
	}
	_, err = client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "fixture-flag", EntityID: "4c47b265-6987-4bc1-8933-006168a31793"})
	require.Error(t, err)
	assert.NotErrorIs(t, err, featureflags.ErrFlagNotFound)
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
			FlagKey: "fixture-flag", EntityID: targetedID,
		})
		require.NoError(t, err)
		other, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
			FlagKey: "fixture-flag", EntityID: "978dc423-2868-481f-b30d-4a88cf791903",
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
				FlagKey: "fixture-flag", EntityID: entityID,
			})
			require.NoError(t, err)
			seen[first.Enabled] = true
			for range 3 {
				repeated, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{
					FlagKey: "fixture-flag", EntityID: entityID,
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
			FlagKey: "fixture-flag",
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
	}, observer, func(ctx context.Context, _ ...fliptsdk.Option) (sdkClient, error) {
		<-release
		ctx.Value(initialFetchKey{}).(*initialFetch).succeeded.Store(true)
		return sdk, nil
	})

	require.NoError(t, err)
	assert.Less(t, time.Since(started), time.Second)
	_, err = client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "fixture-flag"})
	assert.Error(t, err)

	close(release)
	require.Eventually(t, func() bool {
		result, err := client.EvaluateBoolean(context.Background(), featureflags.EvaluationRequest{FlagKey: "fixture-flag"})
		return err == nil && result.Enabled
	}, time.Second, 10*time.Millisecond)
	statuses, _, _ := observer.snapshot()
	assert.Equal(t, []featureflags.InitializationStatus{
		featureflags.InitializationStatusFallback,
		featureflags.InitializationStatusReady,
	}, statuses)
	require.NoError(t, client.Close(context.Background()))
}

func TestCloseReleasesSDK(t *testing.T) {
	sdk := &fakeSDKClient{}
	client := &Client{client: sdk}
	require.NoError(t, client.Close(context.Background()))
	assert.True(t, sdk.closed)
}

func TestSDKUsesMappedTenantNamespaceAndTestEnvironment(t *testing.T) {
	const namespace = "e2e_flipt-0123abcd"
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/internal/v1/evaluation/snapshot/namespace/"+namespace ||
			request.Header.Get("x-flipt-environment") != "test" {
			t.Errorf("SDK target path=%q environment=%q", request.URL.Path, request.Header.Get("x-flipt-environment"))
			http.NotFound(response, request)
			return
		}
		snapshot := strings.Replace(booleanSnapshot(nil), `"key": "default"`, `"key": "`+namespace+`"`, 1)
		snapshot = strings.Replace(snapshot, `"enabled": false`, `"enabled": true`, 1)
		_, _ = response.Write([]byte(snapshot))
	}))
	t.Cleanup(server.Close)

	client, err := New(t.Context(), Config{
		URL:            server.URL,
		Environment:    "test",
		Namespace:      namespace,
		UpdateInterval: time.Minute,
		RequestTimeout: time.Second,
		StartupTimeout: time.Second,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})

	result, err := client.EvaluateBoolean(t.Context(), featureflags.EvaluationRequest{
		FlagKey:  "fixture-flag",
		EntityID: "11111111-1111-4111-8111-111111111111",
	})
	if err != nil || !result.Enabled || result.Stale {
		t.Fatalf("mapped SDK evaluation result=%+v error=%v", result, err)
	}
}
