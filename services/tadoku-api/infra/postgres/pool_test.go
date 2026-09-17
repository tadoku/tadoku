package postgres_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestPoolWorksWithOnlyAnnouncementReadGrants(t *testing.T) {
	t.Parallel()

	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roleName := "tadoku_reader_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = db.Pool.Exec(ctx, `
		create role "`+roleName+`"
		login password 'synthetic-reader'
		nosuperuser nocreatedb nocreaterole
	`)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if _, err := db.Pool.Exec(ctx, `drop owned by "`+roleName+`"`); err != nil {
			t.Errorf("drop reader grants: %v", err)
		}

		if _, err := db.Pool.Exec(ctx, `drop role "`+roleName+`"`); err != nil {
			t.Errorf("drop reader role: %v", err)
		}
	})

	for _, statement := range []string{
		`grant usage on schema public to "` + roleName + `"`,
		`grant select on announcements to "` + roleName + `"`,
	} {
		if _, err := db.Pool.Exec(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}

	dsn, err := url.Parse(db.DSN)
	if err != nil {
		t.Fatal(err)
	}
	dsn.User = url.UserPassword(roleName, "synthetic-reader")

	pool, err := postgres.Open(ctx, dsn.String(), 2)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	if pool.Config().MaxConns != 2 {
		t.Errorf("max connections=%d", pool.Config().MaxConns)
	}
	if pool.Config().MinConns != 1 {
		t.Errorf("min connections=%d", pool.Config().MinConns)
	}
	if mode := pool.Config().ConnConfig.DefaultQueryExecMode; mode != pgx.QueryExecModeCacheStatement {
		t.Errorf("query execution mode=%v, want cache statement", mode)
	}

	var count int
	if err := pool.QueryRow(ctx, "select count(*) from announcements").Scan(&count); err != nil {
		t.Fatal(err)
	}

	_, err = pool.Exec(ctx, "delete from announcements")
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "42501" {
		t.Errorf("reader mutation error=%v, want permission denied", err)
	}
}

func TestPoolRejectsInvalidLimitAndCanceledStartup(t *testing.T) {
	if _, err := postgres.Open(context.Background(), "unused", 0); err == nil {
		t.Error("accepted a zero connection limit")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := postgres.Open(ctx, "postgres://postgres:postgres@127.0.0.1:1/postgres?sslmode=disable", 1)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("canceled startup=%v", err)
	}
}
