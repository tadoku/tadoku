package featureflags

import (
	"context"
	"errors"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

func TestTenantEvaluation(t *testing.T) {
	evaluator := NewEvaluator(nil, nil)
	tenantEvaluator, ok := any(evaluator).(interface {
		TenantBoolean(context.Context, BooleanFlag) bool
	})
	if !ok {
		t.Fatal("evaluator has no tenant evaluation boundary")
	}
	var flag BooleanFlag
	for i := range 255 {
		if candidate := BooleanFlag(i); candidate.Key() == "identity-v2" {
			flag = candidate
			break
		}
	}
	if flag.Key() != "identity-v2" {
		t.Fatal("identity-v2 is not registered")
	}
	branch, err := tenant.Parse("e2e/identity-skeleton")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []tenant.Key{tenant.Production(), branch} {
		if tenantEvaluator.TenantBoolean(tenant.WithKey(t.Context(), key), flag) {
			t.Fatalf("unset flag enabled for %s", key)
		}
	}
	provider := &tenantDecisionProvider{enabled: map[tenant.Key]bool{branch: true}}
	evaluator.provider = provider
	for _, key := range []tenant.Key{branch, tenant.Production(), branch, tenant.Production()} {
		if got := tenantEvaluator.TenantBoolean(tenant.WithKey(t.Context(), key), flag); got != provider.enabled[key] {
			t.Errorf("tenant %s enabled=%t", key, got)
		}
	}
	if tenantEvaluator.TenantBoolean(t.Context(), flag) {
		t.Fatal("missing tenant enabled")
	}
	if provider.calls != 4 {
		t.Fatalf("provider calls=%d want 4", provider.calls)
	}
	provider.err = errors.New("outage")
	if tenantEvaluator.TenantBoolean(tenant.WithKey(t.Context(), branch), flag) {
		t.Fatal("outage did not use false")
	}
	provider.err = ErrFlagNotFound
	if tenantEvaluator.TenantBoolean(tenant.WithKey(t.Context(), branch), flag) {
		t.Fatal("missing flag did not use false")
	}
}

type tenantDecisionProvider struct {
	enabled map[tenant.Key]bool
	calls   int
	err     error
}

func (p *tenantDecisionProvider) EvaluateBoolean(
	ctx context.Context,
	request EvaluationRequest) (ProviderResult,
	error,
) {
	p.calls++
	key, ok := tenant.FromContext(ctx)
	if !ok || request.EntityID != key.String() || request.FlagKey != "identity-v2" {
		return ProviderResult{}, errors.New("wrong tenant or entity")
	}
	return ProviderResult{Enabled: p.enabled[key], Reason: "match"}, p.err
}
