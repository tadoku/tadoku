package authz

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

// TestCheckProperties checks, for random graphs over the reference model, that AllowedFor equals the reference
// evaluator for every question, including the depth error.
func TestCheckProperties(t *testing.T) {
	t.Parallel()

	db := newTestDatabase(t)
	repository := NewRepository(db.AppPool)
	writer := permissions.NewRelationships(repository, refModel)
	checker := permissions.NewChecker(repository, refModel)
	graphs := newGraphPool()
	questions := refQuestions(graphs)
	seen := map[refVerdict]int{}

	for _, seed := range propertySeeds {
		key := fmt.Sprintf("e2e/check-properties-%d", seed)
		ctx := registerTestTenant(t, db, key)
		r := rand.New(rand.NewPCG(seed, 1))
		graph := graphs.randomReferenceGraph(r, r.IntN(31))
		grantAll(t, ctx, writer, graph...)
		ref := newReference(graph)

		for _, question := range questions {
			for _, user := range graphs.users {
				allowed, err := checker.AllowedFor(ctx, question.check, user)

				got, want := checkVerdict(allowed, err), ref.check(question.object, question.permit, user)
				if got != want {
					propertyFailure(
						t,
						seed,
						graph,
						"AllowedFor(%s:%s#%s, %s) = %t, %v; reference %s",
						question.object.Type(),
						question.object.ID(),
						question.permit,
						user,
						allowed,
						err,
						want,
					)
				}
				seen[want]++
			}
		}
	}

	for _, verdict := range []refVerdict{refDenied, refGranted, refFailed} {
		if seen[verdict] == 0 {
			t.Errorf("no generated question was %s; the generator no longer covers it", verdict)
		}
	}
	t.Logf("verdicts over %d seeds: %v", len(propertySeeds), seen)
}

// checkVerdict maps a check result to the reference's verdicts. An error other than the depth error matches none.
func checkVerdict(allowed bool, err error) refVerdict {
	switch {
	case err != nil && strings.Contains(err.Error(), maxDepthError):
		return refFailed
	case err != nil:
		return -1
	case allowed:
		return refGranted
	default:
		return refDenied
	}
}
