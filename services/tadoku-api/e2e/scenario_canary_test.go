package e2e_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"maps"
	"strings"
	"testing"

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
	for _, file := range []string{
		"delete_tenants.sql",
		"update_overrides.sql",
		"write_shared.sql",
		"insert_tenant.sql",
		"write_migrations.sql",
	} {
		t.Run(strings.TrimSuffix(file, ".sql"), func(t *testing.T) {
			failures := runAborted(t, func(test *abortedTest) {
				s := newScenario(test, "identity-a")
				s.seed(test, "identity-a", tenantSeed{setup: []string{"testdata/ScenarioIsolation/" + file}})
			}, func() {})

			if len(failures) != 1 || !strings.Contains(failures[0], "permission denied") {
				t.Errorf("failures=%q, want the seed refused with permission denied", failures)
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
		{
			name: "seed_widens_its_binding",
			body: func(test *abortedTest) {
				s := newScenario(test, "primary")
				s.seed(test, "primary", tenantSeed{setup: []string{"testdata/ScenarioIsolation/widen_binding.sql"}})
			},
			wantFailure: "changed its tenant binding",
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
