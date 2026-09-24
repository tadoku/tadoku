package featureflags

import "context"

// PublicDecisions is the explicit allowlist of feature flag decisions that
// may be exposed through an API transport.
type PublicDecisions struct {
	ReleaseLogEntryV2 bool
}

// EvaluatePublicForSubject evaluates public decisions for a verified subject.
// An empty or guest subject keeps anonymous safe defaults.
func (e *Evaluator) EvaluatePublicForSubject(ctx context.Context, subject string) PublicDecisions {
	return PublicDecisions{
		ReleaseLogEntryV2: e.Boolean(ctx, ReleaseLogEntryV2, subject),
	}
}
