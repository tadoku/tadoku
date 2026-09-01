package flipt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	featureflags "github.com/tadoku/tadoku/services/common/featureflags"
	fliptsdk "go.flipt.io/flipt-client"
)

type Config struct {
	URL            string
	Environment    string
	Namespace      string
	UpdateInterval time.Duration
	RequestTimeout time.Duration
	StartupTimeout time.Duration
	HTTPClient     *http.Client
}

type Observer interface {
	ObserveInitialization(status featureflags.InitializationStatus)
	ObserveConfigRefresh()
	ObserveProviderError(kind string)
}

type sdkClient interface {
	EvaluateBoolean(ctx context.Context, request *fliptsdk.EvaluationRequest) (*fliptsdk.BooleanEvaluationResponse, error)
	Err() error
	Close(ctx context.Context) error
}

// Client adapts the concrete Flipt SDK to Tadoku's vendor-neutral evaluator.
type Client struct {
	mu         sync.RWMutex
	client     sdkClient
	fetchState *providerFetchState
	cancel     context.CancelFunc
	closed     bool
}

func New(ctx context.Context, cfg Config, observer Observer) (*Client, error) {
	return newClient(ctx, cfg, observer, func(ctx context.Context, options ...fliptsdk.Option) (sdkClient, error) {
		return fliptsdk.NewClient(ctx, options...)
	})
}

type sdkFactory func(context.Context, ...fliptsdk.Option) (sdkClient, error)

type initializationResult struct {
	client sdkClient
	err    error
}

func newClient(ctx context.Context, cfg Config, observer Observer, factory sdkFactory) (*Client, error) {
	if err := cfg.validate(); err != nil {
		observeInitialization(observer, featureflags.InitializationStatusError)
		return nil, err
	}

	fetchState := newProviderFetchState(cfg.Namespace, observer)
	httpClient := withObservability(cfg.HTTPClient, fetchState)
	lifetimeCtx, cancel := context.WithCancel(ctx)
	client := &Client{fetchState: fetchState, cancel: cancel}
	results := make(chan initializationResult, 1)
	go func() {
		sdk, err := factory(
			lifetimeCtx,
			fliptsdk.WithURL(cfg.URL),
			fliptsdk.WithEnvironment(cfg.Environment),
			fliptsdk.WithNamespace(cfg.Namespace),
			fliptsdk.WithUpdateInterval(cfg.UpdateInterval),
			fliptsdk.WithRequestTimeout(cfg.RequestTimeout),
			fliptsdk.WithFetchMode(fliptSdkFetchMode),
			fliptsdk.WithErrorStrategy(fliptSdkErrorStrategy),
			fliptsdk.WithHTTPClient(httpClient),
		)
		results <- initializationResult{client: sdk, err: err}
	}()

	timer := time.NewTimer(cfg.StartupTimeout)
	defer timer.Stop()
	select {
	case result := <-results:
		if result.err != nil {
			cancel()
			observeInitialization(observer, featureflags.InitializationStatusError)
			if observer != nil {
				observer.ObserveProviderError("initialization")
			}
			return nil, fmt.Errorf("initialize Flipt client: %w", result.err)
		}
		client.install(result.client, observer)
		return client, nil
	case <-timer.C:
		// Startup must not wait for Flipt. Keep initialization alive so the
		// process can move from safe defaults to normal polling after recovery.
		observeInitialization(observer, featureflags.InitializationStatusFallback)
		go func() {
			result := <-results
			if result.err != nil {
				cancel()
				observeInitialization(observer, featureflags.InitializationStatusError)
				if observer != nil {
					observer.ObserveProviderError("initialization")
				}
				return
			}
			client.install(result.client, observer)
		}()
		return client, nil
	case <-ctx.Done():
		cancel()
		return nil, ctx.Err()
	}
}

func (c *Client) install(client sdkClient, observer Observer) {
	if c.fetchState != nil && client.Err() == nil {
		c.fetchState.confirmPendingSnapshot()
	}
	status := featureflags.InitializationStatusReady
	if c.isStale(client) {
		status = featureflags.InitializationStatusFallback
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		_ = client.Close(context.Background())
		return
	}
	c.client = client
	c.mu.Unlock()
	observeInitialization(observer, status)
}

const (
	fliptSdkFetchMode     = fliptsdk.FetchModePolling
	fliptSdkErrorStrategy = fliptsdk.ErrorStrategyFallback
)

func (c *Client) EvaluateBoolean(ctx context.Context, request featureflags.EvaluationRequest) (featureflags.ProviderResult, error) {
	if err := ctx.Err(); err != nil {
		return featureflags.ProviderResult{}, err
	}
	if c == nil {
		return featureflags.ProviderResult{}, errors.New("Flipt client is not initialized")
	}
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	if client == nil {
		return featureflags.ProviderResult{}, errors.New("Flipt client is not initialized")
	}

	response, err := client.EvaluateBoolean(ctx, &fliptsdk.EvaluationRequest{
		FlagKey:  request.FlagKey,
		EntityID: request.EntityID,
		Context:  cloneContext(request.Context),
	})
	if err != nil {
		if isMissingFlag(err) {
			return featureflags.ProviderResult{}, featureflags.ErrFlagNotFound
		}
		return featureflags.ProviderResult{}, fmt.Errorf("evaluate Flipt boolean: %w", err)
	}
	if response == nil || response.FlagKey == "" || response.FlagKey != request.FlagKey {
		return featureflags.ProviderResult{}, featureflags.ErrInvalidResponse
	}
	if c.fetchState != nil && client.Err() == nil {
		c.fetchState.confirmPendingSnapshot()
	}

	return featureflags.ProviderResult{
		Enabled: response.Enabled,
		Reason:  response.Reason,
		Stale:   c.isStale(client),
	}, nil
}

func (c *Client) isStale(client sdkClient) bool {
	if c.fetchState != nil {
		return c.fetchState.isStale(client.Err() != nil)
	}
	return client.Err() != nil
}

func (c *Client) Close(ctx context.Context) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	client := c.client
	c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
	}
	if client == nil {
		return nil
	}
	return client.Close(ctx)
}

func (cfg Config) validate() error {
	parsedURL, err := url.ParseRequestURI(cfg.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("Flipt URL must be an absolute HTTP or HTTPS URL")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("Flipt URL must use HTTP or HTTPS")
	}
	if strings.TrimSpace(cfg.Environment) == "" {
		return errors.New("Flipt environment is required")
	}
	if strings.TrimSpace(cfg.Namespace) == "" {
		return errors.New("Flipt namespace is required")
	}
	if cfg.UpdateInterval < time.Second {
		return errors.New("Flipt update interval must be at least one second")
	}
	if cfg.RequestTimeout < time.Second {
		return errors.New("Flipt request timeout must be at least one second")
	}
	if cfg.StartupTimeout <= 0 {
		return errors.New("Flipt startup timeout must be positive")
	}
	return nil
}

func cloneContext(contextValues map[string]string) map[string]string {
	cloned := make(map[string]string, len(contextValues))
	for key, value := range contextValues {
		cloned[key] = value
	}
	return cloned
}

func isMissingFlag(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not found") ||
		strings.Contains(message, "no flag") ||
		strings.Contains(message, "failed to get flag information")
}

func observeInitialization(observer Observer, status featureflags.InitializationStatus) {
	if observer != nil {
		observer.ObserveInitialization(status)
	}
}

func withObservability(client *http.Client, state *providerFetchState) *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	cloned := *client
	base := cloned.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	cloned.Transport = &observedTransport{base: base, state: state}
	return &cloned
}

type observedTransport struct {
	base  http.RoundTripper
	state *providerFetchState
}

func (t *observedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(request)
	if err != nil {
		if t.state != nil {
			t.state.failed()
		}
		return response, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		body, readErr := io.ReadAll(response.Body)
		_ = response.Body.Close()
		response.Body = io.NopCloser(bytes.NewReader(body))
		if readErr != nil {
			if t.state != nil {
				t.state.failed()
			}
			return nil, fmt.Errorf("read Flipt snapshot response: %w", readErr)
		}
		if t.state != nil {
			if err := t.state.acceptSnapshot(body); err != nil {
				t.state.failed()
			} else {
				t.state.snapshotReceived(response.Header.Get("ETag"))
			}
		}
	case http.StatusNotModified:
		if t.state != nil {
			t.state.snapshotNotModified(request.Header.Get("If-None-Match"))
		}
	default:
		if t.state != nil {
			t.state.failed()
		}
	}
	return response, nil
}

type providerFetchState struct {
	mu            sync.RWMutex
	namespace     string
	observer      Observer
	observed      bool
	stale         bool
	hasSnapshot   bool
	pending       bool
	recovered     bool
	pendingETag   string
	confirmedETag string
}

func newProviderFetchState(namespace string, observer Observer) *providerFetchState {
	return &providerFetchState{namespace: namespace, observer: observer, stale: true}
}

func (s *providerFetchState) acceptSnapshot(payload []byte) error {
	var snapshot struct {
		Namespace *struct {
			Key string `json:"key"`
		} `json:"namespace"`
		Flags []struct {
			Key string `json:"key"`
		} `json:"flags"`
		Digest string `json:"digest"`
	}
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return fmt.Errorf("decode Flipt snapshot: %w", err)
	}
	if snapshot.Namespace == nil || snapshot.Namespace.Key != s.namespace {
		return errors.New("Flipt snapshot namespace does not match configuration")
	}
	if snapshot.Flags == nil {
		return errors.New("Flipt snapshot flags are missing")
	}
	if snapshot.Digest == "" {
		return errors.New("Flipt snapshot digest is missing")
	}
	for _, flag := range snapshot.Flags {
		if flag.Key == "" {
			return errors.New("Flipt snapshot contains a flag without a key")
		}
	}

	return nil
}

func (s *providerFetchState) snapshotReceived(etag string) {
	s.mu.Lock()
	s.observed = true
	s.pending = true
	s.recovered = false
	s.pendingETag = etag
	s.mu.Unlock()
}

func (s *providerFetchState) confirmPendingSnapshot() {
	s.mu.Lock()
	if !s.pending {
		s.mu.Unlock()
		return
	}
	s.observed = true
	s.stale = false
	s.hasSnapshot = true
	s.pending = false
	s.recovered = false
	s.confirmedETag = s.pendingETag
	s.pendingETag = ""
	s.mu.Unlock()
	if s.observer != nil {
		s.observer.ObserveConfigRefresh()
	}
}

func (s *providerFetchState) snapshotNotModified(etag string) {
	s.mu.Lock()
	accepted := etag != ""
	if s.pending {
		accepted = accepted && etag == s.pendingETag
		if accepted {
			s.hasSnapshot = true
			s.confirmedETag = s.pendingETag
		}
	} else {
		accepted = accepted && s.hasSnapshot && etag == s.confirmedETag
	}
	if !accepted {
		s.observed = true
		s.stale = true
		s.pending = false
		s.recovered = false
		s.pendingETag = ""
		s.mu.Unlock()
		if s.observer != nil {
			s.observer.ObserveProviderError("fetch")
		}
		return
	}

	wasStale := s.stale
	s.observed = true
	s.stale = false
	s.pending = false
	s.recovered = s.recovered || wasStale
	s.pendingETag = ""
	s.mu.Unlock()
	if s.observer != nil {
		s.observer.ObserveConfigRefresh()
	}
}

func (s *providerFetchState) failed() {
	s.mu.Lock()
	s.observed = true
	s.stale = true
	s.pending = false
	s.recovered = false
	s.pendingETag = ""
	s.mu.Unlock()
	if s.observer != nil {
		s.observer.ObserveProviderError("fetch")
	}
}

func (s *providerFetchState) isStale(sdkError bool) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.observed {
		return sdkError
	}
	if sdkError && !s.recovered {
		return true
	}
	return s.stale
}
