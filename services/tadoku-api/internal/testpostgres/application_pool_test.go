package testpostgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestApplicationPoolPrivilegesAndCleanup(t *testing.T) {
	db, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	closed := false
	t.Cleanup(func() {
		if !closed {
			if err := db.Close(); err != nil {
				t.Error(err)
			}
		}
	})
	pool := db.AppPool
	var role string
	var superuser, bypass, metadataSelect, metadataWrite, dml, sequenceUsage bool
	var owned, memberships int
	err = pool.QueryRow(t.Context(), `select current_user, r.rolsuper, r.rolbypassrls,
		(select count(*) from pg_class where relowner = r.oid),
		(select count(*) from pg_auth_members where member = r.oid),
		has_table_privilege('schema_migrations', 'SELECT'),
		has_table_privilege('schema_migrations', 'INSERT,UPDATE,DELETE,TRUNCATE'),
		has_table_privilege('users', 'SELECT') and has_table_privilege('users', 'INSERT')
			and has_table_privilege('users', 'UPDATE') and has_table_privilege('users', 'DELETE'),
		has_sequence_privilege('jobs_id_seq', 'USAGE')
		from pg_roles r where r.rolname = current_user
	`).Scan(&role, &superuser, &bypass, &owned, &memberships, &metadataSelect, &metadataWrite, &dml, &sequenceUsage)
	if err != nil {
		t.Fatal(err)
	}
	if superuser || bypass || owned != 0 || memberships != 0 || !metadataSelect || metadataWrite || !dml || !sequenceUsage {
		t.Fatalf("unsafe application pool: superuser=%t bypass=%t owned=%d memberships=%d metadata=%t/%t DML=%t sequence=%t", superuser, bypass, owned, memberships, metadataSelect, metadataWrite, dml, sequenceUsage)
	}
	if _, err := pool.Exec(t.Context(), "update schema_migrations set version = version"); err == nil {
		t.Fatal("application pool can write migration metadata")
	} else {
		var pgError *pgconn.PgError
		if !errors.As(err, &pgError) || pgError.Code != "42501" {
			t.Fatalf("metadata denial=%v, want PostgreSQL42501", err)
		}
	}
	if err := db.Reset(t.Context(), "testdata/announcements.sql"); err != nil {
		t.Fatal(err)
	}
	connection, err := pool.Acquire(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Release()
	var count int
	err = connection.QueryRow(t.Context(), "select count(*) from announcements").Scan(&count)
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != "42704" {
		t.Fatalf("fresh unscoped read=%v, want PostgreSQL42704", err)
	}
	tx, err := connection.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	if _, err := tx.Exec(t.Context(), "select set_config('tadoku.tenant', 'tadoku/prod', true)"); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(t.Context(), "select count(*) from announcements").Scan(&count); err != nil || count != 2 {
		t.Fatalf("scoped committed fixture count=%d error=%v", count, err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := connection.QueryRow(t.Context(), "select count(*) from announcements").Scan(&count); err != nil || count != 0 {
		t.Fatalf("reused unscoped read count=%d error=%v", count, err)
	}
	_, err = connection.Exec(t.Context(), "insert into announcements (id, namespace, title, content, starts_at, ends_at) values (gen_random_uuid(), 'main', 'unscoped', '', now(), now())")
	connection.Release()
	if !errors.As(err, &pgError) || (pgError.Code != "42501" && (pgError.Code != "23503" || pgError.ConstraintName != "announcements_tenant_fkey")) {
		t.Fatalf("reused unscoped write=%v, want policy denial or exact tenant foreign-key denial", err)
	}
	if err := db.Pool.QueryRow(t.Context(), "select count(*) from announcements where title = 'unscoped'").Scan(&count); err != nil || count != 0 {
		t.Fatalf("unscoped write persisted: rows=%d error=%v", count, err)
	}
	observer, err := pgx.Connect(t.Context(), db.admin.Config().ConnString())
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := observer.Close(context.Background()); err != nil {
			t.Error(err)
		}
	}()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	closed = true
	if err := observer.QueryRow(t.Context(), "select (select count(*) from pg_roles where rolname = $1) + (select count(*) from pg_database where datname = $2)", role, db.name).Scan(&count); err != nil || count != 0 {
		t.Fatalf("owned role/database remain: count=%d error=%v", count, err)
	}
}

func TestCloseCleansPartialApplicationRoleSetup(t *testing.T) {
	complete, err := New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := complete.Close(); err != nil {
			t.Error(err)
		}
	})
	admin, err := pgxpool.New(t.Context(), complete.admin.Config().ConnString())
	if err != nil {
		t.Fatal(err)
	}
	role := "tadoku_test_partial_" + uuid.NewString()[:8]
	if _, err := admin.Exec(t.Context(), "create role "+pgx.Identifier{role}.Sanitize()+" login nosuperuser nobypassrls"); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	partial := &Database{admin: admin, appRole: role}
	if err := partial.Close(); err != nil {
		t.Fatal(err)
	}
	var remaining int
	if err := complete.admin.QueryRow(t.Context(), "select count(*) from pg_roles where rolname = $1", role).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatalf("partial setup role remains: count=%d error=%v", remaining, err)
	}
}
