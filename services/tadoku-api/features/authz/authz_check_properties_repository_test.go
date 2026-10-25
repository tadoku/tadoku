package authz

import (
	"errors"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"pgregory.net/rapid"
)

// TestCheckProperties checks, for random graphs over the reference model, that AllowedFor equals the reference
// evaluator for every drawn question and user, including the depth error. AllowedEach and AllowedForEach equal those
// single checks without sending more queries than the slowest of them, and are unavailable without answers when a
// storage call inside them fails.
func TestCheckProperties(t *testing.T) {
	t.Parallel()

	db := newTestDatabase(t)
	pool, statements := newLoggedPool(t, db)
	repository := NewRepository(pool)
	writer := permissions.NewRelationships(repository, refModel)
	checker := permissions.NewChecker(repository, refModel)
	production := tenant.Production().String()
	graphs := newGraphPool()
	graph := batchOf(graphs.referenceRelationship(), 30)
	questions := rapid.SliceOfN(rapid.SampledFrom(refQuestions(graphs)), 0, 12)
	users := rapid.SliceOfN(rapid.IntRange(0, len(graphs.users)-1), 0, 8)
	seen := map[refVerdict]int{}

	rapid.Check(t, func(t *rapid.T) {
		ctx := tenant.WithKey(t.Context(), tenant.Production())
		clearRelationships(t, db, production)
		relationships := graph.Draw(t, "graph")
		grantAll(t, ctx, writer, relationships...)
		ref := newReference(relationships)
		asked := questions.Draw(t, "questions")
		var askedUsers []uuid.UUID
		for _, i := range users.Draw(t, "users") {
			askedUsers = append(askedUsers, graphs.users[i])
		}

		singles := map[pair]single{}
		for _, question := range asked {
			for _, user := range askedUsers {
				if _, done := singles[pair{check: question.check, user: user}]; done {
					continue
				}
				statements.take()

				allowed, err := checker.AllowedFor(ctx, question.check, user)

				queries := len(statements.take())
				want := ref.check(question.object, question.permit, user)
				if checkVerdict(allowed, err) != want {
					t.Fatalf("AllowedFor(%#v, %s) = %t, %v; reference %s", question, user, allowed, err, want)
				}
				singles[pair{check: question.check, user: user}] = single{verdict: want, queries: queries}
				seen[want]++
			}
		}

		failAt := rapid.IntRange(0, 3).Draw(t, "failing acquisition")
		batch := func(name string, pairs []pair, call func() (map[pair]bool, error)) {
			statements.take()
			statements.failAcquisition(failAt)

			got, err := call()

			sent, failed := len(statements.take()), statements.acquisitionFailed()
			statements.failAcquisition(0)
			if problem := batchProblem(singles, pairs, got, err, sent, failed); problem != "" {
				t.Fatalf("%s: %s", name, problem)
			}
		}
		for _, question := range asked {
			var pairs []pair
			for _, user := range askedUsers {
				pairs = append(pairs, pair{check: question.check, user: user})
			}
			name := fmt.Sprintf("AllowedForEach(%#v, %d users)", question, len(askedUsers))
			batch(name, pairs, func() (map[pair]bool, error) {
				allowed, err := checker.AllowedForEach(ctx, question.check, askedUsers)
				return pairsOf(allowed, func(user uuid.UUID) pair { return pair{question.check, user} }), err
			})
		}
		for _, user := range askedUsers {
			var checks []permissions.Check
			var pairs []pair
			for _, question := range asked {
				checks = append(checks, question.check)
				pairs = append(pairs, pair{check: question.check, user: user})
			}
			name := fmt.Sprintf("AllowedEach(%d checks) as %s", len(checks), user)
			batch(name, pairs, func() (map[pair]bool, error) {
				allowed, err := checker.AllowedEach(signedIn(ctx, user.String()), checks...)
				return pairsOf(allowed, func(check permissions.Check) pair { return pair{check, user} }), err
			})
		}
	})

	for _, verdict := range []refVerdict{refDenied, refGranted, refFailed} {
		if seen[verdict] == 0 && !testing.Short() {
			t.Errorf("no generated question was %s; the generator no longer covers it", verdict)
		}
	}
	t.Logf("verdicts: %v", seen)
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

func (p pair) String() string { return fmt.Sprintf("%+v@%s", p.check, p.user) }

func pairsOf[K comparable](answers map[K]bool, pairOf func(K) pair) map[pair]bool {
	if answers == nil {
		return nil
	}

	pairs := make(map[pair]bool, len(answers))
	for key, allowed := range answers {
		pairs[pairOf(key)] = allowed
	}
	return pairs
}

// batchProblem compares one batched call with the single checks of its pairs. After a storage failure it is
// unavailable with that failure and returns no answers. Otherwise it fails exactly when one of the single checks
// fails, then with the depth error and no answers, otherwise returns their answers, and sends no more queries than the
// slowest of them.
func batchProblem(
	singles map[pair]single,
	pairs []pair,
	got map[pair]bool,
	err error,
	sent int,
	storageFailed bool,
) string {
	if storageFailed {
		if errx.KindOf(err) != errx.Unavailable || !errors.Is(err, errStorage) || got != nil {
			return fmt.Sprintf("after a storage failure: answers %v, error %v; want unavailable with %q", got, err,
				errStorage)
		}
		return ""
	}

	fails, budget, want := false, 0, map[pair]bool{}
	for _, p := range pairs {
		answer := singles[p]
		fails = fails || answer.verdict == refFailed
		budget = max(budget, answer.queries)
		want[p] = answer.verdict == refGranted
	}
	switch {
	case fails && (err == nil || !strings.Contains(err.Error(), maxDepthError)), !fails && err != nil:
		return fmt.Sprintf("error = %v, want the depth error %t", err, fails)
	case err != nil && got != nil:
		return fmt.Sprintf("answers %v with error %v", got, err)
	case sent > budget:
		return fmt.Sprintf("sent %d queries, more than the slowest single check's %d", sent, budget)
	case err == nil && !maps.Equal(got, want):
		return fmt.Sprintf("answers %v, single checks %v", got, want)
	}
	return ""
}
