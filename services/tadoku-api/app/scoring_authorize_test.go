package app

import (
	"context"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

func TestAuthorizeScoringRuleSetChangeReportsPersistedInvariantFailuresAsInternal(t *testing.T) {
	t.Parallel()

	application := &Application{}
	for _, test := range []struct {
		name    string
		ruleSet ScoringRuleSet
		want    string
	}{
		{
			name:    "contest scope without contest id",
			ruleSet: ScoringRuleSet{Scope: "contest"},
			want:    "contest scoring rule set requires contest_id",
		},
		{
			name:    "unknown scope",
			ruleSet: ScoringRuleSet{Scope: "unknown"},
			want:    "scoring rule set scope is invalid",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := application.authorizeScoringRuleSetChange(context.Background(), test.ruleSet)
			if errx.KindOf(err) != errx.Internal || err.Error() != test.want {
				t.Errorf("error=%v, want internal %q", err, test.want)
			}
		})
	}
}
