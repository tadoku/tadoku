-- Add mutable tables here as their slices gain tests. Leave migration-seeded
-- reference tables and schema_migrations untouched; do not use cascade.
-- contests needs delete because PostgreSQL rejects truncate while the preserved
-- scoring_rule_sets table has a foreign key to it, even when no row references it.
truncate table public.announcements, public.contest_logs, public.contest_registrations,
  public.jobs, public.leaderboard_outbox, public.log_tags, public.logs,
  public.moderation_audit_log, public.pages, public.pages_content, public.posts,
  public.posts_content, public.user_roles, public.users restart identity;

-- Restore scoring seed data on every reset. Logs must be cleared first because
-- their provenance references scoring rule sets.
update public.contests set scoring_rule_set_id = null where scoring_rule_set_id is not null;
delete from public.platform_scoring_config;
delete from public.scoring_rules;
delete from public.scoring_rule_sets;
delete from public.contests;

insert into public.scoring_rule_sets
select * from public.tadoku_test_scoring_rule_sets_baseline;

insert into public.scoring_rules
select * from public.tadoku_test_scoring_rules_baseline;

insert into public.platform_scoring_config
select * from public.tadoku_test_platform_scoring_config_baseline;

-- Restore mutable language reference data while retaining every seeded
-- language needed by the migration-seeded scoring-rule foreign keys.
delete from public.languages
where code not in (select code from public.tadoku_test_language_baseline);

insert into public.languages (code, name)
select code, name from public.tadoku_test_language_baseline
on conflict (code) do update set name = excluded.name;
