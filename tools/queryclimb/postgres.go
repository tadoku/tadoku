package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/jackc/pgx/v5"
)

const localDSN = "postgres://postgres:queryclimb@127.0.0.1:55436/postgres?sslmode=disable"

func connect(ctx context.Context, raw, database, role string) (*pgx.Conn, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "postgres" || u.Hostname() != "127.0.0.1" || u.Port() != "55436" ||
		u.User == nil || u.User.Username() != "postgres" || u.Path != "/postgres" ||
		u.RawQuery != "sslmode=disable" || u.Fragment != "" {
		return nil, errors.New("only the owned query-climb loopback instance on port 55436 is allowed")
	}
	password, _ := u.User.Password()
	if password != "queryclimb" {
		return nil, errors.New("only synthetic query-climb credentials are allowed")
	}
	if database != "postgres" && !ownedDatabase(database) {
		return nil, errors.New("database must use the qc_ ownership prefix")
	}
	u.Path = "/" + database
	u.User = url.UserPassword(role, "queryclimb")
	cfg, err := pgx.ParseConfig(u.String())
	if err != nil {
		return nil, err
	}
	cfg.RuntimeParams["application_name"] = "query-climb"
	return pgx.ConnectConfig(ctx, cfg)
}

func initialize(ctx context.Context, root string, args []string) error {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	database := flags.String("name", "qc_schema", "new database; existing databases are never replaced")
	if err := flags.Parse(args); err != nil {
		return err
	}
	admin, err := connect(ctx, localDSN, "postgres", "postgres")
	if err != nil {
		return err
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, "create extension if not exists pg_stat_statements"); err != nil {
		return err
	}

	for _, role := range []string{"tadoku_owner", "tadoku"} {
		var exists bool
		err := admin.QueryRow(ctx, "select exists(select 1 from pg_roles where rolname = $1)", role).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			_, err := admin.Exec(ctx, "create role "+pgx.Identifier{role}.Sanitize()+
				" login nosuperuser nobypassrls password 'queryclimb'")
			if err != nil {
				return err
			}
		}
	}
	if !ownedDatabase(*database) {
		return errors.New("database must start qc_")
	}
	_, err = admin.Exec(ctx, "create database "+pgx.Identifier{*database}.Sanitize()+" owner tadoku_owner")
	if err != nil {
		return err
	}
	owner, err := connect(ctx, localDSN, *database, "tadoku_owner")
	if err != nil {
		return err
	}
	defer owner.Close(ctx)

	migrations, err := filepath.Glob(filepath.Join(root, "services/tadoku-api/migrations/*.up.sql"))
	if err != nil {
		return err
	}
	for _, path := range migrations {
		sql, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := owner.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("migration %s: %w", path, err)
		}
	}
	_, err = owner.Exec(ctx, `grant usage on schema public to tadoku;
  grant select, insert, update, delete on all tables in schema public to tadoku;
  grant usage, select on all sequences in schema public to tadoku`)
	if err != nil {
		return err
	}

	super, err := connect(ctx, localDSN, *database, "postgres")
	if err != nil {
		return err
	}
	defer super.Close(ctx)
	_, err = super.Exec(ctx, "create extension pg_stat_statements")
	if err != nil {
		return err
	}
	return output(map[string]any{"database": *database, "migrations": len(migrations)})
}

func clone(ctx context.Context, template, name string) error {
	admin, err := connect(ctx, localDSN, "postgres", "postgres")
	if err != nil {
		return err
	}
	defer admin.Close(ctx)
	if !ownedDatabase(template) || !ownedDatabase(name) {
		return errors.New("unowned database")
	}

	_, err = admin.Exec(ctx, "create database "+pgx.Identifier{name}.Sanitize()+
		" template "+pgx.Identifier{template}.Sanitize()+" owner tadoku_owner")
	if err != nil {
		return err
	}
	owner, err := connect(ctx, localDSN, name, "tadoku_owner")
	if err != nil {
		return err
	}
	defer owner.Close(ctx)
	rows, err := owner.Query(ctx, `select c.relname::text from pg_class c
      join pg_namespace n on n.oid = c.relnamespace
      where n.nspname = 'public' and c.relkind = 'r' order by c.relname`)
	if err != nil {
		return err
	}
	tables := []string{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			rows.Close()
			return err
		}
		tables = append(tables, table)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, table := range tables {
		_, err := owner.Exec(ctx, "alter table "+pgx.Identifier{table}.Sanitize()+" set (autovacuum_enabled = false)")
		if err != nil {
			return err
		}
	}
	return nil

}

func shell(ctx context.Context, root, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = root
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ownedDatabase(name string) bool {
	return len(name) <= 63 && regexp.MustCompile(`^qc_[a-z0-9_]+$`).MatchString(name)
}
