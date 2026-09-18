insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
values
  ('11111111-1111-4111-8111-111111111111', 'main', 'published-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '2026-09-11 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00', null),
  ('22222222-2222-4222-8222-222222222222', 'main', 'draft-post', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', null, '2026-09-10 13:00:00', '2026-09-10 13:00:00', null),
  ('33333333-3333-4333-8333-333333333333', 'main', 'deleted-post', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '2026-09-10 13:00:00', '2026-09-10 13:00:00', '2026-09-10 13:00:00', '2026-09-11 12:00:00'),
  ('44444444-4444-4444-8444-444444444444', 'main', 'scheduled-post', 'dddddddd-dddd-4ddd-8ddd-ddddddddddd1', '2026-09-13 12:00:00', '2026-09-10 13:00:00', '2026-09-10 13:00:00', null);
insert into posts_content (id, post_id, title, content, created_at)
values
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '11111111-1111-4111-8111-111111111111', 'Latest title', '<p>Latest content</p>', '2026-09-11 12:00:00'),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', 'Tied title', '<p>Tied content</p>', '2026-09-10 12:00:00'),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'Original title', '<p>Original content</p>', '2026-09-10 12:00:00'),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '22222222-2222-4222-8222-222222222222', 'Draft title', '', '2026-09-10 13:00:00'),
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '33333333-3333-4333-8333-333333333333', 'Deleted title', '<p>Deleted content</p>', '2026-09-10 13:00:00'),
  ('dddddddd-dddd-4ddd-8ddd-ddddddddddd1', '44444444-4444-4444-8444-444444444444', 'Scheduled title', '<p>Scheduled content</p>', '2026-09-10 13:00:00');
