package e2e_test

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

type tenantFootprint struct {
	jobs, completedJobs, audits, lockedUsers, cacheKeys int
}

func footprint(t *testing.T, s *fixtureScenario) tenantFootprint {
	t.Helper()
	key := s.tenant(t, primaryRole)

	var got tenantFootprint
	err := api.db.Pool.QueryRow(t.Context(), `select
		(select count(*) from jobs where tenant = $1),
		(select count(*) from jobs where tenant = $1 and state = 'completed'),
		(select count(*) from moderation_audit_log where tenant = $1),
		(select count(*) from users where tenant = $1 and deletion_locked_at is not null)`,
		key.String(),
	).Scan(&got.jobs, &got.completedJobs, &got.audits, &got.lockedUsers)
	if err != nil {
		t.Fatal(err)
	}

	keys, err := leaderboardValkey.keysMatching(t.Context(), "tenant:"+key.String()+":leaderboard:*")
	if err != nil {
		t.Fatal(err)
	}
	got.cacheKeys = len(keys)
	return got
}

// TestFixtureScenarioCasesStayInTheirTenants runs the same golden cases, with identical seeds, clocks and UUID
// streams, in two overlapping fixture scenarios. Each filing must find only its own tenant's user unlocked, each cache
// miss must write under its own tenant's prefix, a verify query must read only its own rows, and the first scenario's
// worker, including the job its job enqueues, must leave the second scenario's job pending.
func TestFixtureScenarioCasesStayInTheirTenants(t *testing.T) {
	impl := implementation{name: "tadoku-api", handler: api.handler}
	cases := []struct {
		directory string
		want      int
	}{
		{directory: "testdata/ImmersionFetchLeaderboardForYear/200_empty_cache_miss", want: http.StatusOK},
		{directory: "testdata/ProfileAccountDeletionCreate/202_normalized_email", want: http.StatusAccepted},
	}
	filed := tenantFootprint{jobs: 1, audits: 1, lockedUsers: 1, cacheKeys: 1}
	deleted := tenantFootprint{jobs: 2, completedJobs: 2, audits: 2, lockedUsers: 1, cacheKeys: 1}

	var started, ran sync.WaitGroup
	started.Add(2)
	ran.Add(2)
	firstWorked := make(chan struct{})

	for _, role := range []string{"first", "second"} {
		t.Run(role, func(t *testing.T) {
			t.Parallel()
			s := newFixtureScenario(t)
			meetScenario(t, &started)

			for _, test := range cases {
				s.runCase(t, test.directory, test.want, fixtureInstant, impl)
			}
			meetScenario(t, &ran)
			if got := footprint(t, s); got != filed {
				t.Fatalf("%s after filing: %+v, want only its own %+v", role, got, filed)
			}
			if got := s.verifyQuery(t, "select count(*) as audits from moderation_audit_log"); got != `[{"audits":1}]` {
				t.Fatalf("%s verify query read %s, want only its own audit row", role, got)
			}

			if role == "second" {
				select {
				case <-firstWorked:
				case <-time.After(30 * time.Second):
					t.Fatal("the first scenario's worker never finished")
				}
				if got := footprint(t, s); got != filed {
					t.Fatalf("the first scenario's worker reached the second: %+v, want %+v", got, filed)
				}
			}
			s.workerStep(t, "")
			if role == "first" {
				close(firstWorked)
			}

			if got := footprint(t, s); got != deleted {
				t.Errorf("%s after its worker: %+v, want %+v", role, got, deleted)
			}
		})
	}
}
