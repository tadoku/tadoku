package featureflags

import (
	"context"
	"errors"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestTenantEvaluation(t *testing.T) {
	evaluator := NewEvaluator(nil, nil)
	branch, err := tenant.Parse("e2e/identity-skeleton")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []tenant.Key{tenant.Production(), branch} {
		enabled, err := evaluator.TenantBoolean(tenant.WithKey(t.Context(), key), IdentityV2)
		if err == nil || errors.Is(err, ErrFlagNotFound) || enabled {
			t.Fatalf("missing provider for %s decided enabled=%t err=%v", key, enabled, err)
		}
	}

	provider := &tenantDecisionProvider{enabled: map[tenant.Key]bool{branch: true}}
	evaluator.provider = provider
	for _, key := range []tenant.Key{branch, tenant.Production(), branch, tenant.Production()} {
		enabled, err := evaluator.TenantBoolean(tenant.WithKey(t.Context(), key), IdentityV2)
		if err != nil || enabled != provider.enabled[key] {
			t.Errorf("tenant %s enabled=%t err=%v", key, enabled, err)
		}
	}
	if _, err := evaluator.TenantBoolean(t.Context(), IdentityV2); err == nil {
		t.Fatal("missing tenant was decided")
	}
	if provider.calls != 4 {
		t.Fatalf("provider calls=%d want 4", provider.calls)
	}

	for _, test := range []struct {
		err      error
		notFound bool
	}{
		{errors.New("outage"), false},
		{ErrInvalidResponse, false},
		{context.DeadlineExceeded, false},
		{ErrFlagNotFound, true},
	} {
		provider.err = test.err
		_, err := evaluator.TenantBoolean(tenant.WithKey(t.Context(), branch), IdentityV2)
		if !errors.Is(err, test.err) || errors.Is(err, ErrFlagNotFound) != test.notFound {
			t.Errorf("provider error %v returned %v", test.err, err)
		}
	}
}

type tenantDecisionProvider struct {
	enabled map[tenant.Key]bool
	calls   int
	err     error
}

func (p *tenantDecisionProvider) EvaluateBoolean(
	ctx context.Context,
	request EvaluationRequest,
) (ProviderResult, error) {
	p.calls++
	key, ok := tenant.FromContext(ctx)
	if !ok || request.EntityID != key.String() || request.FlagKey != "identity-v2" {
		return ProviderResult{}, errors.New("wrong tenant or entity")
	}
	return ProviderResult{Enabled: p.enabled[key], Reason: "match"}, p.err
}
