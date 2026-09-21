-- Add mutable tables here as their slices gain tests. Leave migration-seeded
-- reference tables and schema_migrations untouched; do not use cascade.
-- contests needs delete because PostgreSQL rejects truncate while the preserved
-- scoring_rule_sets table has a foreign key to it, even when no row references it.
delete from public.contests;
truncate table public.announcements, public.moderation_audit_log, public.pages, public.pages_content, public.posts, public.posts_content, public.user_roles, public.users restart identity;

-- Restore mutable language reference data while retaining every seeded
-- language needed by the migration-seeded scoring-rule foreign keys.
delete from public.languages
where code not in (select code from public.tadoku_test_language_baseline);

insert into public.languages (code, name)
select code, name from public.tadoku_test_language_baseline
on conflict (code) do update set name = excluded.name;
