-- Add mutable tables here as their slices gain tests. Leave migration-seeded
-- reference tables and schema_migrations untouched; do not use cascade.
truncate table public.announcements, public.pages, public.pages_content, public.posts, public.posts_content restart identity;
