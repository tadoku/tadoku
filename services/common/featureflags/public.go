package featureflags

import (
	"context"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

type PublicDecisions struct {
	ReleaseLogEntryV2 bool
}

// EvaluatePublicForSubject uses anonymous safe defaults for empty or guest subjects.
func (e *Evaluator) EvaluatePublicForSubject(ctx context.Context, subject string) PublicDecisions {
	if subject == "" || subject == "guest" {
		return e.EvaluatePublic(ctx, nil)
	}
	return e.EvaluatePublic(ctx, &commondomain.UserIdentity{Subject: subject})
}

func (e *Evaluator) EvaluatePublic(ctx context.Context, user *commondomain.UserIdentity) PublicDecisions {
	return PublicDecisions{
		ReleaseLogEntryV2: e.Boolean(ctx, ReleaseLogEntryV2, user),
	}
}
