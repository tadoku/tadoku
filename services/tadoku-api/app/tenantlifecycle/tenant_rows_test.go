package tenantlifecycle_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/testpostgres"
)

func ownedRows(t *testing.T, db *testpostgres.Database, key string) map[string]string {
	t.Helper()
	rows, err := db.Pool.Query(t.Context(), `
		select c.relname
		from pg_class c
		join pg_namespace n on n.oid = c.relnamespace
		join pg_attribute a on a.attrelid = c.oid and a.attname = 'tenant' and not a.attisdropped
		where n.nspname = current_schema()
			and c.relkind in ('r', 'p')
			and not starts_with(c.relname, 'tadoku_test_')
		order by c.relname`)
	if err != nil {
		t.Fatal(err)
	}
	tables, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil || len(tables) == 0 {
		t.Fatalf("tenant-owned inventory=%v error=%v", tables, err)
	}
	result := make(map[string]string, len(tables))
	for _, table := range tables {
		var value string
		query := "select coalesce(jsonb_agg(to_jsonb(row) order by to_jsonb(row)::text), '[]'::jsonb)::text" +
			" from " + pgx.Identifier{table}.Sanitize() + " row where tenant = $1"
		if err := db.Pool.QueryRow(t.Context(), query, key).Scan(&value); err != nil {
			t.Fatal(err)
		}
		result[table] = value
	}
	encoded, _ := json.Marshal(result)
	t.Logf("tenant table snapshot %s=%s", key, encoded)
	return result
}

func seedEveryTenantTable(t *testing.T, db *testpostgres.Database, key string) {
	t.Helper()
	seed := `
		select set_config('tadoku.tenant', $1, true);
		insert into users (tenant, id, display_name)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'), 'Lifecycle user');
		insert into user_roles (tenant, user_id, role)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'), 'admin');
		insert into profiles (tenant, user_id)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'));
		insert into account_deletion_requests (tenant, identity_id, status, accepted_at)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'), 'receipt_pending', now());
		insert into moderation_audit_log (tenant, user_id, action)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'), 'lifecycle-fixture');
		insert into contests (tenant, id, owner_user_id, owner_user_display_name, private,
			contest_start, contest_end, registration_end, title, activity_type_id_allow_list, official)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/contest'), uuid_generate_v5(uuid_nil(), $1 || '/user'),
			'Lifecycle user', false, current_date, current_date + 7, current_date + 7, 'Lifecycle', array[1], false);
		insert into contest_registrations (tenant, contest_id, user_id, language_codes)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/contest'), uuid_generate_v5(uuid_nil(), $1 || '/user'), array['jpn']);
		insert into logs (tenant, id, user_id, language_code, log_activity_id, unit_id, unit_key,
			amount, modifier, computed_score, eligible_official_leaderboard)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/log'), uuid_generate_v5(uuid_nil(), $1 || '/user'),
			'jpn', 1, (select id from log_units where unit_key = 'reading_page' limit 1), 'reading_page', 10, 1, 10, true);
		insert into contest_logs (tenant, contest_id, log_id, unit_key, amount, modifier, computed_score)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/contest'), uuid_generate_v5(uuid_nil(), $1 || '/log'),
			'reading_page', 10, 1, 10);
		insert into log_tags (tenant, log_id, user_id, tag)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/log'), uuid_generate_v5(uuid_nil(), $1 || '/user'), 'lifecycle');
		insert into pages (tenant, id, namespace, slug, current_content_id)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/page'), 'tadoku', 'lifecycle',
			uuid_generate_v5(uuid_nil(), $1 || '/page-content'));
		insert into pages_content (tenant, id, page_id, title, html)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/page-content'), uuid_generate_v5(uuid_nil(), $1 || '/page'),
			'Lifecycle', '<p>Fixture</p>');
		insert into posts (tenant, id, namespace, slug, current_content_id)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/post'), 'tadoku', 'lifecycle',
			uuid_generate_v5(uuid_nil(), $1 || '/post-content'));
		insert into posts_content (tenant, id, post_id, title, content)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/post-content'), uuid_generate_v5(uuid_nil(), $1 || '/post'),
			'Lifecycle', 'Fixture');
		insert into announcements (tenant, namespace, title, content, starts_at, ends_at)
		values ($1, 'tadoku', 'Lifecycle', 'Fixture', now(), now() + interval '1 day');
		insert into scoring_rule_sets (tenant, id, scope, version, status)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/rule-set'), 'platform', 999, 'draft');
		insert into scoring_rules (tenant, rule_set_id, priority, stackable, activity_id, score_source, rate)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/rule-set'), 1, false, 1, 'amount', 1);
		insert into scoring_rule_sets (tenant, id, scope, contest_id, version, status, mode)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/contest-rule-set'), 'contest',
			uuid_generate_v5(uuid_nil(), $1 || '/contest'), 1, 'draft', 'replace');
		insert into platform_scoring_config (tenant, singleton, active_rule_set_id)
		values ($1, true, uuid_generate_v5(uuid_nil(), $1 || '/rule-set'))
		on conflict (tenant, singleton) do nothing;
		insert into jobs (tenant, task_type, payload)
		values ($1, 'fixture.lifecycle.v1', '{}');
		insert into tenant_overrides (tenant, component)
		select $1, 'tadoku-worker' where $1 <> 'tadoku/prod'
		on conflict do nothing;
		insert into accounts (tenant, user_id, email, email_normalized, password_hash)
		values ($1, uuid_generate_v5(uuid_nil(), $1 || '/user'),
			'lifecycle@example.test', 'lifecycle@example.test', 'synthetic-hash');
		insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at)
		values ($1, 'synthetic-session-hash', uuid_generate_v5(uuid_nil(), $1 || '/user'),
			now(), now() + interval '1 day');
		insert into recovery_tokens (tenant, token_hash, user_id, expires_at)
		values ($1, 'synthetic-recovery-hash', uuid_generate_v5(uuid_nil(), $1 || '/user'),
			now() + interval '1 hour');
		insert into emails (tenant, template, recipient, subject, body, params)
		values ($1, 'account.recovery', 'lifecycle@example.test', 'Lifecycle', 'Fixture', '{}');
		insert into access_tokens (tenant, name, token_hash, scopes, expires_at)
		values ($1, 'Lifecycle', 'synthetic-access-hash', array['email:read'], now() + interval '1 day');
	`

	tx, err := db.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = tx.Rollback(ctx)
	})
	if _, err := tx.Exec(t.Context(), seed, pgx.QueryExecModeSimpleProtocol, key); err != nil {
		t.Fatalf("seed %s: %v", key, err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
}
