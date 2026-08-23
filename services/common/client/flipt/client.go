package flipt

import (
	"context"
	"errors"
	"fmt"
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
	mu     sync.RWMutex
	client sdkClient
	cancel context.CancelFunc
	closed bool
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

	httpClient := withObservability(cfg.HTTPClient, observer)
	lifetimeCtx, cancel := context.WithCancel(ctx)
	client := &Client{cancel: cancel}
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
	status := featureflags.InitializationStatusReady
	if client.Err() != nil {
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

	return featureflags.ProviderResult{
		Enabled: response.Enabled,
		Reason:  response.Reason,
		Stale:   client.Err() != nil,
	}, nil
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

func withObservability(client *http.Client, observer Observer) *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	cloned := *client
	base := cloned.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	cloned.Transport = &observedTransport{base: base, observer: observer}
	return &cloned
}

type observedTransport struct {
	base     http.RoundTripper
	observer Observer
}

func (t *observedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(request)
	if err != nil {
		if t.observer != nil {
			t.observer.ObserveProviderError("fetch")
		}
		return response, err
	}
	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusNotModified {
		if t.observer != nil {
			t.observer.ObserveConfigRefresh()
		}
	} else if t.observer != nil {
		t.observer.ObserveProviderError("fetch")
	}
	return response, nil
}
