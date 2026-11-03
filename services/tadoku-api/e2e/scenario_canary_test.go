package e2e_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

var scenarioIsolationSeed = tenantSeed{relationships: []string{"testdata/ScenarioIsolation/relationships.json"}}

var identityRoles = []string{"identity-a", "identity-b"}

func TestScenarioTeardownPreservesActiveScenario(t *testing.T) {
	removedReady := make(chan []tenant.TestKey, 1)
	activeReady := make(chan struct{})
	removedTornDown := make(chan struct{})

	t.Run("pair", func(t *testing.T) {
		t.Run("removed", func(t *testing.T) {
			t.Parallel()
			t.Cleanup(func() { close(removedTornDown) })
			removed := newScenario(t, identityRoles...)
			for _, role := range identityRoles {
				populateScenario(t, removed, role)
			}
			removedReady <- []tenant.TestKey{removed.tenant(t, "identity-a"), removed.tenant(t, "identity-b")}

			<-activeReady
		})
		t.Run("active", func(t *testing.T) {
			t.Parallel()
			active := newScenario(t, identityRoles...)
			before := map[string]map[string]string{}
			for _, role := range identityRoles {
				populateScenario(t, active, role)
				before[role] = scenarioState(t, active.tenant(t, role))
			}
			close(activeReady)

			removedKeys := <-removedReady
			<-removedTornDown

			for _, key := range removedKeys {
				requireNoScenarioResidue(t, key)
			}
			for _, role := range identityRoles {
				key := active.tenant(t, role)
				if after := scenarioState(t, key); !maps.Equal(before[role], after) {
					t.Errorf("tearing down %v changed %s: before=%v after=%v", removedKeys, key, before[role], after)
				}
				requireScenarioReadable(t, active, role)
			}
		})
	})
}

func TestScenarioSeedIsConfinedToItsTenant(t *testing.T) {
	active := newScenario(t, "identity-b")
	populateScenario(t, active, "identity-b")
	activeKey := active.tenant(t, "identity-b")
	if err := scenarioLifecycle().SetOverride(t.Context(), activeKey, jobqueue.WorkerComponent); err != nil {
		t.Fatal(err)
	}
	activeBefore := scenarioState(t, activeKey)
	sharedBefore := sharedSeedTargets(t)

	confined := newScenario(t, "identity-a")
	confined.seed(t, "identity-a", tenantSeed{finalize: []string{"testdata/ScenarioIsolation/unscoped_delete.sql"}})

	if rows := tenantRows(t, confined.tenant(t, "identity-a").Key(), "accounts"); rows != "[]" {
		t.Errorf("unscoped delete left the seeding tenant's accounts: %s", rows)
	}
	switched := filepath.Join(t.TempDir(), "switch_tenant.sql")
	switchScript := fmt.Sprintf(`select set_config('canary.saved', current_setting('tadoku.tenant'), true);
select set_config('tadoku.tenant', '%s', true);
delete from sessions;
select set_config('tadoku.tenant', current_setting('canary.saved'), true);
`, activeKey)
	err := os.WriteFile(switched, []byte(switchScript), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	refused := []struct {
		name string
		file string
		want string
	}{
		{name: "delete_tenants", file: "testdata/ScenarioIsolation/delete_tenants.sql", want: "permission denied"},
		{name: "update_overrides", file: "testdata/ScenarioIsolation/update_overrides.sql", want: "permission denied"},
		{name: "write_shared", file: "testdata/ScenarioIsolation/write_shared.sql", want: "permission denied"},
		{name: "insert_tenant", file: "testdata/ScenarioIsolation/insert_tenant.sql", want: "permission denied"},
		{name: "write_migrations", file: "testdata/ScenarioIsolation/write_migrations.sql", want: "permission denied"},
		{name: "write_jobs", file: "testdata/ScenarioIsolation/write_jobs.sql", want: "permission denied"},
		{
			name: "widen_and_restore",
			file: "testdata/ScenarioIsolation/widen_and_restore.sql",
			want: "changes session state",
		},
		{name: "switch_tenant_and_restore", file: switched, want: "changes session state"},
	}
	for _, test := range refused {
		t.Run(test.name, func(t *testing.T) {
			failures := runAborted(t, func(aborted *abortedTest) {
				s := newScenario(aborted, "identity-a")
				s.seed(aborted, "identity-a", tenantSeed{setup: []string{test.file}})
			}, func() {})

			if len(failures) != 1 || !strings.Contains(failures[0], test.want) {
				t.Errorf("failures=%q, want the seed refused with %q", failures, test.want)
			}
		})
	}

	if after := scenarioState(t, activeKey); !maps.Equal(activeBefore, after) {
		t.Errorf("another scenario's seed changed %s: before=%v after=%v", activeKey, activeBefore, after)
	}
	if after := sharedSeedTargets(t); after != sharedBefore {
		t.Errorf("scenario seeds changed shared state: before=%s after=%s", sharedBefore, after)
	}
	requireScenarioReadable(t, active, "identity-b")
}

// sharedSeedTargets snapshots the canonical rows, test tenant registry and shared reference rows that the confinement
// canary's seeds try to change.
func sharedSeedTargets(t *testing.T) string {
	t.Helper()
	var shared string
	err := api.db.Pool.QueryRow(t.Context(), `select jsonb_build_array(
		(select jsonb_agg(key order by key) from tenants where key not like 'e2e/identity-a-%'),
		(select jsonb_agg(to_jsonb(language) order by code) from languages language)
	)::text`).Scan(&shared)
	if err != nil {
		t.Fatal(err)
	}
	return shared + tenantRows(t, tenant.Production(), "accounts") +
		tenantRows(t, tenant.Production(), "permission_relationships")
}

func TestScenarioWorkerClaimsOnlyItsTenant(t *testing.T) {
	var keys []tenant.TestKey

	t.Run("workers", func(t *testing.T) {
		scoped := newScenario(t, "primary")
		bystander := newScenario(t, "primary")
		keys = []tenant.TestKey{scoped.tenant(t, "primary"), bystander.tenant(t, "primary")}
		queued := map[*scenario]int{}
		for _, s := range []*scenario{scoped, bystander} {
			populateScenario(t, s, "primary")
			queued[s] = scenarioJobs(t, s.tenant(t, "primary")).total
		}
		bystanderJobs := tenantRows(t, bystander.tenant(t, "primary").Key(), "jobs")

		scoped.runWorker(t, "primary")

		if got := tenantRows(t, bystander.tenant(t, "primary").Key(), "jobs"); got != bystanderJobs {
			t.Errorf("scoped worker touched the other scenario's jobs: before=%s after=%s", bystanderJobs, got)
		}
		for _, s := range []*scenario{scoped, bystander} {
			s.runWorker(t, "primary")
			jobs := scenarioJobs(t, s.tenant(t, "primary"))
			if jobs.completed != jobs.total || jobs.total <= queued[s] {
				t.Errorf("tenant %s jobs=%+v, want all completed including follow-ups beyond %d queued",
					s.tenant(t, "primary"), jobs, queued[s])
			}
		}
	})

	for _, key := range keys {
		requireNoScenarioResidue(t, key)
	}
}

func TestScenarioCleanupAfterAbortedSetup(t *testing.T) {
	active := newScenario(t, "primary")
	populateScenario(t, active, "primary")
	activeKey := active.tenant(t, "primary")
	activeBefore := scenarioState(t, activeKey)
	productionBefore := tenantRows(t, tenant.Production(), "users")

	cases := []struct {
		name          string
		body          func(*abortedTest)
		beforeCleanup func()
		wantFailure   string
	}{
		{
			name: "fatal_assertion_with_running_worker",
			body: func(test *abortedTest) {
				s := newScenario(test, identityRoles...)
				populateScenario(test, s, "identity-a")
				s.startWorker(test, "identity-a")
				test.Fatal("deliberate failure")
			},
			wantFailure: "deliberate failure",
		},
		{
			name: "cancelled_context",
			body: func(test *abortedTest) {
				s := newScenario(test, "primary")
				test.cancel()
				s.seed(test, "primary", scenarioIsolationSeed)
			},
			wantFailure: "context canceled",
		},
		{
			name: "failed_provisioning",
			body: func(test *abortedTest) {
				flipt.SetUnavailable(true)
				newScenario(test, identityRoles...)
			},
			beforeCleanup: func() { flipt.SetUnavailable(false) },
			wantFailure:   "provision scenario tenant",
		},
		{
			name: "seed_writes_another_tenant",
			body: func(test *abortedTest) {
				s := newScenario(test, "primary")
				s.seed(test, "primary", tenantSeed{setup: []string{"testdata/ScenarioIsolation/foreign_tenant.sql"}})
			},
			wantFailure: "row-level security",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var keys []tenant.TestKey
			before := allocatedTenants()

			failures := runAborted(t, test.body, func() {
				keys = newlyAllocated(before)
				if test.beforeCleanup != nil {
					test.beforeCleanup()
				}
			})

			if len(failures) != 1 || !strings.Contains(failures[0], test.wantFailure) {
				t.Errorf("failures=%q, want exactly one containing %q", failures, test.wantFailure)
			}
			if len(keys) == 0 {
				t.Fatal("aborted scenario allocated no tenants")
			}
			for _, key := range keys {
				requireNoScenarioResidue(t, key)
			}
		})
	}

	if after := scenarioState(t, activeKey); !maps.Equal(activeBefore, after) {
		t.Errorf("aborted scenarios changed active %s: before=%v after=%v", activeKey, activeBefore, after)
	}
	if after := tenantRows(t, tenant.Production(), "users"); after != productionBefore {
		t.Errorf("aborted scenarios changed canonical users: before=%s after=%s", productionBefore, after)
	}
	requireScenarioReadable(t, active, "primary")
}

const scenarioScoringSQL = `select coalesce(jsonb_agg(jsonb_build_array(
		rule.priority, rule.stackable, rule.activity_id, rule.unit_key, rule.language_code, rule.tag,
		rule.score_source, rule.rate
	) order by rule.priority), '[]')::text
	from %s config
	join %s rule on rule.rule_set_id = config.active_rule_set_id and rule.tenant = config.tenant
	where config.tenant = $1`

// Test safety: this exclusive test removes the canonical platform scoring and publishes a canonical page; its
// cleanup restores both after the scenario's teardown.
func TestScenarioProvisioningIgnoresCanonicalState(t *testing.T) {
	rows, err := api.db.Pool.Query(t.Context(),
		"delete from platform_scoring_config where tenant = 'tadoku/prod' returning active_rule_set_id",
	)
	if err != nil {
		t.Fatal(err)
	}
	removed, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		t.Fatal(err)
	}
	pageID, contentID := uuid.New(), uuid.New()
	_, err = api.db.Pool.Exec(t.Context(), `insert into pages
			(tenant, id, "namespace", slug, current_content_id, published_at)
		values ('tadoku/prod', $1, 'tadoku', $3, $2, now() - interval '1 day')`,
		pageID,
		contentID,
		"scenario-canary-"+hex.EncodeToString(pageID[:4]),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = api.db.Pool.Exec(t.Context(), `insert into pages_content (tenant, id, page_id, title, html)
		values ('tadoku/prod', $2, $1, 'Canonical canary', '<p>Canonical</p>')`,
		pageID,
		contentID,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		batch := &pgx.Batch{}
		batch.Queue(`insert into platform_scoring_config (tenant, singleton, active_rule_set_id)
			select 'tadoku/prod', true, unnest($1::uuid[])`, removed)
		batch.Queue("delete from pages_content where id = $1", contentID)
		batch.Queue("delete from pages where id = $1", pageID)
		if err := api.db.Pool.SendBatch(context.Background(), batch).Close(); err != nil {
			t.Error(err)
		}
	})

	s := newScenario(t, "primary")

	key := s.tenant(t, "primary").Key()
	var baseline, scoring string
	err = api.db.Pool.QueryRow(t.Context(), fmt.Sprintf(
		scenarioScoringSQL,
		"tadoku_test_platform_scoring_config_baseline",
		"tadoku_test_scoring_rules_baseline",
	), tenant.Production().String()).Scan(&baseline)
	if err != nil {
		t.Fatal(err)
	}
	err = api.db.Pool.QueryRow(t.Context(), fmt.Sprintf(scenarioScoringSQL, "platform_scoring_config", "scoring_rules"),
		key.String(),
	).Scan(&scoring)
	if err != nil {
		t.Fatal(err)
	}
	if scoring == "[]" || scoring != baseline {
		t.Errorf("scenario platform scoring=%s, want migration baseline %s", scoring, baseline)
	}
	if pages := tenantRows(t, key, "pages") + tenantRows(t, key, "pages_content"); pages != "[][]" {
		t.Errorf("scenario copied canonical pages: %s", pages)
	}
}

type scenarioTimestamps struct {
	emailID                       uuid.UUID
	emailCreated, recoveryExpires time.Time
	jobCreated, jobCompleted      time.Time
}

func TestScenarioClocksDivergeUnderOverlap(t *testing.T) {
	offsets := map[string]time.Duration{"fixture": 0, "ahead": 72 * time.Hour}
	var started, requested sync.WaitGroup
	started.Add(len(offsets))
	requested.Add(len(offsets))

	for name, offset := range offsets {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			s := newIdentityScenario(t)
			key := s.tenant(t, "identity-a")
			s.clock.Advance(offset)
			meetScenario(t, &started)

			requireRecoveryAccepted(t, recoverIn(t, s, api.handler, "reader@example.test", "198.51.100.70", nil))
			requireRecoveryWrites(t, key.String(), 1)
			meetScenario(t, &requested)

			s.clock.Advance(time.Minute)
			s.runWorker(t, "identity-a")

			var got scenarioTimestamps
			err := api.db.Pool.QueryRow(t.Context(), `select
				email.id, email.created_at, token.expires_at, job.created_at, job.completed_at
				from emails email, recovery_tokens token, jobs job
				where email.tenant = $1 and token.tenant = $1 and job.tenant = $1 and job.task_type = 'email.send.v1'`,
				key.String(),
			).Scan(&got.emailID, &got.emailCreated, &got.recoveryExpires, &got.jobCreated, &got.jobCompleted)
			if err != nil {
				t.Fatal(err)
			}
			for _, at := range []*time.Time{&got.emailCreated, &got.recoveryExpires, &got.jobCreated, &got.jobCompleted} {
				*at = at.UTC()
			}

			ids := newScenarioIDs(key)
			correlationID, emailID := uuid.Must(uuid.NewV7FromReader(ids)), uuid.Must(uuid.NewRandomFromReader(ids))
			requestedAt := fixtureInstant.Add(offset)
			want := scenarioTimestamps{
				emailID:         emailID,
				emailCreated:    requestedAt,
				recoveryExpires: requestedAt.Add(time.Hour),
				jobCreated:      requestedAt,
				jobCompleted:    requestedAt.Add(time.Minute),
			}
			t.Logf("%s clock +%s after correlation ID %s wrote %s", key, offset, correlationID, got)
			if got != want {
				t.Errorf("%s wrote %s, want its own clock and IDs %s", key, got, want)
			}
		})
	}
}

func (got scenarioTimestamps) String() string {
	return fmt.Sprintf("email %s created %s, recovery expires %s, job created %s and completed %s",
		got.emailID, got.emailCreated.Format(time.RFC3339), got.recoveryExpires.Format(time.RFC3339),
		got.jobCreated.Format(time.RFC3339), got.jobCompleted.Format(time.RFC3339))
}

// meetScenario blocks until every scenario of group has arrived, so their next steps overlap.
func meetScenario(t *testing.T, group *sync.WaitGroup) {
	t.Helper()
	group.Done()
	arrived := make(chan struct{})
	go func() {
		group.Wait()
		close(arrived)
	}()
	select {
	case <-arrived:
	case <-time.After(30 * time.Second):
		t.Fatal("the overlapping scenario never arrived; it needs -test.parallel of at least 2")
	}
}
