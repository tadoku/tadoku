package flipt

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	fliptsdk "go.flipt.io/flipt-client"
)

func TestProviderRoutesAndBoundsTenantClients(t *testing.T) {
	targets, err := NewTargets("production", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	observer := &recordingObserver{}
	clients := make(map[Target]*fakeSDKClient)
	factory := func(_ context.Context, cfg Config, observed Observer) (*Client, error) {
		target := Target{Environment: cfg.Environment, Namespace: cfg.Namespace}
		if target.Environment == "production" && observed != observer {
			t.Error("canonical provider has no lifecycle observer")
		}
		if target.Environment == "test" && observed != nil {
			t.Error("test provider changed canonical lifecycle metrics")
		}
		sdk := &fakeSDKClient{response: &fliptsdk.BooleanEvaluationResponse{
			FlagKey: "release-log-entry-v2",
			Enabled: target.Environment == "test",
		}}
		clients[target] = sdk
		return &Client{client: sdk}, nil
	}

	provider, err := newProvider(t.Context(), Config{}, targets, observer, factory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := provider.Close(context.Background()); err != nil {
			t.Error(err)
		}
	})
	canonical := Target{Environment: "production", Namespace: "default"}
	if len(clients) != 1 || clients[canonical] == nil {
		t.Fatal("canonical client was not created at startup")
	}

	request := featureflags.EvaluationRequest{FlagKey: "release-log-entry-v2", EntityID: "11111111-1111-4111-8111-111111111111"}
	if _, err := provider.EvaluateBoolean(t.Context(), request); err == nil || clients[canonical].request != nil {
		t.Fatal("missing tenant evaluated production")
	}
	for index := range 16 {
		key, err := tenant.Parse(fmt.Sprintf("e2e/flipt-%d", index))
		if err != nil {
			t.Fatal(err)
		}
		result, err := provider.EvaluateBoolean(tenant.WithKey(t.Context(), key), request)
		if err != nil || !result.Enabled {
			t.Fatalf("test tenant %d: result=%+v error=%v", index, result, err)
		}
	}
	if clients[canonical].request != nil {
		t.Fatal("test evaluations reached canonical client")
	}

	key, _ := tenant.Parse("e2e/flipt-0")
	if _, err := provider.EvaluateBoolean(tenant.WithKey(t.Context(), key), request); err != nil {
		t.Fatal(err)
	}
	key, _ = tenant.Parse("e2e/flipt-16")
	if _, err := provider.EvaluateBoolean(tenant.WithKey(t.Context(), key), request); err != nil {
		t.Fatal(err)
	}
	for index := range 17 {
		sdk := clients[Target{Environment: "test", Namespace: fmt.Sprintf("e2e_flipt-%d", index)}]
		if sdk == nil || sdk.closed != (index == 1) {
			t.Errorf("tenant %d missing=%t closed=%t", index, sdk == nil, sdk != nil && sdk.closed)
		}
	}

	result, err := provider.EvaluateBoolean(tenant.WithKey(t.Context(), tenant.Production()), request)
	if err != nil || result.Enabled || clients[canonical].closed {
		t.Fatalf("canonical evaluation result=%+v error=%v closed=%t", result, err, clients[canonical].closed)
	}
	if err := provider.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	for target, client := range clients {
		if !client.closed {
			t.Errorf("client %v remains open", target)
		}
	}
	if _, err := provider.EvaluateBoolean(tenant.WithKey(t.Context(), tenant.Production()), request); err == nil {
		t.Error("closed provider accepted evaluation")
	}
}

func TestTargetsRejectUnsafeEnvironmentAndKeepSlashKeysDistinct(t *testing.T) {
	for _, environment := range []string{"", " ", "production"} {
		if _, err := NewTargets("production", "default", environment); err == nil {
			t.Errorf("unsafe test environment %q accepted", environment)
		}
	}

	targets, err := NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[Target]bool)
	for _, raw := range []string{"a-b/c", "a/b-c", "e2e/run-1"} {
		key, err := tenant.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		target, err := targets.Resolve(tenant.WithKey(t.Context(), key))
		if err != nil || target.Environment != "test" || seen[target] {
			t.Fatalf("target %q = %+v error=%v duplicate=%t", raw, target, err, seen[target])
		}
		seen[target] = true
	}
}

func TestProviderCancelledEvaluationDoesNotCreateTenantClient(t *testing.T) {
	targets, err := NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	provider, err := newProvider(t.Context(), Config{RequestTimeout: time.Second}, targets, nil,
		func(context.Context, Config, Observer) (*Client, error) {
			calls++
			return &Client{client: &fakeSDKClient{}}, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close(context.Background()) })
	key, _ := tenant.Parse("e2e/cancelled")
	ctx, cancel := context.WithCancel(tenant.WithKey(t.Context(), key))
	cancel()
	_, err = provider.EvaluateBoolean(ctx, featureflags.EvaluationRequest{})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("cancelled evaluation error=%v factories=%d", err, calls)
	}
}
