package featureflags

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluatePublicReturnsAllowlistedDecisionForAuthenticatedUser(t *testing.T) {
	provider := &fakeProvider{result: ProviderResult{Enabled: true, Reason: "match"}}
	evaluator := NewEvaluator(provider, nil)
	subject := uuid.NewString()

	decisions := evaluator.EvaluatePublicForSubject(context.Background(), subject)

	assert.True(t, decisions.ReleaseLogEntryV2)
	require.Len(t, provider.requests, 1)
	assert.Equal(t, ReleaseLogEntryV2.Key(), provider.requests[0].FlagKey)
}

func TestEvaluatePublicTreatsGuestAsAnonymous(t *testing.T) {
	provider := &fakeProvider{result: ProviderResult{Enabled: true, Reason: "match"}}
	observer := &recordingObserver{}
	evaluator := NewEvaluator(provider, observer)

	decisions := evaluator.EvaluatePublicForSubject(context.Background(), "guest")

	assert.False(t, decisions.ReleaseLogEntryV2)
	assert.Empty(t, provider.requests)
	require.Len(t, observer.observations, 1)
	assert.Equal(t, EvaluationReasonAnonymous, observer.observations[0].Reason)
	assert.False(t, observer.observations[0].Err)
}

func TestEvaluatePublicUsesSafeDefaultWhenProviderFails(t *testing.T) {
	provider := &fakeProvider{err: errors.New("provider unavailable")}
	evaluator := NewEvaluator(provider, nil)

	decisions := evaluator.EvaluatePublicForSubject(context.Background(), uuid.NewString())

	assert.False(t, decisions.ReleaseLogEntryV2)
}
