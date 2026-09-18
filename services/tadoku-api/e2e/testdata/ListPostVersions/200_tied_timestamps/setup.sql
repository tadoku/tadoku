insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
values
  ('11111111-1111-4111-8111-111111111111', 'main', 'scheduled-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '2026-09-14 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00', null),
  ('22222222-2222-4222-8222-222222222222', 'main', 'another-post', 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', null, '2026-09-10 12:00:00', '2026-09-11 12:00:00', null);

insert into posts_content (id, post_id, title, content, created_at)
values
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', 'Revised title', 'Revised content', '2026-09-11 12:00:00'),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'Original title', 'Original content', '2026-09-10 12:00:00'),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '22222222-2222-4222-8222-222222222222', 'Another post', 'Another content', '2026-09-10 12:00:00');

update posts_content set created_at = '2026-09-10 12:00:00' where post_id = '11111111-1111-4111-8111-111111111111';
