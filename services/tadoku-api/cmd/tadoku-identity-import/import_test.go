package main_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func TestImportPreservationDryRunAndRepeat(t *testing.T) {
	f := newFixture(t)
	before := f.snapshot(t)

	out, err := f.command(t, "--dry-run")
	if err != nil || !strings.Contains(out, "source_identities=2") ||
		!strings.Contains(out, "insert_accounts=2") || !strings.Contains(out, "dry_run=true") {
		t.Fatalf("dry run did not report proposed import: error=%v output=%s", err, out)
	}
	f.unchanged(t, before)

	out, err = f.command(t, "--apply", "--expect-count", "2")
	if err != nil || !strings.Contains(out, "committed_accounts=2") {
		t.Fatalf("apply failed: error=%v output=%s", err, out)
	}
	var preserved bool
	err = f.target.Pool.QueryRow(t.Context(), `select count(*) = 2 and bool_and(
		a.email = case a.user_id when $1::uuid then 'MixedCase@example.test' else 'second@example.test' end
		and a.email_normalized = lower(a.email) and a.password_hash = $2
		and a.created_at = $3 and u.created_at at time zone 'UTC' = $3
		and a.state = case a.user_id when $1::uuid then 'active' else 'inactive' end
		and u.display_name = case a.user_id when $1::uuid then '読書する人' else 'Second User' end)
		from accounts a join users u on (u.tenant, u.id) = (a.tenant, a.user_id)
		where a.tenant = 'tadoku/prod'`, firstID, f.hash, f.created).Scan(&preserved)
	if err != nil || !preserved {
		t.Fatal("import did not preserve identity fields, state or UTC creation timestamps")
	}
	var otherName string
	err = f.target.Pool.QueryRow(t.Context(), `select display_name from users
		where tenant = 'e2e/import-other' and id = $1`, firstID).Scan(&otherName)
	if err != nil || otherName != "Other tenant" {
		t.Fatal("import changed another tenant's colliding UUID")
	}
	f.sql(t, false, `insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at)
		values ('tadoku/prod', 'fixture-session-marker', $1, now(), now() + interval '1 day')`, firstID)
	before = f.snapshot(t)

	for _, args := range [][]string{nil, {"--apply", "--expect-count", "2"}} {
		out, err := f.command(t, args...)
		if err != nil || !strings.Contains(out, "unchanged_accounts=2") {
			t.Fatalf("repeat failed: error=%v output=%s", err, out)
		}
		f.unchanged(t, before)
	}
}

func TestImportPreflight(t *testing.T) {
	for _, tc := range []struct {
		name, sql, field string
	}{
		{
			"missing-password",
			"delete from identity_credentials where identity_id = '" + secondID + "'",
			"password_credentials",
		},
		{"short-hash", `update identity_credentials set config = '{"hashed_password":"invalid"}'`, "password_hash"},
		{"hash-alphabet", `update identity_credentials set config = jsonb_build_object(
			'hashed_password', '$2a$04$' || repeat('!',53))`, "password_hash"},
		{
			"migration-hook",
			`update identity_credentials set config = config || '{"use_password_migration_hook":true}'`,
			"password_hash",
		},
		{"duplicate-email", `update identities set traits = jsonb_set(traits, '{email}', '"mixedcase@example.test"')
			where id = '` + secondID + "'", "email_normalized"},
		{"invalid-email", `update identities set traits = jsonb_set(traits, '{email}', '"invalid"')`, "email"},
		{"short-name", `update identities set traits = jsonb_set(traits, '{display_name}', '"a"')`, "display_name"},
		{
			"long-name",
			`update identities set traits = jsonb_set(traits, '{display_name}', to_jsonb(repeat('界',33)))`,
			"display_name",
		},
		{"blank-name", `update identities set traits = jsonb_set(traits, '{display_name}', '"  "')`, "display_name"},
		{"invalid-trait-type", `update identities set traits = jsonb_set(traits, '{email}', '42')`, "traits"},
		{"multiple-networks", `insert into networks(id, created_at, updated_at)
			values ('33333333-3333-4333-8333-333333333333', now(), now())`, "network"},
		{"additional-credential", `insert into identity_credentials
			(id, nid, identity_id, identity_credential_type_id, config, version, created_at, updated_at)
			select '44444444-4444-4444-8444-444444444444', i.nid, i.id, ct.id, '{}', 0, now(), now()
			from identities i cross join identity_credential_types ct
			where i.id = '` + secondID + `' and ct.name = 'totp'`, "credential_types"},
		{"empty-source", "delete from identities", "source_count"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			f.sql(t, true, tc.sql)
			before := f.snapshot(t)

			out, err := f.command(t, "--apply", "--expect-count", "2")
			if err == nil || !strings.Contains(out, tc.field) {
				t.Fatalf("preflight did not reject %s: error=%v output=%s", tc.field, err, out)
			}
			f.unchanged(t, before)
		})
	}
}

func TestImportCountAndArgumentGuards(t *testing.T) {
	f := newFixture(t)
	before := f.snapshot(t)
	for _, args := range [][]string{
		{"--apply"},
		{"--apply", "--expect-count", "3"},
		{"--apply", "--dry-run", "--expect-count", "2"},
		{"--tenant", "e2e/other"},
		{"--expect-count", "-1"},
	} {
		if _, err := f.command(t, args...); err == nil {
			t.Fatal("unsafe invocation accepted")
		}
		f.unchanged(t, before)
	}
}

func TestImportTargetConflicts(t *testing.T) {
	for _, field := range []string{
		"email", "password_hash", "created_at", "display_name", "user-created-at", "deleted_at", "locked", "foreign-email",
	} {
		t.Run(field, func(t *testing.T) {
			f := newFixture(t)
			out, err := f.command(t, "--apply", "--expect-count", "2")
			if err != nil {
				t.Fatalf("initial import failed: %s", out)
			}
			f.sql(t, false, "delete from users where tenant = 'tadoku/prod' and id = $1", firstID)
			switch field {
			case "email":
				f.sql(t, false, `update accounts set email = 'changed@example.test' where user_id = $1`, secondID)
			case "password_hash":
				f.sql(t, false, `update accounts set password_hash = 'changed' where user_id = $1`, secondID)
			case "created_at":
				f.sql(t, false, `update accounts set created_at = now() where user_id = $1`, secondID)
			case "display_name":
				f.sql(t, false, `update users set display_name = 'Changed' where id = $1`, secondID)
			case "user-created-at":
				f.sql(t, false, `update users set created_at = now() where id = $1`, secondID)
			case "deleted_at":
				f.sql(t, false, `update users set deleted_at = now() where id = $1`, secondID)
			case "locked":
				f.sql(t, false, `update users set deletion_locked_at = now() where id = $1`, secondID)
			case "foreign-email":
				f.sql(t, false, `update accounts set email_normalized = 'mixedcase@example.test' where user_id = $1`, secondID)
			}
			before := f.snapshot(t)

			for _, args := range [][]string{{"--dry-run"}, {"--apply", "--expect-count", "2"}} {
				out, err := f.command(t, args...)
				if err == nil || !strings.Contains(out, "conflict") {
					t.Fatalf("target conflict was accepted: error=%v output=%s", err, out)
				}
				f.unchanged(t, before)
			}
		})
	}
}

func TestImportExistingUsersAndAtomicFailure(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "existing-user", true: "late-write-failure"}[fail], func(t *testing.T) {
			f := newFixture(t)
			f.sql(t, false, `insert into users (tenant, id, display_name, created_at, updated_at)
				values ('tadoku/prod', $1, '読書する人', $2::timestamptz at time zone 'UTC', '2024-01-01')`,
				firstID, f.created)
			if fail {
				f.sql(t, false, `alter table accounts add constraint fixture_reject_second
					check (user_id <> '`+secondID+`')`)
			}
			before := f.snapshot(t)

			out, err := f.command(t, "--apply", "--expect-count", "2")
			if fail {
				if err == nil || !strings.Contains(out, "target write") {
					t.Fatalf("late SQL failure did not abort: error=%v output=%s", err, out)
				}
				f.unchanged(t, before)
			} else if err != nil || !strings.Contains(out, "committed_accounts=2") {
				t.Fatalf("matching existing user rejected: error=%v output=%s", err, out)
			} else {
				var unchanged bool
				err := f.target.Pool.QueryRow(t.Context(), `select updated_at = timestamp '2024-01-01'
					from users where tenant = 'tadoku/prod' and id = $1`, firstID).Scan(&unchanged)
				if err != nil || !unchanged {
					t.Fatal("matching legacy user's metadata was overwritten")
				}
			}
		})
	}
}

func TestImportSourceSchemaFailureIsSafe(t *testing.T) {
	f := newFixture(t)
	f.sql(t, true, `alter table identity_credentials rename column config to unexpected_config`)
	before := f.snapshot(t)

	out, err := f.command(t, "--apply", "--expect-count", "2")
	if err == nil || !strings.Contains(out, "source read") {
		t.Fatalf("unrecognized source schema accepted: error=%v output=%s", err, out)
	}
	f.unchanged(t, before)
}

func TestImportConcurrentApply(t *testing.T) {
	f := newFixture(t)
	var wg sync.WaitGroup
	for range 2 {
		wg.Go(func() {
			out, err := f.command(t, "--apply", "--expect-count", "2")
			if err != nil {
				t.Errorf("concurrent apply failed: error=%v output=%s", err, out)
			}
		})
	}
	wg.Wait()
	var count int
	err := f.target.Pool.QueryRow(t.Context(), "select count(*) from accounts where tenant = 'tadoku/prod'").Scan(&count)
	if err != nil || count != 2 {
		t.Fatal("concurrent import lost or duplicated accounts")
	}
}

func TestImportExtraTargetAccount(t *testing.T) {
	f := newFixture(t)
	id := uuid.NewString()
	f.sql(t, false, `insert into users (tenant, id, display_name) values ('tadoku/prod', $1, 'Extra User')`, id)
	f.sql(t, false, `insert into accounts (tenant, user_id, email, email_normalized, password_hash)
		values ('tadoku/prod', $1, 'extra@example.test', 'extra@example.test', $2)`, id, f.hash)
	before := f.snapshot(t)

	out, err := f.command(t, "--apply", "--expect-count", "2")
	if err == nil || !strings.Contains(out, "conflict") {
		t.Fatalf("extra target account accepted: error=%v output=%s", err, out)
	}
	f.unchanged(t, before)
}
