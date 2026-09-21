-- Only the explicitly selected disposable application database is seeded.
-- Kratos, Keto, Valkey and all shared application databases are untouched.
begin;

insert into pages (id, namespace, slug, current_content_id, published_at)
values ('a70c0000-0000-4000-8000-000000000001', 'main', 'dev-cli-pilot',
        'a70c0000-0000-4000-8000-000000000002', '2026-01-01 00:00:00')
on conflict (id) do update set current_content_id = excluded.current_content_id;

insert into pages_content (id, page_id, title, html)
values ('a70c0000-0000-4000-8000-000000000002', 'a70c0000-0000-4000-8000-000000000001',
        'Dev CLI isolated database', '<p>This page exists only in the explicitly seeded pilot database.</p>')
on conflict (id) do update set title = excluded.title, html = excluded.html;

commit;
