package featureflags

import (
	"context"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

// PublicDecisions is the explicit allowlist of feature flag decisions that
// may be exposed through an API transport.
type PublicDecisions struct {
	ReleaseLogEntryV2 bool
}

// EvaluatePublic evaluates only decisions approved for public API responses.
func (e *Evaluator) EvaluatePublic(ctx context.Context, user *commondomain.UserIdentity) PublicDecisions {
	return PublicDecisions{
		ReleaseLogEntryV2: e.Boolean(ctx, ReleaseLogEntryV2, user),
	}
}
