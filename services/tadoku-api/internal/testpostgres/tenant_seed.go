package testpostgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

// Test safety: the seeding login gets DML only on tables whose every policy binds both reads and writes to
// tadoku.tenant, which excludes the registry, overrides and shared reference tables.
const tenantScopedTablesSQL = `select c.oid::regclass::text
	from pg_class c
	join pg_namespace n on n.oid = c.relnamespace
	join pg_policy p on p.polrelid = c.oid
	where n.nspname = current_schema()
		and c.relkind in ('r', 'p')
		and c.relrowsecurity
	group by c.oid
	having bool_and(
		coalesce(pg_get_expr(p.polqual, p.polrelid), '') like '%tenant = current_setting(''tadoku.tenant''::text%'
		and coalesce(pg_get_expr(p.polwithcheck, p.polrelid), '') like '%tenant = current_setting(''tadoku.tenant''::text%'
	)
	order by 1`

const ownedSequencesSQL = `select distinct s.oid::regclass::text
	from pg_depend d
	join pg_class s on s.oid = d.objid and s.relkind = 'S'
	where d.refobjid = any($1::regclass[])
		and d.deptype in ('a', 'i')
	order by 1`

func (d *Database) openSeedPool(ctx context.Context, cfg *pgxpool.Config, role string) error {
	roleSQL := pgx.Identifier{role}.Sanitize()
	password := uuid.NewString()
	_, err := d.admin.Exec(ctx, "create role "+roleSQL+" login nosuperuser nobypassrls password '"+password+"'")
	if err != nil {
		return fmt.Errorf("create test seeding role: %w", err)
	}
	d.seedRole = role

	rows, err := d.Pool.Query(ctx, tenantScopedTablesSQL)
	if err != nil {
		return fmt.Errorf("list tenant-scoped tables: %w", err)
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return fmt.Errorf("list tenant-scoped tables: %w", err)
	}
	if len(tables) == 0 {
		return errors.New("no tenant-scoped tables to grant the test seeding role")
	}
	rows, err = d.Pool.Query(ctx, ownedSequencesSQL, tables)
	if err != nil {
		return fmt.Errorf("list tenant-scoped sequences: %w", err)
	}
	sequences, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return fmt.Errorf("list tenant-scoped sequences: %w", err)
	}

	batch := &pgx.Batch{}
	batch.Queue("grant usage on schema public to " + roleSQL)
	for _, table := range tables {
		batch.Queue("grant select, insert, update, delete on " + table + " to " + roleSQL)
	}
	for _, sequence := range sequences {
		batch.Queue("grant usage, select on sequence " + sequence + " to " + roleSQL)
	}
	if err := d.Pool.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("grant test seeding privileges: %w", err)
	}

	seedCfg := cfg.Copy()
	seedCfg.ConnConfig.User = role
	seedCfg.ConnConfig.Password = password
	seedCfg.MaxConns = 2
	d.seedPool, err = pgxpool.NewWithConfig(ctx, seedCfg)
	if err != nil {
		return fmt.Errorf("open test seeding pool: %w", err)
	}
	if err := d.seedPool.Ping(ctx); err != nil {
		return fmt.Errorf("ping test seeding pool: %w", err)
	}
	return nil
}

// SeedTenant runs seed in a transaction bound to key, as a login whose only privileges are DML on tenant-scoped tables,
// so row-level security confines every statement to key's rows and the registry, overrides, shared reference tables
// and migration metadata are refused. It fails if seed leaves the transaction bound to a different tenant or to all
// tenants.
func (d *Database) SeedTenant(ctx context.Context, key tenant.TestKey, seed func(context.Context, pgx.Tx) error) error {
	if key.String() == "" {
		return errors.New("tenant seeding requires a parsed test tenant")
	}

	return pgx.BeginFunc(ctx, d.seedPool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "select set_config('tadoku.tenant', $1, true)", key.String()); err != nil {
			return err
		}
		if err := seed(ctx, tx); err != nil {
			return err
		}

		var bound bool
		err := tx.QueryRow(ctx, `select current_setting('tadoku.tenant') = $1
			and coalesce(current_setting('tadoku.all_tenants', true), '') <> 'on'`,
			key.String(),
		).Scan(&bound)
		if err != nil {
			return err
		}
		if !bound {
			return errors.New("tenant seed changed its tenant binding")
		}
		return nil
	})
}
