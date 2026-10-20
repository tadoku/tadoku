package main_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
	"golang.org/x/crypto/bcrypt"
)

const firstID = "11111111-1111-4111-8111-111111111111"
const secondID = "22222222-2222-4222-8222-222222222222"

type fixture struct {
	source, target *testpostgres.Database
	binary         string
	hash           string
	created        time.Time
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	f := &fixture{}
	for _, db := range []**testpostgres.Database{&f.source, &f.target} {
		var err error
		*db, err = testpostgres.New(t.Context())
		if err != nil {
			t.Fatal("create disposable database failed")
		}
		owned := *db
		t.Cleanup(func() {
			if owned.Close() != nil {
				t.Error("dispose owned database failed")
			}
		})
	}

	f.sql(t, true, "drop schema public cascade; create schema public")
	f.sql(t, true, `grant usage on schema public to "`+f.source.AppPool.Config().ConnConfig.User+`"`)

	kratos, err := runfiles.Rlocation("kratos_v26_2_0/kratos")
	if err != nil {
		t.Fatal(err)
	}
	migration := exec.CommandContext(t.Context(), kratos, "migrate", "sql", "--read-from-env", "--yes")
	migration.Env = []string{"HOME=" + t.TempDir(), "DSN=" + f.source.DSN, "LOG_LEVEL=error"}
	if migration.Run() != nil {
		t.Fatal("pinned Kratos PostgreSQL migration failed")
	}
	f.sql(t, true, `insert into networks(id, created_at, updated_at) values ($1, now(), now())`, uuid.NewString())
	f.binary, err = runfiles.Rlocation(
		"tadoku/services/tadoku-api/cmd/tadoku-identity-import/tadoku-identity-import_/tadoku-identity-import",
	)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("fixture-password-only"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	f.hash = string(hash)
	f.created, err = time.Parse(time.RFC3339Nano, "2019-01-02T15:04:05.123456+05:30")
	if err != nil {
		t.Fatal(err)
	}
	f.created = f.created.UTC()

	for i, id := range []string{firstID, secondID} {
		f.sql(t, true, `insert into identities
			(id, nid, schema_id, traits, state, created_at, updated_at)
			select $1, id, 'user', jsonb_build_object('email', $2::text, 'display_name', $3::text),
			$4, $5, $5 from networks`,
			id, []string{"MixedCase@example.test", "second@example.test"}[i],
			[]string{"読書する人", "Second User"}[i], []string{"active", "inactive"}[i], f.created,
		)
		f.sql(t, true, `insert into identity_credentials
			(id, nid, identity_id, identity_credential_type_id, config, version, created_at, updated_at)
			select $1, i.nid, i.id, ct.id, jsonb_build_object('hashed_password', $2::text), 0, i.created_at, i.updated_at
			from identities i cross join identity_credential_types ct
			where i.id = $3 and ct.name = 'password'`, uuid.NewString(), f.hash, id,
		)
	}
	var identities, credentials int
	err = f.source.Pool.QueryRow(t.Context(), `select
		(select count(*) from identities), (select count(*) from identity_credentials)`).Scan(&identities, &credentials)
	if err != nil || identities != 2 || credentials != 2 {
		t.Fatalf(
			"source fixture requires two identities and credentials: identities=%d credentials=%d", identities, credentials,
		)
	}

	role := f.source.AppPool.Config().ConnConfig.User
	f.sql(t, true, "grant select on all tables in schema public to "+`"`+role+`"`)
	f.sql(t, false, `insert into tenants(key, kind) values ('e2e/import-other', 'test')`)
	f.sql(t, false, `insert into users (tenant, id, display_name)
		values ('e2e/import-other', $1, 'Other tenant')`, firstID)
	f.sql(t, false, `alter role "`+f.target.AppPool.Config().ConnConfig.User+`" set timezone to 'Asia/Tokyo'`)

	return f
}

func (f *fixture) sql(t *testing.T, source bool, sql string, args ...any) {
	t.Helper()

	db := f.target
	if source {
		db = f.source
	}
	if _, err := db.Pool.Exec(t.Context(), sql, args...); err != nil {
		t.Fatal("fixture SQL failed (values suppressed)")
	}
}

func (f *fixture) command(t *testing.T, args ...string) (string, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, f.binary, args...)
	command.Env = []string{"HOME=" + t.TempDir()}
	for prefix, db := range map[string]*testpostgres.Database{
		"IMPORT_SOURCE_POSTGRES": f.source,
		"IMPORT_TARGET_POSTGRES": f.target,
	} {
		cfg := db.AppPool.Config().ConnConfig
		for key, value := range map[string]string{
			"HOST": cfg.Host, "PORT": strconv.Itoa(int(cfg.Port)), "DATABASE": cfg.Database,
			"USER": cfg.User, "PASSWORD": cfg.Password, "SSLMODE": "disable",
		} {
			command.Env = append(command.Env, prefix+"_"+key+"="+value)
		}
	}
	out, err := command.CombinedOutput()
	text := string(out)
	for _, secret := range []string{
		f.hash, "fixture-password-only", "postgres://", "MixedCase@example.test",
		f.source.AppPool.Config().ConnConfig.Password, f.target.AppPool.Config().ConnConfig.Password,
	} {
		if strings.Contains(text, secret) {
			t.Fatal("CLI output exposed a protected value (output suppressed)")
		}
	}
	if ctx.Err() != nil {
		t.Fatal("CLI exceeded bounded execution deadline")
	}
	t.Logf("CLI arguments=%q\n%s", args, text)
	return text, err
}

func (f *fixture) snapshot(t *testing.T) []string {
	t.Helper()

	var result []string
	for _, entry := range []struct {
		db     *testpostgres.Database
		tables []string
	}{
		{f.source, []string{"identities", "identity_credentials"}},
		{f.target, []string{"users", "accounts", "sessions"}},
	} {
		for _, table := range entry.tables {
			var value string
			err := entry.db.Pool.QueryRow(t.Context(),
				"select coalesce(jsonb_agg(to_jsonb(t) order by to_jsonb(t)::text)::text, '[]') from "+table+" t",
			).Scan(&value)
			if err != nil {
				t.Fatal("snapshot failed")
			}
			result = append(result, fmt.Sprintf("%x", sha256.Sum256([]byte(value))))
		}
	}
	return result
}

func (f *fixture) unchanged(t *testing.T, before []string) {
	t.Helper()

	if !reflect.DeepEqual(before, f.snapshot(t)) {
		t.Fatal("command changed source, target or unrelated tenant data")
	}
}
