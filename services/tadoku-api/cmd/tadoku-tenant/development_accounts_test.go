package main

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"golang.org/x/crypto/bcrypt"
)

const (
	developmentTenant      = "e2e/tokens-0123abcd"
	otherDevelopmentTenant = "e2e/other-0123abcd"
	developmentAdmin       = "edda0741-cf0c-4b71-a540-472c9726b1b5"
	developmentReader      = "0d141891-b6f8-4150-9d2f-6272d68b3db0"
)

func developmentDatabase(t *testing.T) *testpostgres.Database {
	t.Helper()

	db := accessTokenDatabase(t)
	name := "fixture_owner_" + uuid.NewString()
	role := pgx.Identifier{name}.Sanitize()
	_, err := db.Pool.Exec(t.Context(), "create role "+role+" login nosuperuser nobypassrls password 'fixture-only'")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, err := db.Pool.Exec(context.Background(), "reassign owned by "+role+" to postgres; drop role "+role)
		if err != nil {
			t.Error(err)
		}
	})
	for _, table := range []string{"tenants", "users", "accounts", "sessions"} {
		_, err := db.Pool.Exec(t.Context(), "alter table "+table+" owner to "+role)
		if err != nil {
			t.Fatal(err)
		}
	}

	var ownerBypasses bool
	err = db.Pool.QueryRow(t.Context(), `select not r.rolsuper and not r.rolbypassrls
		and c.relrowsecurity and not c.relforcerowsecurity
		from pg_class c
		join pg_roles r on r.oid = c.relowner
		where c.oid = 'accounts'::regclass and r.rolname = $1`,
		name,
	).Scan(&ownerBypasses)
	if err != nil || !ownerBypasses {
		t.Fatalf("fixture must use a non-superuser relation owner bypassing RLS: %v", err)
	}
	t.Setenv("TENANT_POSTGRES_USER", name)
	t.Setenv("TENANT_POSTGRES_PASSWORD", "fixture-only")
	t.Setenv("TADOKU_DEV_ADMIN_PASSWORD", "tadoku")
	t.Setenv("TADOKU_DEV_READER_PASSWORD", "tadoku")

	return db
}

func developmentArgs(key string) []string {
	return []string{
		"development-accounts",
		"--tenant", key,
		"--admin-id", developmentAdmin,
		"--reader-id", developmentReader,
	}
}

func developmentSnapshot(t *testing.T, db *testpostgres.Database) []string {
	t.Helper()

	var out []string
	for _, table := range []string{"users", "accounts", "sessions"} {
		var rows string
		err := db.Pool.QueryRow(t.Context(),
			"select coalesce(jsonb_agg(to_jsonb(t) order by to_jsonb(t)::text)::text, '[]') from "+table+" t",
		).Scan(&rows)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, rows)
	}
	return out
}

func assertDevelopmentFixtures(t *testing.T, db *testpostgres.Database, key, password string) {
	t.Helper()

	for _, fixture := range []struct{ id, email, name string }{
		{developmentAdmin, "dev@tadoku.app", "Dev Admin"},
		{developmentReader, "reader@tadoku.app", "Dev Reader"},
	} {
		var email, name, hash string
		err := db.Pool.QueryRow(t.Context(), `select a.email, u.display_name, a.password_hash
			from accounts a
			join users u on (u.tenant, u.id) = (a.tenant, a.user_id)
			where a.tenant = $1 and a.user_id = $2`,
			key,
			fixture.id,
		).Scan(&email, &name, &hash)
		if err != nil || email != fixture.email || name != fixture.name ||
			bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
			t.Fatalf("native fixture %s in %s has wrong identity or password (error=%v)", fixture.name, key, err)
		}
	}
}

func TestDevelopmentAccountsCommand(t *testing.T) {
	db := developmentDatabase(t)
	args := developmentArgs(developmentTenant)

	if err := run(t.Context(), args); err != nil {
		t.Fatalf("seed native development accounts: %v", err)
	}
	assertDevelopmentFixtures(t, db, developmentTenant, "tadoku")
	before := developmentSnapshot(t, db)
	for range 2 {
		if err := run(t.Context(), args); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(before, developmentSnapshot(t, db)) {
		t.Fatal("rerun changed existing users, accounts or sessions")
	}

	for _, blocked := range []string{"tadoku/prod", "e2e/missing-0123abcd"} {
		if err := run(t.Context(), developmentArgs(blocked)); err == nil {
			t.Fatalf("seed accepted %s", blocked)
		}
	}
	t.Setenv("TENANT_PRODUCTION", "true")
	if err := run(t.Context(), args); err == nil {
		t.Fatal("production accepted development fixtures")
	}
	if !reflect.DeepEqual(before, developmentSnapshot(t, db)) {
		t.Fatal("denied invocation changed data")
	}
}

func TestDevelopmentAccountsOwnerIsolation(t *testing.T) {
	db := developmentDatabase(t)
	t.Setenv("TADOKU_DEV_ADMIN_PASSWORD", "other-branch-password")
	t.Setenv("TADOKU_DEV_READER_PASSWORD", "other-branch-password")
	if err := run(t.Context(), developmentArgs(otherDevelopmentTenant)); err != nil {
		t.Fatal(err)
	}
	otherBefore := developmentSnapshot(t, db)
	t.Setenv("TADOKU_DEV_ADMIN_PASSWORD", "tadoku")
	t.Setenv("TADOKU_DEV_READER_PASSWORD", "tadoku")

	if err := run(t.Context(), developmentArgs(developmentTenant)); err != nil {
		t.Fatalf("other tenant's changed password blocked a fresh target: %v", err)
	}
	assertDevelopmentFixtures(t, db, developmentTenant, "tadoku")
	assertDevelopmentFixtures(t, db, otherDevelopmentTenant, "other-branch-password")
	_, err := db.Pool.Exec(t.Context(), "delete from tenants where key = $1", developmentTenant)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(otherBefore, developmentSnapshot(t, db)) {
		t.Fatal("target provisioning changed another tenant")
	}
}

func TestDevelopmentAccountsEmailCollision(t *testing.T) {
	for _, email := range []string{"dev@tadoku.app", "reader@tadoku.app"} {
		t.Run(email, func(t *testing.T) {
			db := developmentDatabase(t)
			if err := run(t.Context(), developmentArgs(otherDevelopmentTenant)); err != nil {
				t.Fatal(err)
			}
			id := uuid.NewString()
			_, err := db.Pool.Exec(t.Context(), `insert into users (tenant, id, display_name)
				values ($1, $2, 'Branch-owned user')`, developmentTenant, id)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Pool.Exec(t.Context(), `insert into accounts
				(tenant, user_id, email, email_normalized, password_hash)
				values ($1, $2, $3, $3, 'branch-owned-hash')`, developmentTenant, id, email)
			if err != nil {
				t.Fatal(err)
			}
			before := developmentSnapshot(t, db)

			if err := run(t.Context(), developmentArgs(developmentTenant)); err == nil {
				t.Fatal("email held by another target UUID reported false success")
			}
			if !reflect.DeepEqual(before, developmentSnapshot(t, db)) {
				t.Fatal("failed fixture pair committed partial data")
			}
		})
	}
}

func TestDevelopmentAccountsPreserveBranchChanges(t *testing.T) {
	for _, partial := range []bool{false, true} {
		t.Run(map[bool]string{false: "both-existing", true: "one-missing"}[partial], func(t *testing.T) {
			db := developmentDatabase(t)
			if err := run(t.Context(), developmentArgs(developmentTenant)); err != nil {
				t.Fatal(err)
			}
			hash, err := bcrypt.GenerateFromPassword([]byte("branch-owned-password"), bcrypt.MinCost)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Pool.Exec(t.Context(), `update accounts
				set email = 'changed@example.com', email_normalized = 'changed@example.com',
				password_hash = $1, state = 'inactive', updated_at = now() - interval '1 day'
				where tenant = $2 and user_id = $3`, string(hash), developmentTenant, developmentAdmin)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Pool.Exec(t.Context(), `update users
				set display_name = 'Branch name', deleted_at = now(), deletion_locked_at = now()
				where tenant = $1 and id = $2`, developmentTenant, developmentAdmin)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Pool.Exec(t.Context(), `insert into sessions
				(tenant, token_hash, user_id, authenticated_at, expires_at)
				values ($1, 'owned-session', $2, now(), now() + interval '1 day')`,
				developmentTenant, developmentAdmin)
			if err != nil {
				t.Fatal(err)
			}
			if partial {
				_, err = db.Pool.Exec(t.Context(), "delete from users where tenant = $1 and id = $2",
					developmentTenant, developmentReader)
				if err != nil {
					t.Fatal(err)
				}
			}
			before := developmentSnapshot(t, db)

			if err := run(t.Context(), developmentArgs(developmentTenant)); err != nil {
				t.Fatalf("branch changes blocked reseed: %v", err)
			}
			if partial {
				var count int
				err := db.Pool.QueryRow(t.Context(), "select count(*) from accounts where tenant = $1",
					developmentTenant).Scan(&count)
				if err != nil || count != 2 {
					t.Fatalf("missing sibling not initialized: count=%d error=%v", count, err)
				}
				_, err = db.Pool.Exec(t.Context(), "delete from users where tenant = $1 and id = $2",
					developmentTenant, developmentReader)
				if err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(before, developmentSnapshot(t, db)) {
				t.Fatal("reseed changed branch-owned metadata, credentials, state, timestamps or sessions")
			}
		})
	}
}

func TestDevelopmentAccountsConcurrentSeed(t *testing.T) {
	db := developmentDatabase(t)
	var wg sync.WaitGroup
	results := make(chan error, 4)
	for range 4 {
		wg.Go(func() { results <- run(t.Context(), developmentArgs(developmentTenant)) })
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	assertDevelopmentFixtures(t, db, developmentTenant, "tadoku")
	var count int
	err := db.Pool.QueryRow(t.Context(), "select count(*) from accounts where tenant = $1",
		developmentTenant).Scan(&count)
	if err != nil || count != 2 {
		t.Fatalf("concurrent seeds: accounts=%d error=%v", count, err)
	}
}

func TestDevelopmentAccountsCanonicalAge(t *testing.T) {
	db := developmentDatabase(t)
	_, err := db.Pool.Exec(t.Context(), `insert into users (tenant, id, display_name, created_at)
		values ('tadoku/prod', $1, 'Dev Reader', now() - interval '2 months')`, developmentReader)
	if err != nil {
		t.Fatal(err)
	}

	if err := run(t.Context(), developmentArgs(developmentTenant)); err != nil {
		t.Fatal(err)
	}
	var sameAge bool
	err = db.Pool.QueryRow(t.Context(), `select a.created_at = canonical.created_at at time zone 'UTC'
		and a.created_at < now() - interval '1 month'
		from accounts a
		join users canonical on canonical.tenant = 'tadoku/prod' and canonical.id = a.user_id
		where a.tenant = $1 and a.user_id = $2`, developmentTenant, developmentReader).Scan(&sameAge)
	if err != nil || !sameAge {
		t.Fatalf("new branch reader lost canonical fixture age: %v", err)
	}
}
