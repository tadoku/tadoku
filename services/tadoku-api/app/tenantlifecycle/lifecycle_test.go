package tenantlifecycle_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/app/tenantlifecycle"
	"github.com/tadoku/tadoku/services/tadoku-api/features/jobqueue"
	"github.com/tadoku/tadoku/services/tadoku-api/features/leaderboard"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/flipt"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/keto"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/valkey"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/permissions"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testflipt"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testketo"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testvalkey"
	valkeygo "github.com/valkey-io/valkey-go"
)

func TestTenantLifecycle(t *testing.T) {
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	ketoFixture, err := testketo.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := ketoFixture.Close(); err != nil {
			t.Error(err)
		}
	})
	fliptFixture := testflipt.New()
	t.Cleanup(fliptFixture.Close)
	rawURL, err := testvalkey.URL()
	if err != nil {
		t.Fatal(err)
	}
	client, err := valkey.Open(t.Context(), rawURL, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(client.Close)
	targets, err := flipt.NewTargets("local", "default", "test")
	if err != nil {
		t.Fatal(err)
	}
	ketoClient := keto.NewClient(ketoFixture.ReadURL(), ketoFixture.WriteURL(),
		keto.WithHTTPClient(&http.Client{Timeout: 2 * time.Second}))
	application := tenantlifecycle.NewApplication(
		db.Pool,
		permissions.NewTenantManager(ketoClient),
		leaderboard.NewCache(client, time.Second),
		fliptmanagement.NewClient(fliptmanagement.Config{URL: fliptFixture.URL(), Targets: targets}),
	)
	key := lifecycleKey(t, "removed")
	other := lifecycleKey(t, "preserved")
	reader := uuid.New()
	features, err := fliptmanagement.ParseFeatures(strings.NewReader(`version: "1.6"
namespace:
  key: default
flags:
  - key: lifecycle
    name: Lifecycle
    type: BOOLEAN_FLAG_TYPE
    enabled: false
segments: []
`))
	if err != nil {
		t.Fatal(err)
	}

	for _, target := range []tenant.TestKey{key, key, other} {
		if err := application.Provision(t.Context(), target, []uuid.UUID{reader}, features); err != nil {
			t.Fatal(err)
		}
	}
	var registry int
	if err := db.Pool.QueryRow(t.Context(), "select count(*) from tenants where key = $1", key.String()).
		Scan(&registry); err != nil || registry != 1 {
		t.Fatalf("idempotent tenant registry: rows=%d error=%v", registry, err)
	}
	for _, relation := range []string{"testers", "access"} {
		allowed, err := ketoClient.CheckPermission(t.Context(), "app", key.String(), relation, keto.Subject{ID: reader.String()})
		if err != nil || !allowed {
			t.Fatalf("provisioned %s: allowed=%t error=%v", relation, allowed, err)
		}
	}
	if err := application.SetOverride(t.Context(), key, jobqueue.WorkerComponent); err != nil {
		t.Fatal(err)
	}
	if err := application.SetOverride(t.Context(), key, jobqueue.WorkerComponent); err != nil {
		t.Fatal(err)
	}
	if err := application.ClearOverride(t.Context(), key, jobqueue.WorkerComponent); err != nil {
		t.Fatal(err)
	}
	if err := application.ClearOverride(t.Context(), key, jobqueue.WorkerComponent); err != nil {
		t.Fatal(err)
	}
	if err := application.SetOverride(t.Context(), key, "unregistered-worker"); err == nil {
		t.Fatal("unregistered component was accepted")
	}
	missing := lifecycleKey(t, "missing")
	if err := application.SetOverride(t.Context(), missing, jobqueue.WorkerComponent); err == nil {
		t.Fatal("override without a tenant row was accepted")
	}
	if err := application.SetOverride(t.Context(), key, jobqueue.WorkerComponent); err != nil {
		t.Fatal(err)
	}

	for _, target := range []string{tenant.Production().String(), key.String(), other.String()} {
		seedEveryTenantTable(t, db, target)
	}
	beforeProduction := ownedRows(t, db, tenant.Production().String())
	beforeOther := ownedRows(t, db, other.String())
	for table, rows := range ownedRows(t, db, key.String()) {
		if rows == "[]" {
			t.Fatalf("teardown fixture leaves %s empty", table)
		}
	}
	cacheKeys := []string{
		"tenant:" + key.String() + ":leaderboard:global",
		"tenant:" + other.String() + ":leaderboard:global",
		"tenant:" + tenant.Production().String() + ":leaderboard:lifecycle-" + uuid.NewString(),
	}
	for _, name := range cacheKeys {
		if err := client.Do(t.Context(), client.B().Set().Key(name).Value("preserve").Build()).Error(); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := client.Do(ctx, client.B().Unlink().Key(cacheKeys...).Build()).Error(); err != nil {
			t.Error(err)
		}
	})

	fliptFixture.SetUnavailable(true)
	if err := application.Teardown(t.Context(), key); err == nil {
		t.Fatal("injected Flipt failure did not fail teardown")
	}
	if err := db.Pool.QueryRow(t.Context(), "select count(*) from tenants where key = $1", key.String()).
		Scan(&registry); err != nil || registry != 1 {
		t.Fatalf("failed teardown lost retryable registry: rows=%d error=%v", registry, err)
	}
	fliptFixture.SetUnavailable(false)
	for range 2 {
		if err := application.Teardown(t.Context(), key); err != nil {
			t.Fatal(err)
		}
	}
	for table, rows := range ownedRows(t, db, key.String()) {
		if rows != "[]" {
			t.Errorf("cascade left %s=%s", table, rows)
		}
	}
	if after := ownedRows(t, db, tenant.Production().String()); !reflect.DeepEqual(beforeProduction, after) {
		t.Errorf("canonical rows changed: before=%v after=%v", beforeProduction, after)
	}
	if after := ownedRows(t, db, other.String()); !reflect.DeepEqual(beforeOther, after) {
		t.Errorf("other test rows changed: before=%v after=%v", beforeOther, after)
	}
	for index, name := range cacheKeys {
		value, err := client.Do(t.Context(), client.B().Get().Key(name).Build()).ToString()
		if index == 0 && err != valkeygo.Nil {
			t.Errorf("removed cache key survives: value=%q error=%v", value, err)
		}
		if index > 0 && (err != nil || value != "preserve") {
			t.Errorf("preserved cache key changed: value=%q error=%v", value, err)
		}
	}
	allowed, err := ketoClient.CheckPermission(t.Context(), "app", other.String(), "access", keto.Subject{ID: reader.String()})
	if err != nil || !allowed {
		t.Errorf("other Keto access changed: allowed=%t error=%v", allowed, err)
	}
	allowed, err = ketoClient.CheckPermission(t.Context(), "app", key.String(), "access", keto.Subject{ID: reader.String()})
	if err != nil || allowed {
		t.Errorf("removed Keto access remains: allowed=%t error=%v", allowed, err)
	}
	if fliptFixture.Exists(key.Key()) || !fliptFixture.Exists(other.Key()) || !fliptFixture.Exists(tenant.Production()) {
		t.Error("Flipt teardown failed to remove only its tenant namespace")
	}

	orphan := lifecycleKey(t, "orphan")
	if err := application.Provision(t.Context(), orphan, []uuid.UUID{reader}, features); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Pool.Exec(t.Context(), "delete from tenants where key = $1", orphan.String()); err != nil {
		t.Fatal(err)
	}
	if err := application.Teardown(t.Context(), orphan); err != nil {
		t.Fatal(err)
	}
	if fliptFixture.Exists(orphan.Key()) {
		t.Fatal("missing registry row prevented provider cleanup")
	}

	if err := application.Provision(t.Context(), tenant.TestKey{}, nil, features); err == nil {
		t.Fatal("provision accepted zero test key")
	}
	if err := application.Teardown(t.Context(), tenant.TestKey{}); err == nil {
		t.Fatal("teardown accepted zero test key")
	}
	if err := application.SetOverride(t.Context(), tenant.TestKey{}, jobqueue.WorkerComponent); err == nil {
		t.Fatal("override accepted zero test key")
	}
	requestsBefore := fliptFixture.RequestCount()
	runtimeApplication := tenantlifecycle.NewApplication(db.AppPool, permissions.NewTenantManager(ketoClient),
		leaderboard.NewCache(client, time.Second),
		fliptmanagement.NewClient(fliptmanagement.Config{URL: fliptFixture.URL(), Targets: targets}))
	if err := runtimeApplication.Provision(t.Context(), other, []uuid.UUID{reader}, features); err == nil {
		t.Fatal("runtime credentials were accepted for provisioning")
	}
	if err := runtimeApplication.Teardown(t.Context(), other); err == nil {
		t.Fatal("runtime credentials were accepted for teardown")
	}
	if err := runtimeApplication.SetOverride(t.Context(), other, jobqueue.WorkerComponent); err == nil {
		t.Fatal("runtime credentials were accepted for override set")
	}
	if err := runtimeApplication.ClearOverride(t.Context(), other, jobqueue.WorkerComponent); err == nil {
		t.Fatal("runtime credentials were accepted for override clear")
	}
	if fliptFixture.RequestCount() != requestsBefore {
		t.Fatal("runtime credentials reached the provider")
	}

	// Test safety: only this disposable database loses its production-key check to simulate a corrupted registry.
	if _, err := db.Pool.Exec(t.Context(), "alter table tenants drop constraint tenants_single_production"); err != nil {
		t.Fatal(err)
	}
	corrupt := lifecycleKey(t, "production-row")
	if _, err := db.Pool.Exec(t.Context(), "insert into tenants (key, kind) values ($1, 'production')", corrupt.String()); err != nil {
		t.Fatal(err)
	}
	if err := application.Provision(t.Context(), corrupt, nil, features); err == nil {
		t.Fatal("provision accepted production-kind row")
	}
	if err := application.Teardown(t.Context(), corrupt); err == nil {
		t.Fatal("teardown accepted production-kind row")
	}
	if err := application.SetOverride(t.Context(), corrupt, jobqueue.WorkerComponent); err == nil {
		t.Fatal("override accepted production-kind row")
	}
	if err := application.ClearOverride(t.Context(), corrupt, jobqueue.WorkerComponent); err == nil {
		t.Fatal("override clear accepted production-kind row")
	}
	t.Log("lifecycle workflow passed: idempotent provision, all-table cascade, preserved canonical/other rows, retry and refusals")
}

func lifecycleKey(t *testing.T, name string) tenant.TestKey {
	t.Helper()
	key, err := tenant.ParseTestTenant("e2e/" + name + "-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func ownedRows(t *testing.T, db *testpostgres.Database, key string) map[string]string {
	t.Helper()
	rows, err := db.Pool.Query(t.Context(), `
		select c.relname
		from pg_class c
		join pg_namespace n on n.oid = c.relnamespace
		join pg_attribute a on a.attrelid = c.oid and a.attname = 'tenant' and not a.attisdropped
		where n.nspname = current_schema()
			and c.relkind in ('r', 'p')
			and not starts_with(c.relname, 'tadoku_test_')
		order by c.relname`)
	if err != nil {
		t.Fatal(err)
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil || len(tables) == 0 {
		t.Fatalf("tenant-owned inventory=%v error=%v", tables, err)
	}
	result := make(map[string]string, len(tables))
	for _, table := range tables {
		var value string
		query := "select coalesce(jsonb_agg(to_jsonb(row) order by to_jsonb(row)::text), '[]'::jsonb)::text" +
			" from " + pgx.Identifier{table}.Sanitize() + " row where tenant = $1"
		if err := db.Pool.QueryRow(t.Context(), query, key).Scan(&value); err != nil {
			t.Fatal(err)
		}
		result[table] = value
	}
	encoded, _ := json.Marshal(result)
	t.Logf("tenant table snapshot %s=%s", key, encoded)
	return result
}

func seedEveryTenantTable(t *testing.T, db *testpostgres.Database, key string) {
	t.Helper()
	seed := `
		select set_config('tadoku.tenant', $1, true);
		insert into users (tenant, id, display_name)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'), 'Lifecycle user');
		insert into user_roles (tenant, user_id, role)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'), 'admin');
		insert into profiles (tenant, user_id)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'));
		insert into account_deletion_requests (tenant, identity_id, status, accepted_at)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'), 'receipt_pending', now());
		insert into moderation_audit_log (tenant, user_id, action)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'), 'lifecycle-fixture');
		insert into contests (tenant, id, owner_user_id, owner_user_display_name, private,
			contest_start, contest_end, registration_end, title, activity_type_id_allow_list, official)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/contest'), uuid_generate_v5(uuid_nil(), $1 || '/user'),
			'Lifecycle user', false, current_date, current_date + 7, current_date + 7, 'Lifecycle', array[1], false);
		insert into contest_registrations (tenant, contest_id, user_id, language_codes)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/contest'), uuid_generate_v5(uuid_nil(), $1 || '/user'), array['jpn']);
		insert into logs (tenant, id, user_id, language_code, log_activity_id, unit_id, unit_key,
			amount, modifier, computed_score, eligible_official_leaderboard)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/log'), uuid_generate_v5(uuid_nil(), $1 || '/user'),
			'jpn', 1, (select id from log_units where unit_key = 'reading_page' limit 1), 'reading_page', 10, 1, 10, true);
		insert into contest_logs (tenant, contest_id, log_id, unit_key, amount, modifier, computed_score)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/contest'), uuid_generate_v5(uuid_nil(), $1 || '/log'),
			'reading_page', 10, 1, 10);
		insert into log_tags (tenant, log_id, user_id, tag)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/log'), uuid_generate_v5(uuid_nil(), $1 || '/user'), 'lifecycle');
		insert into pages (tenant, id, namespace, slug, current_content_id)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/page'), 'tadoku', 'lifecycle',
			uuid_generate_v5(uuid_nil(), $1 || '/page-content'));
		insert into pages_content (tenant, id, page_id, title, html)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/page-content'), uuid_generate_v5(uuid_nil(), $1 || '/page'),
			'Lifecycle', '<p>Fixture</p>');
		insert into posts (tenant, id, namespace, slug, current_content_id)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/post'), 'tadoku', 'lifecycle',
			uuid_generate_v5(uuid_nil(), $1 || '/post-content'));
		insert into posts_content (tenant, id, post_id, title, content)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/post-content'), uuid_generate_v5(uuid_nil(), $1 || '/post'),
			'Lifecycle', 'Fixture');
		insert into announcements (tenant, namespace, title, content, starts_at, ends_at)
		values ($1, 'tadoku', 'Lifecycle', 'Fixture', now(), now() + interval '1 day');
		insert into scoring_rule_sets (tenant, id, scope, version, status)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/rule-set'), 'platform', 999, 'draft');
		insert into scoring_rules (tenant, rule_set_id, priority, stackable, activity_id, score_source, rate)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/rule-set'), 1, false, 1, 'amount', 1);
		insert into platform_scoring_config (tenant, singleton, active_rule_set_id)
		values ($1, true, uuid_generate_v5(uuid_nil(), $1 || '/rule-set'))
		on conflict (tenant, singleton) do nothing;
		insert into jobs (tenant, task_type, payload)
		values ($1, 'fixture.lifecycle.v1', '{}');
		insert into tenant_overrides (tenant, component)
		select $1, 'tadoku-worker' where $1 <> 'tadoku/prod'
		on conflict do nothing;
	`
	tx, err := db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tx.Rollback(ctx)
	})
	if _, err := tx.Exec(t.Context(), seed, pgx.QueryExecModeSimpleProtocol, key); err != nil {
		t.Fatalf("seed %s: %v", key, err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
}
