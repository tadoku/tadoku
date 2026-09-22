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

// EvaluatePublicForSubject evaluates public decisions for a verified subject.
// An empty or guest subject keeps anonymous safe defaults.
func (e *Evaluator) EvaluatePublicForSubject(ctx context.Context, subject string) PublicDecisions {
	if subject == "" || subject == "guest" {
		return e.EvaluatePublic(ctx, nil)
	}
	return e.EvaluatePublic(ctx, &commondomain.UserIdentity{Subject: subject})
}

// EvaluatePublic evaluates only decisions approved for public API responses.
func (e *Evaluator) EvaluatePublic(ctx context.Context, user *commondomain.UserIdentity) PublicDecisions {
	return PublicDecisions{
		ReleaseLogEntryV2: e.Boolean(ctx, ReleaseLogEntryV2, user),
	}
}
