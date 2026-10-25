package featureflags

import "context"

type PublicDecisions struct {
	ReleaseLogEntryV2 bool
}

func (e *Evaluator) EvaluatePublicForSubject(ctx context.Context, subject string) PublicDecisions {
	return PublicDecisions{
		ReleaseLogEntryV2: e.Boolean(ctx, ReleaseLogEntryV2, subject),
	}
}
