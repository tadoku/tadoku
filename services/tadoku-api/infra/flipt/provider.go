package flipt

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
)

// ponytail: Keep at most 16 test clients; increase only when active-tenant measurements require it.
const maxTestClients = 16

type clientFactory func(context.Context, Config, Observer) (*Client, error)

type tenantClient struct {
	target Target
	client *Client
}

type Provider struct {
	mu         sync.Mutex
	ctx        context.Context
	cfg        Config
	targets    Targets
	factory    clientFactory
	production *Client
	tests      map[Target]*list.Element
	recent     list.List
	closed     atomic.Bool
}

func NewProvider(ctx context.Context, cfg Config, targets Targets, observer Observer) (*Provider, error) {
	return newProvider(ctx, cfg, targets, observer, New)
}

func newProvider(
	ctx context.Context,
	cfg Config,
	targets Targets,
	observer Observer,
	factory clientFactory,
) (*Provider, error) {
	cfg.Environment = targets.production.Environment
	cfg.Namespace = targets.production.Namespace
	production, err := factory(ctx, cfg, observer)
	if err != nil {
		return nil, err
	}
	return &Provider{
		ctx:        ctx,
		cfg:        cfg,
		targets:    targets,
		factory:    factory,
		production: production,
		tests:      make(map[Target]*list.Element),
	}, nil
}

func (provider *Provider) EvaluateBoolean(
	ctx context.Context,
	request featureflags.EvaluationRequest,
) (featureflags.ProviderResult, error) {
	if err := ctx.Err(); err != nil {
		return featureflags.ProviderResult{}, err
	}
	if provider == nil {
		return featureflags.ProviderResult{}, errors.New("Flipt provider is not initialized")
	}
	target, err := provider.targets.Resolve(ctx)
	if err != nil {
		return featureflags.ProviderResult{}, err
	}

	if provider.closed.Load() {
		return featureflags.ProviderResult{}, errors.New("Flipt provider is closed")
	}
	if target == provider.targets.production {
		return provider.production.EvaluateBoolean(ctx, request)
	}

	provider.mu.Lock()
	defer provider.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return featureflags.ProviderResult{}, err
	}
	if provider.closed.Load() {
		return featureflags.ProviderResult{}, errors.New("Flipt provider is closed")
	}

	element := provider.tests[target]
	if element == nil {
		if len(provider.tests) == maxTestClients {
			oldest := provider.recent.Back()
			previous := oldest.Value.(tenantClient)
			delete(provider.tests, previous.target)
			provider.recent.Remove(oldest)
			if err := previous.client.Close(ctx); err != nil {
				return featureflags.ProviderResult{}, err
			}
		}
		cfg := provider.cfg
		cfg.Environment = target.Environment
		cfg.Namespace = target.Namespace
		client, err := provider.factory(provider.ctx, cfg, nil)
		if err != nil {
			return featureflags.ProviderResult{}, err
		}
		element = provider.recent.PushFront(tenantClient{target: target, client: client})
		provider.tests[target] = element
	} else {
		provider.recent.MoveToFront(element)
	}
	return element.Value.(tenantClient).client.EvaluateBoolean(ctx, request)
}

func (provider *Provider) Close(ctx context.Context) error {
	if provider == nil {
		return nil
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if provider.closed.Swap(true) {
		return nil
	}

	err := provider.production.Close(ctx)
	for _, element := range provider.tests {
		err = errors.Join(err, element.Value.(tenantClient).client.Close(ctx))
	}
	provider.tests = nil
	provider.recent.Init()
	return err
}
