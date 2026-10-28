package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

func seed(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("seed", flag.ContinueOnError)
	database := flags.String("name", "qc_schema", "empty migrated database")
	scale := flags.Int("scale", 1, "1 or 10")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *scale != 1 && *scale != 10 {
		return fmt.Errorf("scale must be 1 or 10")
	}
	owner, err := connect(ctx, localDSN, *database, "tadoku_owner")
	if err != nil {
		return err
	}
	defer owner.Close(ctx)

	statements := []string{
		`insert into tenants (key, kind) values ('climb/other', 'test')`,
		`insert into languages (code, name) values
          ('ace', 'Achinese'),
          ('ady', 'Adyghe'),
          ('ain', 'Ainu'),
          ('akk', 'Akkadian'),
          ('amh', 'Amharic'),
          ('ast', 'Asturian')`,

		`insert into users (tenant, id, display_name, created_at, updated_at, deleted_at)
   select tenant, md5('user-' || i)::uuid,
    (case
      when tenant = 'tadoku/prod' then 'Learner '
      else 'Other learner '
    end) || i,
    '2020-01-01'::timestamp, '2026-01-01'::timestamp,
    case
      when i % 97 = 0 then '2026-02-01'::timestamp
    end
   from generate_series(1, 5000 * SCALE) i
   cross join (values
    ('tadoku/prod'),
    ('climb/other')
   ) t(tenant)
   where tenant = 'tadoku/prod' or i = 1`,
		`insert into contests (tenant, id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
    registration_end, title, language_code_allow_list, activity_type_id_allow_list,
    official, created_at, updated_at, deleted_at)
   select 'tadoku/prod', md5('contest-' || i)::uuid, md5('user-1')::uuid, 'Learner 1', i % 7 = 0,
    '2026-01-01'::date, '2100-12-31'::date, '2100-12-01'::date, 'Contest ' || i,
    case
      when i % 3 = 0 then array['jpn', 'eng']::varchar[]
    end, array[1,2], i % 2 = 0,
    '2026-01-01'::timestamp + i * interval '1 minute', '2026-01-01'::timestamp,
    case
      when i % 23 = 0 then '2026-02-01'::timestamp
    end
   from generate_series(1, 143 * SCALE) i`,
		`insert into contests (tenant, id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
    registration_end, title, activity_type_id_allow_list, official, created_at, updated_at)
   values ('climb/other', md5('other-contest-1')::uuid, md5('user-1')::uuid, 'Learner 1', false,
    '2026-01-01', '2100-12-31', '2100-12-01', 'Other tenant', array[1,2], true,
    '2026-01-01', '2026-01-01')`,
		`insert into contest_registrations (tenant, id, contest_id, user_id, language_codes,
    created_at, updated_at, deleted_at)
   select 'tadoku/prod', md5('registration-' || c || '-' || u)::uuid,
    md5('contest-' || c)::uuid, md5('user-' || u)::uuid,
    case
      when u % 3 = 0 then array['jpn']::varchar[]
      else array['jpn', 'eng', 'deu']::varchar[]
    end,
    '2026-01-01'::timestamp + (u / 2) * interval '1 minute', '2026-01-01'::timestamp,
    case
      when u % 101 = 0 then '2026-02-01'::timestamp
    end
   from generate_series(1, 143 * SCALE) c
   cross join lateral generate_series(1, case
     when c <= 4 then 60 * SCALE
     else 60
   end) u`,
		`insert into logs (tenant, id, user_id, language_code, log_activity_id, unit_id, unit_key,
    description, amount, modifier, duration_seconds, computed_score,
    eligible_official_leaderboard, created_at, updated_at, deleted_at)
   select 'tadoku/prod', md5('log-' || i)::uuid,
    md5('user-' || case
     when i <= 1700 * SCALE then 1
     when i <= 2050 * SCALE then 2
     when i <= 7000 * SCALE then 2 + i % 59
     when i <= 7030 * SCALE then 950
     else 2 + i % (1999 * SCALE)
    end)::uuid,
    (array['jpn', 'jpn', 'eng', 'deu'])[1 + i % 4], 1 + i % 2,
    case
     when i % 20 = 0 then null
     else
     (select id from log_units
      where unit_key = case
       when i % 2 = 0 then 'reading_page'
       else 'listening_minute'
      end
      order by id
      limit 1)
    end,
    case
      when i % 2 = 0 then 'reading_page'
      else 'listening_minute'
    end,
    'Synthetic activity ' || i || repeat(' reading', 20), case
      when i % 71 = 0 then 0
      else (i % 5) * 10
    end, 1,
    case
      when i % 3 = 0 or i % 20 = 0 then 120
    end,
    case
      when i % 5 = 0 then (i % 5) * 10
    end,
    i % 11 <> 0,
    (case
     when i <= 18000 * SCALE then '2026-01-01'::timestamp
     else '2022-01-01'::timestamp + (i % 4) * interval '1 year'
    end) + (i / 2) * interval '1 minute', '2026-10-01'::timestamp,
    case
      when i % 83 = 0 then '2026-10-01'::timestamp
    end
   from generate_series(1, 40000 * SCALE) i`,
		`insert into logs (tenant, id, user_id, language_code, log_activity_id, unit_key,
    description, amount, modifier, duration_seconds, eligible_official_leaderboard, created_at, updated_at)
   values ('climb/other', md5('other-log-1')::uuid, md5('user-1')::uuid, 'jpn', 1,
    'reading_page', 'Other tenant activity', 999, 1, 120, true, '2026-01-01', '2026-01-01')`,
		`insert into contest_logs (tenant, contest_id, log_id, unit_key, amount, modifier,
    duration_seconds, computed_score)
   select 'tadoku/prod', md5('contest-' ||
    case
      when i <= 7000 * SCALE then 1
      when i <= 12000 * SCALE then 2
      when i <= 15300 * SCALE then 3
      else 4
    end)::uuid,
    md5('log-' || i)::uuid, 'reading_page', case
      when i % 71 = 0 then 0
      else i % 5 * 10
    end,
    1, case
      when i % 3 = 0 or i % 20 = 0 then 120
    end, case
      when i % 5 = 0 then 0
    end
   from generate_series(1, 15360 * SCALE) i`,
		`insert into log_tags (tenant, log_id, user_id, tag)
   select l.tenant, l.id, l.user_id, tag
   from logs l
   cross join (values
    ('book'),
    ('fiction'),
    ('ebook'),
    ('news'),
    ('literature'),
    ('science'),
    ('history'),
    ('essay')
   ) t(tag)
   where l.description !~ '^Synthetic activity [0-9]*0 '`,
		`insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at, updated_at)
   select tenant, id, id || '@example.invalid', id || '@example.invalid', 'synthetic-hash',
    '2026-01-01', '2026-01-01' from users`,
		`insert into sessions (tenant, token_hash, user_id, authenticated_at, expires_at, created_at)
   select tenant, 'session-' || id, id, '2026-01-01', '2100-01-01', '2026-01-01' from users
   where id in (select md5('user-' || i)::uuid from generate_series(1, 17) i)`,
		`insert into jobs (tenant, task_type, payload, state, created_at, next_attempt_at, completed_at)
   select 'tadoku/prod', 'leaderboard.invalidate.v1', '{}'::jsonb, 'pending',
    '2026-01-01', '2100-01-01', null from generate_series(1, 2414 * SCALE) i`,
		`insert into announcements (tenant, id, namespace, title, content, starts_at, ends_at, created_at, updated_at)
         select 'tadoku/prod', md5('announcement-' || i)::uuid, 'tadoku', 'Notice ' || i,
          'Synthetic notice', '2026-01-01', '2100-01-01',
          '2026-01-01'::timestamp + i * interval '1 minute', '2026-01-01'
         from generate_series(1, 50 * SCALE) i`,
		`insert into pages (tenant, id, namespace, slug, current_content_id, published_at, created_at, updated_at)
         select 'tadoku/prod', md5('page-' || i)::uuid, 'tadoku', 'page-' || i,
          md5('page-content-' || i)::uuid, '2026-01-01',
          '2026-01-01'::timestamp + i * interval '1 minute', '2026-01-01'
         from generate_series(1, 50 * SCALE) i`,
		`insert into pages_content (tenant, id, page_id, title, html, created_at)
         select tenant, current_content_id, id, slug, '<p>Synthetic page</p>', created_at from pages`,
		`insert into posts (tenant, id, namespace, slug, current_content_id, published_at, created_at, updated_at)
         select 'tadoku/prod', md5('post-' || i)::uuid, 'tadoku', 'post-' || i,
          md5('post-content-' || i)::uuid, '2026-01-01',
          '2026-01-01'::timestamp + i * interval '1 minute', '2026-01-01'
         from generate_series(1, 50 * SCALE) i`,
		`insert into posts_content (tenant, id, post_id, title, content, created_at)
         select tenant, current_content_id, id, slug, 'Synthetic post', created_at from posts`,
		`insert into recovery_tokens (tenant, token_hash, user_id, expires_at)
         select tenant, 'recovery-' || id, id, '2100-01-01' from users`,
		`insert into access_tokens (tenant, id, name, token_hash, scopes, created_at, expires_at)
         values ('tadoku/prod', md5('token-1')::uuid, 'Synthetic token', 'access-token',
          array['email:read'], '2026-01-01', '2100-01-01')`,
		`insert into emails (tenant, id, template, recipient, subject, body, params, created_at)
         select 'tadoku/prod', md5('email-' || i)::uuid, 'recovery', 'learner@example.invalid',
          'Synthetic subject', 'Synthetic body', '{}'::jsonb,
          '2026-01-01'::timestamp + i * interval '1 minute'
         from generate_series(1, 50 * SCALE) i`,
		`insert into permission_relationships
          (tenant, object_type, object_id, relation, subject_type, subject_id, created_at)
         select 'tadoku/prod', 'contest', md5('contest-1')::uuid::text, 'owner', 'user',
          md5('user-' || i)::uuid::text, '2026-01-01'
         from generate_series(1, 10 * SCALE) i`,
		`insert into contest_events (tenant, id, contest_id, log_id, event, occurred_at)
         select 'tadoku/prod', md5('event-' || i)::uuid, md5('contest-1')::uuid, md5('log-' || i)::uuid,
          'log.submitted', '2026-01-01'::timestamp + i * interval '1 minute'
         from generate_series(1, 100 * SCALE) i`,
		`insert into jobs (tenant, task_type, payload, state, attempts, claim_token, lease_expires_at,
          created_at, next_attempt_at, failed_at, last_error)
         values
          ('tadoku/prod', 'climb.work.v1', '{}'::jsonb, 'pending', 0, null, null,
           '2026-01-01', '2026-01-01', null, null),
          ('tadoku/prod', 'climb.work.v1', '{}'::jsonb, 'running', 1, md5('claim-1')::uuid, '2100-01-01',
           '2026-01-01', '2026-01-01', null, null),
          ('tadoku/prod', 'climb.work.v1', '{}'::jsonb, 'pending', 10, null, null,
           '2026-01-01', '2026-01-01', null, null),
          ('tadoku/prod', 'climb.work.v1', '{}'::jsonb, 'failed', 10, null, null,
           '2026-01-01', '2026-01-01', '2026-01-02', 'attempts_exhausted'),
          ('tadoku/prod', 'email.send.v1', jsonb_build_object('email_id', md5('email-1')::uuid),
           'failed', 10, null, null, '2026-01-01', '2026-01-01', '2026-01-02', 'attempts_exhausted')`,
		`insert into scoring_rule_sets (tenant, id, scope, contest_id, version, status, mode, created_at)
         values ('tadoku/prod', md5('draft-scoring')::uuid, 'contest', md5('contest-1')::uuid,
          1, 'draft', 'replace', '2026-01-01')`,
		`insert into users (tenant, id, display_name, created_at, updated_at)
         values ('tadoku/prod', md5('user-unregistered')::uuid, 'Unregistered learner',
          '2026-01-01', '2026-01-01')`,
		`insert into tenant_overrides (tenant, component) values ('climb/other', 'climb-existing')`,
		`update sessions set expires_at = '2026-02-01'
         where tenant = 'tadoku/prod' and user_id = md5('user-2')::uuid`,
		`update recovery_tokens set expires_at = '2026-02-01'
         where tenant = 'tadoku/prod' and user_id = md5('user-2')::uuid`,
		`insert into contest_registrations (tenant, id, contest_id, user_id, language_codes, created_at, updated_at)
         values ('tadoku/prod', md5('registration-empty')::uuid, md5('contest-1')::uuid,
          md5('user-unregistered')::uuid, array['jpn', 'eng'], '2026-01-01', '2026-01-01')`,
		`update logs set amount = 0, computed_score = 0 where user_id = md5('user-3')::uuid`,
		`update contest_logs set amount = 0, computed_score = 0 where log_id in
          (select id from logs where user_id = md5('user-3')::uuid)`,
		`insert into jobs (tenant, task_type, payload, state, created_at, next_attempt_at, completed_at)
         values ('tadoku/prod', 'climb.cleanup.v1', '{}'::jsonb, 'completed',
          '2020-01-01', '2020-01-01', '2020-01-02')`,
		`insert into tenant_overrides (tenant, component)
         select 'climb/other', 'climb-fixture-' || i from generate_series(1, 200) i`,
		`insert into logs (tenant, id, user_id, language_code, log_activity_id, unit_key,
          description, amount, modifier, duration_seconds, eligible_official_leaderboard, created_at, updated_at)
         select 'tadoku/prod', md5('fractional-' || i)::uuid, md5('user-' || 5000 * SCALE)::uuid,
          'jpn', 5, 'study_minute', 'Fractional edge', i::real / 10, 1.6, 120, true,
          '2026-01-02'::timestamp + i * interval '1 second', '2026-01-02'::timestamp
         from generate_series(1, 5) i`,
		`insert into logs (tenant, id, user_id, language_code, log_activity_id, unit_key,
          description, amount, modifier, duration_seconds, eligible_official_leaderboard, created_at, updated_at)
         values
          ('tadoku/prod', md5('missing-user-log')::uuid, md5('missing-user')::uuid,
           'jpn', 5, 'study_minute', 'Missing user edge', 1, 1, 120, false, '2026-01-02', '2026-01-02'),
          ('tadoku/prod', md5('missing-language-log')::uuid, md5('user-1')::uuid,
           'qzz', 5, 'study_minute', 'Missing language edge', 1, 1, 120, false, '2026-01-02', '2026-01-02')`,
		`insert into contest_logs (tenant, contest_id, log_id, unit_key, amount, modifier, duration_seconds)
         select 'tadoku/prod', md5(case
          when description = 'Fractional edge' then 'contest-4'
          else 'contest-1'
         end)::uuid,
          id, unit_key, amount, modifier, duration_seconds
         from logs where description in ('Fractional edge', 'Missing user edge', 'Missing language edge')`,
		`insert into contest_registrations (tenant, id, contest_id, user_id, language_codes, created_at, updated_at)
         values ('tadoku/prod', md5('fractional-registration')::uuid, md5('contest-4')::uuid,
          md5('user-' || 5000 * SCALE)::uuid, array['jpn'], '2026-01-01', '2026-01-01')`,
		`insert into log_tags (tenant, log_id, user_id, tag)
         select 'tadoku/prod', id, user_id, tag from logs
         cross join (values
          ('a'),
          ('á'),
          ('z'),
          ('あ')
         ) t(tag)
         where id = md5('fractional-1')::uuid`,
	}
	tx, err := owner.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "select set_config('tadoku.tenant', $1, true)", "tadoku/prod"); err != nil {
		return err
	}
	for i, statement := range statements {
		sql := strings.ReplaceAll(statement, "SCALE", fmt.Sprint(*scale))
		if _, err := tx.Exec(ctx, sql); err != nil {
			return fmt.Errorf("seed statement %d: %w", i+1, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	rows, err := owner.Query(ctx, `select c.relname, a.attname
  from pg_class c
  join pg_namespace n on n.oid = c.relnamespace
  join pg_attribute a on a.attrelid = c.oid
  where n.nspname = 'public' and c.relkind = 'r' and a.attnum > 0 and not a.attisdropped
   and a.attname in ('tenant', 'user_id', 'contest_id', 'log_id', 'language_code',
    'log_activity_id', 'deleted_at', 'year', 'created_at', 'state', 'task_type')
  order by c.relname, a.attname`)
	if err != nil {
		return err
	}
	var alters []string
	for rows.Next() {
		var table, column string
		if err := rows.Scan(&table, &column); err != nil {
			rows.Close()
			return err
		}
		alters = append(alters, "alter table "+pgx.Identifier{table}.Sanitize()+
			" alter column "+pgx.Identifier{column}.Sanitize()+" set statistics 1000")
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, sql := range alters {
		if _, err := owner.Exec(ctx, sql); err != nil {
			return err
		}
	}
	if _, err := owner.Exec(ctx, "vacuum (analyze)"); err != nil {
		return err
	}
	var bytes int64
	if err := owner.QueryRow(ctx, "select pg_database_size(current_database())").Scan(&bytes); err != nil {
		return err
	}
	return output(map[string]any{"database": *database, "scale": *scale, "seed": 1, "bytes": bytes})
}
