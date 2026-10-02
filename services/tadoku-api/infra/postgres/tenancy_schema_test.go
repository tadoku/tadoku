package postgres_test

import (
	"testing"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func TestTenancySchemaGuard(t *testing.T) {
	db, err := testpostgres.New(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	var violations []string
	err = db.Pool.QueryRow(t.Context(), `
		with ordinary as (
			select c.oid, c.relname, c.relrowsecurity, a.attnotnull,
				pg_get_expr(d.adbin, d.adrelid) as tenant_default,
				case when c.relname = 'jobs' then
					'((tenant = current_setting(''tadoku.tenant''::text, true)) OR (current_setting(''tadoku.all_tenants''::text, true) = ''on''::text))'
				else '(tenant = current_setting(''tadoku.tenant''::text))' end as policy_expression
			from pg_class c
			join pg_namespace n on n.oid = c.relnamespace
			left join pg_attribute a on a.attrelid = c.oid and a.attname = 'tenant' and not a.attisdropped
			left join pg_attrdef d on d.adrelid = c.oid and d.adnum = a.attnum
			where n.nspname = current_schema() and c.relkind in ('r', 'p')
				and c.relname not in ('tenants', 'tenant_overrides', 'languages', 'log_units', 'schema_migrations')
				and not starts_with(c.relname, 'tadoku_test_')
		), shared_policy(name, command, using_expression, check_expression) as (
			values ('shared_read', 'SELECT', 'true', null),
				('production_insert', 'INSERT', null, '(current_setting(''tadoku.tenant''::text, true) = ''tadoku/prod''::text)'),
				('production_update', 'UPDATE', '(current_setting(''tadoku.tenant''::text, true) = ''tadoku/prod''::text)', null),
				('production_delete', 'DELETE', '(current_setting(''tadoku.tenant''::text, true) = ''tadoku/prod''::text)', null)
		), violations as (
			select 'ordinary table: ' || o.relname as violation from ordinary o
			where not o.relrowsecurity or not coalesce(o.attnotnull, false)
				or o.tenant_default is distinct from 'current_setting(''tadoku.tenant''::text)'
				or (select count(*) from pg_policies p where p.schemaname = current_schema() and p.tablename = o.relname) <> 1
				or not exists (select 1 from pg_policies p
					where p.schemaname = current_schema() and p.tablename = o.relname
						and p.policyname = 'tenant_isolation' and p.cmd = 'ALL'
						and p.permissive = 'PERMISSIVE' and p.roles = array['public']::name[]
						and p.qual = o.policy_expression and p.with_check = o.policy_expression)
			union all
			select 'shared table: ' || shared.table_name from unnest(array['languages', 'log_units']) shared(table_name)
			where not exists (select 1 from pg_class c join pg_namespace n on n.oid = c.relnamespace
				where n.nspname = current_schema() and c.relname = shared.table_name and c.relrowsecurity)
				or (select count(*) from pg_policies p where p.schemaname = current_schema() and p.tablename = shared.table_name) <> 4
				or exists (select 1 from shared_policy expected where not exists (
					select 1 from pg_policies p where p.schemaname = current_schema() and p.tablename = shared.table_name
						and p.policyname = expected.name and p.cmd = expected.command
						and p.permissive = 'PERMISSIVE' and p.roles = array['public']::name[]
						and p.qual is not distinct from expected.using_expression
						and p.with_check is not distinct from expected.check_expression))
			union all
			select 'startup tenant setting: ' || split_part(setting, '=', 1)
			from pg_db_role_setting cross join lateral unnest(setconfig) settings(setting)
			where starts_with(setting, 'tadoku.')
		)
		select coalesce(array_agg(violation order by violation), array[]::text[]) from violations
	`).Scan(&violations)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Fatalf("unprotected tenancy schema: %v", violations)
	}
}
