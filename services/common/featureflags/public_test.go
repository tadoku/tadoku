package featureflags

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

func TestEvaluatePublicReturnsAllowlistedDecisionForAuthenticatedUser(t *testing.T) {
	provider := &fakeProvider{result: ProviderResult{Enabled: true, Reason: "match"}}
	evaluator := NewEvaluator(provider, nil, commondomain.NewMockClock(time.Time{}))
	user := &commondomain.UserIdentity{Subject: uuid.NewString()}

	decisions := evaluator.EvaluatePublic(context.Background(), user)

	assert.True(t, decisions.ReleaseLogEntryV2)
	require.Len(t, provider.requests, 1)
	assert.Equal(t, ReleaseLogEntryV2.Key(), provider.requests[0].FlagKey)
}

func TestEvaluatePublicTreatsGuestAsAnonymous(t *testing.T) {
	provider := &fakeProvider{result: ProviderResult{Enabled: true, Reason: "match"}}
	observer := &recordingObserver{}
	evaluator := NewEvaluator(provider, observer, commondomain.NewMockClock(time.Time{}))

	decisions := evaluator.EvaluatePublic(context.Background(), &commondomain.UserIdentity{Subject: "guest"})

	assert.False(t, decisions.ReleaseLogEntryV2)
	assert.Empty(t, provider.requests)
	require.Len(t, observer.observations, 1)
	assert.Equal(t, EvaluationReasonAnonymous, observer.observations[0].Reason)
	assert.False(t, observer.observations[0].Err)
}

func TestEvaluatePublicUsesSafeDefaultWhenProviderFails(t *testing.T) {
	provider := &fakeProvider{err: errors.New("provider unavailable")}
	evaluator := NewEvaluator(provider, nil, commondomain.NewMockClock(time.Time{}))

	decisions := evaluator.EvaluatePublic(context.Background(), &commondomain.UserIdentity{Subject: uuid.NewString()})

	assert.False(t, decisions.ReleaseLogEntryV2)
}
