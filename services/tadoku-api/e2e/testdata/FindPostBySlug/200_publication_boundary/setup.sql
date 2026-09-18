insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
values
  ('11111111-1111-4111-8111-111111111111', 'main', 'welcome', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '2026-09-11 10:00:00', '2026-09-10 08:00:00', '2026-09-11 09:30:00', null),
  ('22222222-2222-4222-8222-222222222222', 'main', 'draft', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', null, '2026-09-10 08:00:00', '2026-09-11 09:30:00', null),
  ('33333333-3333-4333-8333-333333333333', 'main', 'scheduled', 'cccccccc-cccc-4ccc-8ccc-cccccccccccc', '2026-09-13 10:00:00', '2026-09-10 08:00:00', '2026-09-11 09:30:00', null),
  ('44444444-4444-4444-8444-444444444444', 'main', 'deleted', 'dddddddd-dddd-4ddd-8ddd-dddddddddddd', '2026-09-11 10:00:00', '2026-09-10 08:00:00', '2026-09-11 09:30:00', '2026-09-12 10:00:00'),
  ('55555555-5555-4555-8555-555555555555', 'main', 'missing-content', 'eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee', '2026-09-11 10:00:00', '2026-09-10 08:00:00', '2026-09-11 09:30:00', null);

insert into posts_content (id, post_id, title, content, created_at)
values
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa', '11111111-1111-4111-8111-111111111111', 'Welcome', '<p>Read this</p>', '2026-09-11 09:30:00'),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb', '22222222-2222-4222-8222-222222222222', 'Draft', 'Draft content', '2026-09-11 09:30:00'),
  ('cccccccc-cccc-4ccc-8ccc-cccccccccccc', '33333333-3333-4333-8333-333333333333', 'Scheduled', 'Scheduled content', '2026-09-11 09:30:00'),
  ('dddddddd-dddd-4ddd-8ddd-dddddddddddd', '44444444-4444-4444-8444-444444444444', 'Deleted', 'Deleted content', '2026-09-11 09:30:00'),
  ('ffffffff-ffff-4fff-8fff-ffffffffffff', '11111111-1111-4111-8111-111111111111', 'Previous title', 'Previous content', '2026-09-10 08:00:00');

update posts set published_at = '2026-09-12 12:00:00' where id = '11111111-1111-4111-8111-111111111111';
