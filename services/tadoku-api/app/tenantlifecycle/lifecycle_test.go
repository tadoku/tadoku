package tenantlifecycle_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
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
	lifecycle := lifecyclePool(t, db)
	application := tenantlifecycle.NewApplication(
		lifecycle,
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

	seedEveryTenantTable(t, db, tenant.Production().String())
	seedProductionPages(t, db)
	productionBeforeProvision := ownedRows(t, db, tenant.Production().String())
	for _, target := range []tenant.TestKey{key, key, other} {
		if err := application.Provision(t.Context(), target, []uuid.UUID{reader}, features); err != nil {
			t.Fatal(err)
		}
	}
	assertCopiedPlatformScoring(t, db, key)
	assertCopiedPublishedPages(t, db, key)
	if after := ownedRows(t, db, tenant.Production().String()); !reflect.DeepEqual(productionBeforeProvision, after) {
		t.Errorf("provisioning changed canonical rows: before=%v after=%v", productionBeforeProvision, after)
	}
	for _, copySQL := range []string{copyScoringSQL, copyPagesSQL} {
		if _, err := lifecycle.Exec(t.Context(), copySQL, tenant.Production().String()); err == nil {
			t.Errorf("lifecycle credentials ran %q for the canonical tenant", copySQL)
		}

		_, err = db.AppPool.Exec(t.Context(), copySQL, other.String())
		var pgError *pgconn.PgError
		if !errors.As(err, &pgError) || pgError.Code != "42501" {
			t.Errorf("runtime %q=%v, want PostgreSQL42501", copySQL, err)
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

	for _, target := range []string{key.String(), other.String()} {
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
	ownerApplication := tenantlifecycle.NewApplication(db.Pool, permissions.NewTenantManager(ketoClient),
		leaderboard.NewCache(client, time.Second),
		fliptmanagement.NewClient(fliptmanagement.Config{URL: fliptFixture.URL(), Targets: targets}))
	if err := ownerApplication.Teardown(t.Context(), other); err != nil {
		t.Fatalf("owner credentials were refused for teardown: %v", err)
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

const (
	copyScoringSQL = "select copy_production_platform_scoring($1)"
	copyPagesSQL   = "select copy_production_published_pages($1)"
)

func lifecycleKey(t *testing.T, name string) tenant.TestKey {
	t.Helper()
	key, err := tenant.ParseTestTenant("e2e/" + name + "-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// lifecyclePool connects as a login whose only grants are the production tadoku_tenant_lifecycle role's.
func lifecyclePool(t *testing.T, db *testpostgres.Database) *pgxpool.Pool {
	t.Helper()
	role := "tadoku_test_lifecycle_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	password := uuid.NewString()
	roleSQL := pgx.Identifier{role}.Sanitize()
	setup := `
		do $$
		begin
			create role tadoku_tenant_lifecycle nologin;
		exception when duplicate_object or unique_violation then
			null;
		end
		$$;
		grant usage on schema public to tadoku_tenant_lifecycle;
		grant select, insert, delete on tenants, tenant_overrides to tadoku_tenant_lifecycle;
		grant execute on function copy_production_platform_scoring(text) to tadoku_tenant_lifecycle;
		grant execute on function copy_production_published_pages(text) to tadoku_tenant_lifecycle;
		create role ` + roleSQL + ` login nosuperuser nobypassrls in role tadoku_tenant_lifecycle password '` + password + `'`
	if _, err := db.Pool.Exec(t.Context(), setup); err != nil {
		t.Fatal(err)
	}

	cfg, err := pgxpool.ParseConfig(db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.User = role
	cfg.ConnConfig.Password = password
	cfg.MaxConns = 2
	pool, err := pgxpool.NewWithConfig(t.Context(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		pool.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if _, err := db.Pool.Exec(ctx, "drop role "+roleSQL); err != nil {
			t.Error(err)
		}
	})
	return pool
}
