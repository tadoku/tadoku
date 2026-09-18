insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at, deleted_at)
values ('11111111-1111-4111-8111-111111111111', 'main', 'first-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2',
  '2026-09-11 12:00:00', '2026-09-10 12:00:00', '2026-09-11 12:00:00', '2026-09-12 10:00:00');

insert into posts_content (id, post_id, title, content, created_at)
values
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'Original title', 'Original content', '2026-09-10 12:00:00'),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', 'Revised title', 'Revised content', '2026-09-11 12:00:00');
