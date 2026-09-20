-- Add mutable tables here as their slices gain tests. Leave migration-seeded
-- reference tables and schema_migrations untouched; do not use cascade.
truncate table public.account_deletion_requests, public.announcements, public.pages, public.pages_content, public.posts, public.posts_content restart identity;

-- Restore mutable language reference data while retaining every seeded
-- language needed by the migration-seeded scoring-rule foreign keys.
delete from public.languages
where code not in (select code from public.tadoku_test_language_baseline);

insert into public.languages (code, name)
select code, name from public.tadoku_test_language_baseline
on conflict (code) do update set name = excluded.name;
