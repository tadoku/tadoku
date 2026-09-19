insert into pages (id, namespace, slug, current_content_id, published_at, created_at, updated_at)
values ('11111111-1111-4111-8111-111111111111', 'main', 'existing-page', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '2026-09-11 10:00:00', '2026-09-11 09:00:00', '2026-09-11 09:00:00');
insert into pages_content (id, page_id, title, html, created_at)
values ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '11111111-1111-4111-8111-111111111111', 'Existing', '<p>Existing</p>', '2026-09-11 09:00:00');
