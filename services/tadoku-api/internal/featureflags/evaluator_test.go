package featureflags

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeProvider struct {
	result   ProviderResult
	err      error
	evaluate func(EvaluationRequest) (ProviderResult, error)
	requests []EvaluationRequest
}

func (p *fakeProvider) EvaluateBoolean(_ context.Context, request EvaluationRequest) (ProviderResult, error) {
	p.requests = append(p.requests, request)
	if p.evaluate != nil {
		return p.evaluate(request)
	}
	return p.result, p.err
}

type recordingObserver struct {
	observations []Observation
}

func (o *recordingObserver) ObserveEvaluation(observation Observation) {
	o.observations = append(o.observations, observation)
}

func TestRegistryOwnsPilotKeyAndSafeDefault(t *testing.T) {
	assert.Equal(t, "release-log-entry-v2", ReleaseLogEntryV2.Key())
	assert.False(t, ReleaseLogEntryV2.SafeDefault())
}

func TestEvaluatorUsesOnlyStableSubjectAndAllowlistedContext(t *testing.T) {
	provider := &fakeProvider{result: ProviderResult{Enabled: true, Reason: "MATCH_EVALUATION_REASON"}}
	observer := &recordingObserver{}
	evaluator := NewEvaluator(provider, observer)
	subject := uuid.NewString()

	enabled := evaluator.Boolean(context.Background(), ReleaseLogEntryV2, subject)

	assert.True(t, enabled)
	require.Len(t, provider.requests, 1)
	assert.Equal(t, "release-log-entry-v2", provider.requests[0].FlagKey)
	assert.Equal(t, subject, provider.requests[0].EntityID)
	assert.Equal(t, map[string]string{"authenticated": "true"}, provider.requests[0].Context)
	assert.NotContains(t, provider.requests[0].Context, "email")
	assert.NotContains(t, provider.requests[0].Context, "display_name")
	require.Len(t, observer.observations, 1)
	assert.Equal(t, EvaluationReasonMatch, observer.observations[0].Reason)
	assert.Equal(t, EvaluationSourceProvider, observer.observations[0].Source)
}

func TestEvaluatorPreservesNamedTargetingAndStickyEntity(t *testing.T) {
	targetedSubject := uuid.NewString()
	otherSubject := uuid.NewString()
	provider := &fakeProvider{evaluate: func(request EvaluationRequest) (ProviderResult, error) {
		return ProviderResult{
			Enabled: request.EntityID == targetedSubject,
			Reason:  "match",
		}, nil
	}}
	evaluator := NewEvaluator(provider, nil)
	targetedUser := targetedSubject

	for range 3 {
		assert.True(t, evaluator.Boolean(context.Background(), ReleaseLogEntryV2, targetedUser))
	}
	assert.False(t, evaluator.Boolean(context.Background(), ReleaseLogEntryV2, otherSubject))

	require.Len(t, provider.requests, 4)
	for _, request := range provider.requests[:3] {
		assert.Equal(t, targetedSubject, request.EntityID)
	}
	assert.Equal(t, otherSubject, provider.requests[3].EntityID)
}

func TestEvaluatorUsesRegistryDefaultForUnsafeStates(t *testing.T) {
	tests := []struct {
		name       string
		provider   BooleanProvider
		ctx        func() context.Context
		subject    string
		flag       BooleanFlag
		wantReason EvaluationReason
		wantError  bool
	}{
		{name: "anonymous cold start", provider: &fakeProvider{}, ctx: context.Background, flag: ReleaseLogEntryV2, wantReason: EvaluationReasonAnonymous},
		{name: "invalid identity", provider: &fakeProvider{}, ctx: context.Background, subject: "not-a-uuid", flag: ReleaseLogEntryV2, wantReason: EvaluationReasonInvalidIdentity, wantError: true},
		{name: "nil UUID identity", provider: &fakeProvider{}, ctx: context.Background, subject: uuid.Nil.String(), flag: ReleaseLogEntryV2, wantReason: EvaluationReasonInvalidIdentity, wantError: true},
		{name: "provider not initialized", ctx: context.Background, subject: uuid.NewString(), flag: ReleaseLogEntryV2, wantReason: EvaluationReasonProviderError, wantError: true},
		{name: "empty provider missing flag", provider: &fakeProvider{err: ErrFlagNotFound}, ctx: context.Background, subject: uuid.NewString(), flag: ReleaseLogEntryV2, wantReason: EvaluationReasonNotFound},
		{name: "malformed provider response", provider: &fakeProvider{err: ErrInvalidResponse}, ctx: context.Background, subject: uuid.NewString(), flag: ReleaseLogEntryV2, wantReason: EvaluationReasonInvalidResponse, wantError: true},
		{name: "provider outage", provider: &fakeProvider{err: errors.New("connection refused")}, ctx: context.Background, subject: uuid.NewString(), flag: ReleaseLogEntryV2, wantReason: EvaluationReasonProviderError, wantError: true},
		{name: "provider deadline", provider: &fakeProvider{err: context.DeadlineExceeded}, ctx: context.Background, subject: uuid.NewString(), flag: ReleaseLogEntryV2, wantReason: EvaluationReasonCanceled, wantError: true},
		{name: "unknown flag", provider: &fakeProvider{result: ProviderResult{Enabled: true}}, ctx: context.Background, subject: uuid.NewString(), flag: BooleanFlag(200), wantReason: EvaluationReasonInvalidFlag, wantError: true},
		{name: "canceled", provider: &fakeProvider{result: ProviderResult{Enabled: true}}, ctx: canceledContext, subject: uuid.NewString(), flag: ReleaseLogEntryV2, wantReason: EvaluationReasonCanceled, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observer := &recordingObserver{}
			evaluator := NewEvaluator(tt.provider, observer)

			assert.False(t, evaluator.Boolean(tt.ctx(), tt.flag, tt.subject))
			require.Len(t, observer.observations, 1)
			assert.Equal(t, tt.wantReason, observer.observations[0].Reason)
			assert.Equal(t, EvaluationSourceDefault, observer.observations[0].Source)
			assert.Equal(t, tt.wantError, observer.observations[0].Err)
		})
	}
}

func TestEvaluatorMarksLastKnownGoodDecisionAsStale(t *testing.T) {
	provider := &fakeProvider{result: ProviderResult{Enabled: true, Reason: "match", Stale: true}}
	observer := &recordingObserver{}
	evaluator := NewEvaluator(provider, observer)

	assert.True(t, evaluator.Boolean(context.Background(), ReleaseLogEntryV2, uuid.NewString()))
	require.Len(t, observer.observations, 1)
	assert.Equal(t, EvaluationSourceStaleCache, observer.observations[0].Source)
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
