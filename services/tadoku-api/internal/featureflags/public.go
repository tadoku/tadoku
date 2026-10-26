package featureflags

import "context"

type PublicDecisions struct {
	ReleaseWebV3 bool
}

func (e *Evaluator) EvaluatePublicForSubject(ctx context.Context, subject string) PublicDecisions {
	return PublicDecisions{
		ReleaseWebV3: e.Boolean(ctx, ReleaseWebV3, subject),
	}
}
