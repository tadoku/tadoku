-- Add mutable tables here as their slices gain tests. Leave migration-seeded
-- reference tables and schema_migrations untouched; do not use cascade.
truncate table public.announcements, public.posts, public.posts_content restart identity;
