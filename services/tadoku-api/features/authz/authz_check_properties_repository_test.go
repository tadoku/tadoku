package authz

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
)

// TestCheckProperties checks, for random graphs over the reference model, that AllowedFor equals the reference
// evaluator for every question, including the depth error, and that AllowedEach and AllowedForEach equal the single
// checks for random batches without sending more queries than their slowest single check.
func TestCheckProperties(t *testing.T) {
	t.Parallel()

	db := newTestDatabase(t)
	pool, statements := newLoggedPool(t, db)
	repository := NewRepository(pool)
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

		singles := map[permissions.Check]map[uuid.UUID]single{}
		for _, question := range questions {
			singles[question.check] = map[uuid.UUID]single{}
			for _, user := range graphs.users {
				statements.take()

				allowed, err := checker.AllowedFor(ctx, question.check, user)

				queries := len(statements.take())
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
				singles[question.check][user] = single{verdict: want, queries: queries}
				seen[want]++
			}
		}

		for range 6 {
			question := questions[r.IntN(len(questions))]
			users := randomUsers(r, graphs.users, r.IntN(9))
			statements.take()

			allowed, err := checker.AllowedForEach(ctx, question.check, users)

			sent := len(statements.take())
			var pairs []pair
			for _, user := range users {
				pairs = append(pairs, pair{check: question.check, user: user})
			}
			got := map[pair]bool{}
			for user, passed := range allowed {
				got[pair{check: question.check, user: user}] = passed
			}
			if problem := batchProblem(singles, pairs, got, err, sent); problem != "" {
				propertyFailure(t, seed, graph, "AllowedForEach(%s:%s#%s, %d users): %s",
					question.object.Type(), question.object.ID(), question.permit, len(users), problem)
			}
		}

		for range 6 {
			user := graphs.users[r.IntN(len(graphs.users))]
			var checks []permissions.Check
			var pairs []pair
			for range r.IntN(2 * len(questions)) {
				check := questions[r.IntN(len(questions))].check
				checks = append(checks, check)
				pairs = append(pairs, pair{check: check, user: user})
			}
			statements.take()

			allowed, err := checker.AllowedEach(signedIn(ctx, user.String()), checks...)

			sent := len(statements.take())
			got := map[pair]bool{}
			for check, passed := range allowed {
				got[pair{check: check, user: user}] = passed
			}
			if problem := batchProblem(singles, pairs, got, err, sent); problem != "" {
				propertyFailure(t, seed, graph, "AllowedEach(%d checks) as %s: %s", len(checks), user, problem)
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

type single struct {
	verdict refVerdict
	queries int
}

type pair struct {
	check permissions.Check
	user  uuid.UUID
}

// batchProblem compares one batched call with the single checks of its pairs: it fails exactly when one of them
// fails, otherwise returns their answers, and sends no more queries than the slowest of them.
func batchProblem(
	singles map[permissions.Check]map[uuid.UUID]single,
	pairs []pair,
	got map[pair]bool,
	err error,
	sent int,
) string {
	fails, budget, want := false, 0, map[pair]bool{}
	for _, p := range pairs {
		answer := singles[p.check][p.user]
		fails = fails || answer.verdict == refFailed
		budget = max(budget, answer.queries)
		want[p] = answer.verdict == refGranted
	}

	switch {
	case fails != (err != nil && strings.Contains(err.Error(), maxDepthError)):
		return fmt.Sprintf("error = %v, want the depth error %t", err, fails)
	case sent > budget:
		return fmt.Sprintf("sent %d queries, more than the slowest single check's %d", sent, budget)
	case err == nil && !maps.Equal(got, want):
		return fmt.Sprintf("answers %v, single checks %v", got, want)
	}
	return ""
}
