package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"golang.org/x/crypto/bcrypt"
)

const canonicalTenant = "tadoku/prod"

var canonicalArgs = []string{
	"canonical-development-accounts",
	"--admin-id", developmentAdmin,
	"--reader-id", developmentReader,
}

func canonicalDatabase(t *testing.T) *testpostgres.Database {
	t.Helper()

	db := accessTokenDatabase(t)
	name := "fixture_runtime_" + uuid.NewString()
	role := pgx.Identifier{name}.Sanitize()
	_, err := db.Pool.Exec(t.Context(), "create role "+role+" login nosuperuser nobypassrls password 'fixture-only'")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, err := db.Pool.Exec(context.Background(), "drop owned by "+role+"; drop role "+role)
		if err != nil {
			t.Error(err)
		}
	})
	_, err = db.Pool.Exec(t.Context(), "grant select, insert, update, delete on all tables in schema public to "+role)
	if err != nil {
		t.Fatal(err)
	}

	t.Setenv("TENANT_POSTGRES_USER", name)
	t.Setenv("TENANT_POSTGRES_PASSWORD", "fixture-only")
	t.Setenv("TADOKU_DEV_ADMIN_PASSWORD", "tadoku")
	t.Setenv("TADOKU_DEV_READER_PASSWORD", "tadoku")

	return db
}

func TestCanonicalDevelopmentAccountsCommand(t *testing.T) {
	db := canonicalDatabase(t)
	args := canonicalArgs

	if err := run(t.Context(), args); err != nil {
		t.Fatalf("seed canonical native development accounts: %v", err)
	}
	assertDevelopmentFixtures(t, db, canonicalTenant, "tadoku")
	before := developmentSnapshot(t, db)
	for range 2 {
		if err := run(t.Context(), args); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(before, developmentSnapshot(t, db)) {
		t.Fatal("canonical rerun changed users, accounts or sessions")
	}

	if err := run(t.Context(), append(canonicalArgs, "--tenant", developmentTenant)); err == nil {
		t.Fatal("canonical seed accepted a tenant argument")
	}
	t.Setenv("TENANT_POSTGRES_HOST", "tadoku-db.production.example")
	if err := run(t.Context(), args); err == nil {
		t.Fatal("canonical seed accepted a non-development database host")
	}
	t.Setenv("TENANT_POSTGRES_HOST", "127.0.0.1")
	t.Setenv("TENANT_PRODUCTION", "true")
	if err := run(t.Context(), args); err == nil {
		t.Fatal("production accepted canonical development fixtures")
	}
	if !reflect.DeepEqual(before, developmentSnapshot(t, db)) {
		t.Fatal("denied canonical invocation changed data")
	}
}

func TestCanonicalDevelopmentAccountsKeepExistingFixtures(t *testing.T) {
	db := canonicalDatabase(t)
	hash, err := bcrypt.GenerateFromPassword([]byte("imported-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Pool.Exec(t.Context(), `insert into users (tenant, id, display_name, created_at, updated_at)
		values
			('tadoku/prod', $1, 'Imported Admin', now() - interval '3 years', now() - interval '1 year'),
			('tadoku/prod', $2, 'Legacy Reader', now() - interval '2 years', now() - interval '1 year')`,
		developmentAdmin,
		developmentReader,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Pool.Exec(t.Context(), `insert into accounts
		(tenant, user_id, email, email_normalized, password_hash, state, created_at, updated_at)
		values ('tadoku/prod', $1, 'Imported@Tadoku.app', 'imported@tadoku.app', $2, 'inactive',
			now() - interval '3 years', now() - interval '1 year')`,
		developmentAdmin,
		string(hash),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Pool.Exec(t.Context(), `insert into sessions
		(tenant, token_hash, user_id, authenticated_at, expires_at)
		values ('tadoku/prod', 'imported-session', $1, now(), now() + interval '1 day')`, developmentAdmin)
	if err != nil {
		t.Fatal(err)
	}
	before := developmentSnapshot(t, db)

	if err := run(t.Context(), canonicalArgs); err != nil {
		t.Fatalf("existing canonical fixtures blocked the seed: %v", err)
	}

	var accounts int
	var readerAge, readerLogin bool
	err = db.Pool.QueryRow(t.Context(), `select count(*),
			bool_and(u.id <> $1 or a.created_at = u.created_at at time zone 'UTC'),
			bool_and(u.id <> $1 or a.email = 'reader@tadoku.app')
		from accounts a
		join users u on (u.tenant, u.id) = (a.tenant, a.user_id)
		where a.tenant = 'tadoku/prod'`,
		developmentReader,
	).Scan(&accounts, &readerAge, &readerLogin)
	if err != nil || accounts != 2 || !readerAge || !readerLogin {
		t.Fatalf("missing reader account not initialized from its user: accounts=%d age=%t email=%t error=%v",
			accounts, readerAge, readerLogin, err)
	}
	_, err = db.Pool.Exec(t.Context(), "delete from accounts where tenant = 'tadoku/prod' and user_id = $1",
		developmentReader)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, developmentSnapshot(t, db)) {
		t.Fatal("canonical seed changed existing users, accounts, credentials, timestamps or sessions")
	}
}

func TestCanonicalDevelopmentAccountsEmailCollision(t *testing.T) {
	for _, email := range []string{"dev@tadoku.app", "reader@tadoku.app"} {
		t.Run(email, func(t *testing.T) {
			db := canonicalDatabase(t)
			id := uuid.NewString()
			_, err := db.Pool.Exec(t.Context(), `insert into users (tenant, id, display_name)
				values ('tadoku/prod', $1, 'Unrelated user')`, id)
			if err != nil {
				t.Fatal(err)
			}
			_, err = db.Pool.Exec(t.Context(), `insert into accounts
				(tenant, user_id, email, email_normalized, password_hash)
				values ('tadoku/prod', $1, $2, $2, 'unrelated-hash')`, id, email)
			if err != nil {
				t.Fatal(err)
			}
			before := developmentSnapshot(t, db)

			if err := run(t.Context(), canonicalArgs); err == nil {
				t.Fatal("email held by an unrelated canonical account reported success")
			}

			if !reflect.DeepEqual(before, developmentSnapshot(t, db)) {
				t.Fatal("failed canonical fixture pair committed partial data")
			}
		})
	}
}

func TestCanonicalDevelopmentAccountsTenantIsolation(t *testing.T) {
	db := canonicalDatabase(t)
	_, err := db.Pool.Exec(t.Context(), `insert into users (tenant, id, display_name)
		values ($1, $2, 'Branch Admin')`, developmentTenant, developmentAdmin)
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Pool.Exec(t.Context(), `insert into accounts
		(tenant, user_id, email, email_normalized, password_hash)
		values ($1, $2, 'dev@tadoku.app', 'dev@tadoku.app', 'branch-hash')`, developmentTenant, developmentAdmin)
	if err != nil {
		t.Fatal(err)
	}
	branch := func() string {
		var rows string
		err := db.Pool.QueryRow(t.Context(), `select jsonb_agg(item order by item::text)::text
			from (
				select to_jsonb(u) as item from users u where tenant = $1
				union all
				select to_jsonb(a) from accounts a where tenant = $1
			) branch_rows`,
			developmentTenant,
		).Scan(&rows)
		if err != nil {
			t.Fatal(err)
		}
		return rows
	}
	before := branch()

	if err := run(t.Context(), canonicalArgs); err != nil {
		t.Fatalf("branch account with the same UUID and email blocked the canonical seed: %v", err)
	}

	assertDevelopmentFixtures(t, db, canonicalTenant, "tadoku")
	var branchHash string
	err = db.Pool.QueryRow(t.Context(), "select password_hash from accounts where tenant = $1 and user_id = $2",
		developmentTenant, developmentAdmin).Scan(&branchHash)
	if err != nil || branchHash != "branch-hash" || branch() != before {
		t.Fatalf("canonical seed changed the branch tenant: error=%v", err)
	}
}
