insert into pages (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
values
  ('10000000-0000-4000-8000-000000000004', 'main', 'boundary', '20000000-0000-4000-8000-000000000004', '2026-09-12T12:00:00Z', '2026-09-12T10:00:00Z', '2026-09-12T11:00:00Z', null),
  ('10000000-0000-4000-8000-000000000005', 'main', 'after-boundary', '20000000-0000-4000-8000-000000000005', '2026-09-12T12:00:00.000001Z', '2026-09-12T10:30:00Z', '2026-09-12T11:00:00Z', null);

insert into pages_content (id, page_id, title, html, created_at)
values
  ('20000000-0000-4000-8000-000000000004', '10000000-0000-4000-8000-000000000004', 'Boundary', '<p>Boundary</p>', '2026-09-12T11:00:00Z'),
  ('20000000-0000-4000-8000-000000000005', '10000000-0000-4000-8000-000000000005', 'After boundary', '<p>After boundary</p>', '2026-09-12T11:00:00Z');
