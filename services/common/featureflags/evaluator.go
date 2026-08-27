package featureflags

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
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

// BooleanProvider is the vendor-neutral boundary used by the typed evaluator.
// Vendor SDK types stay in the client/infrastructure package.
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

// Evaluator resolves typed flags and always falls back to the registry-owned
// safe default. It never returns provider failures to product code.
type Evaluator struct {
	provider BooleanProvider
	observer Observer
	clock    commondomain.Clock
}

func NewEvaluator(provider BooleanProvider, observer Observer, clock commondomain.Clock) *Evaluator {
	return &Evaluator{provider: provider, observer: observer, clock: clock}
}

func (e *Evaluator) Boolean(ctx context.Context, flag BooleanFlag, user *commondomain.UserIdentity) bool {
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
	if user == nil || user.Subject == "" || user.Subject == "guest" {
		return defaultResult(EvaluationReasonAnonymous, false)
	}
	subject, err := uuid.Parse(user.Subject)
	if err != nil || subject == uuid.Nil {
		return defaultResult(EvaluationReasonInvalidIdentity, true)
	}
	if e.provider == nil {
		return defaultResult(EvaluationReasonProviderError, true)
	}

	result, err := e.provider.EvaluateBoolean(ctx, EvaluationRequest{
		FlagKey:  definition.key,
		EntityID: subject.String(),
		Context: map[string]string{
			"authenticated": "true",
		},
	})
	if err != nil {
		if errors.Is(err, ErrFlagNotFound) {
			return defaultResult(EvaluationReasonNotFound, false)
		}
		if errors.Is(err, ErrInvalidResponse) {
			return defaultResult(EvaluationReasonInvalidResponse, true)
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return defaultResult(EvaluationReasonCanceled, true)
		}
		return defaultResult(EvaluationReasonProviderError, true)
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
	return result.Enabled
}

func (e *Evaluator) now() time.Time {
	if e.clock == nil {
		return time.Time{}
	}
	return e.clock.Now()
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
