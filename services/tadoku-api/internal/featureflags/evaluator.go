package featureflags

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type EvaluationRequest struct {
	FlagKey  string
	EntityID string
	Context  map[string]string
}

type ProviderResult struct {
	Enabled bool
	Reason  string
	Stale   bool
}

type BooleanProvider interface {
	EvaluateBoolean(ctx context.Context, request EvaluationRequest) (ProviderResult, error)
}

type EvaluationSource string

const (
	EvaluationSourceProvider   EvaluationSource = "provider"
	EvaluationSourceStaleCache EvaluationSource = "stale_cache"
	EvaluationSourceDefault    EvaluationSource = "safe_default"
)

type EvaluationReason string

const (
	EvaluationReasonMatch           EvaluationReason = "match"
	EvaluationReasonDefault         EvaluationReason = "default"
	EvaluationReasonDisabled        EvaluationReason = "disabled"
	EvaluationReasonOther           EvaluationReason = "other"
	EvaluationReasonAnonymous       EvaluationReason = "anonymous"
	EvaluationReasonInvalidIdentity EvaluationReason = "invalid_identity"
	EvaluationReasonInvalidFlag     EvaluationReason = "invalid_flag"
	EvaluationReasonInvalidResponse EvaluationReason = "invalid_response"
	EvaluationReasonNotFound        EvaluationReason = "not_found"
	EvaluationReasonCanceled        EvaluationReason = "canceled"
	EvaluationReasonProviderError   EvaluationReason = "provider_error"
)

var (
	ErrFlagNotFound    = errors.New("feature flag not found")
	ErrInvalidResponse = errors.New("invalid feature flag provider response")
)

type Observation struct {
	Flag     BooleanFlag
	Enabled  bool
	Reason   EvaluationReason
	Source   EvaluationSource
	Duration time.Duration
	Err      bool
}

type Observer interface {
	ObserveEvaluation(observation Observation)
}

type Evaluator struct {
	provider BooleanProvider
	observer Observer
	now      func() time.Time
}

func NewEvaluator(provider BooleanProvider, observer Observer) *Evaluator {
	return &Evaluator{provider: provider, observer: observer, now: time.Now}
}

func (e *Evaluator) Boolean(ctx context.Context, flag BooleanFlag, subject string) bool {
	definition := flag.definition()
	started := e.now()

	defaultResult := func(reason EvaluationReason, isError bool) bool {
		e.observe(Observation{
			Flag:     flag,
			Enabled:  definition.safeDefault,
			Reason:   reason,
			Source:   EvaluationSourceDefault,
			Duration: e.now().Sub(started),
			Err:      isError,
		})
		return definition.safeDefault
	}

	if definition.key == "" {
		return defaultResult(EvaluationReasonInvalidFlag, true)
	}
	if err := ctx.Err(); err != nil {
		return defaultResult(EvaluationReasonCanceled, true)
	}
	if subject == "" || subject == "guest" {
		return defaultResult(EvaluationReasonAnonymous, false)
	}
	userID, err := uuid.Parse(subject)
	if err != nil || userID == uuid.Nil {
		return defaultResult(EvaluationReasonInvalidIdentity, true)
	}
	if e.provider == nil {
		return defaultResult(EvaluationReasonProviderError, true)
	}

	enabled, err := e.evaluate(ctx, flag, EvaluationRequest{
		FlagKey:  definition.key,
		EntityID: userID.String(),
		Context:  map[string]string{"authenticated": "true"},
	}, started)
	if err != nil {
		return defaultResult(providerErrorReason(err))
	}
	return enabled
}

// TenantBoolean returns the context tenant's decision. Its error wraps ErrFlagNotFound when the provider's
// configuration lacks the flag; any other error means no decision was made, and false is not a decision.
func (e *Evaluator) TenantBoolean(ctx context.Context, flag BooleanFlag) (bool, error) {
	started := e.now()
	fail := func(reason EvaluationReason, err error) (bool, error) {
		e.observe(Observation{
			Flag:     flag,
			Enabled:  flag.SafeDefault(),
			Reason:   reason,
			Source:   EvaluationSourceDefault,
			Duration: e.now().Sub(started),
			Err:      reason != EvaluationReasonNotFound,
		})
		return false, err
	}

	if flag.Key() == "" {
		return fail(EvaluationReasonInvalidFlag, errors.New("feature flag is not declared"))
	}
	if err := ctx.Err(); err != nil {
		return fail(EvaluationReasonCanceled, err)
	}
	key, ok := tenant.FromContext(ctx)
	if !ok {
		return fail(EvaluationReasonInvalidIdentity, errors.New("feature flag evaluation requires a tenant"))
	}
	if e.provider == nil {
		return fail(EvaluationReasonProviderError, errors.New("feature flag provider is not configured"))
	}

	enabled, err := e.evaluate(ctx, flag, EvaluationRequest{FlagKey: flag.Key(), EntityID: key.String()}, started)
	if err != nil {
		reason, _ := providerErrorReason(err)
		return fail(reason, err)
	}
	return enabled, nil
}

func (e *Evaluator) evaluate(
	ctx context.Context,
	flag BooleanFlag,
	request EvaluationRequest,
	started time.Time,
) (bool, error) {
	result, err := e.provider.EvaluateBoolean(ctx, request)
	if err != nil {
		return false, err
	}

	source := EvaluationSourceProvider
	if result.Stale {
		source = EvaluationSourceStaleCache
	}
	e.observe(Observation{
		Flag:     flag,
		Enabled:  result.Enabled,
		Reason:   normalizeReason(result.Reason),
		Source:   source,
		Duration: e.now().Sub(started),
	})
	return result.Enabled, nil
}

func providerErrorReason(err error) (EvaluationReason, bool) {
	switch {
	case errors.Is(err, ErrFlagNotFound):
		return EvaluationReasonNotFound, false
	case errors.Is(err, ErrInvalidResponse):
		return EvaluationReasonInvalidResponse, true
	case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
		return EvaluationReasonCanceled, true
	default:
		return EvaluationReasonProviderError, true
	}
}

func (e *Evaluator) observe(observation Observation) {
	if e.observer != nil {
		e.observer.ObserveEvaluation(observation)
	}
}

func normalizeReason(reason string) EvaluationReason {
	switch reason {
	case "MATCH_EVALUATION_REASON", "match":
		return EvaluationReasonMatch
	case "DEFAULT_EVALUATION_REASON", "default":
		return EvaluationReasonDefault
	case "FLAG_DISABLED_EVALUATION_REASON", "disabled":
		return EvaluationReasonDisabled
	default:
		return EvaluationReasonOther
	}
}
