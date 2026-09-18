insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at)
values ('22222222-2222-4222-8222-222222222222', 'other', 'first-post', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1',
  '2026-09-10 12:00:00', '2026-09-10 12:00:00', '2026-09-10 12:00:00');

insert into posts_content (id, post_id, title, content, created_at)
values ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '22222222-2222-4222-8222-222222222222', 'Existing post', 'Existing content', '2026-09-10 12:00:00');
