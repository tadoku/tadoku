insert into posts (id, namespace, slug, current_content_id, published_at, created_at, updated_at)
select
  ('10000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  'main',
  'post-' || n,
  ('20000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  timestamp '2026-09-12 11:00:00',
  timestamp '2026-09-12 09:00:00' + n * interval '1 minute',
  timestamp '2026-09-12 09:00:00' + n * interval '1 minute'
from generate_series(1, 11) as n;

insert into posts_content (id, post_id, title, content, created_at)
select
  ('20000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  ('10000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  'Post ' || n,
  'Content ' || n,
  timestamp '2026-09-12 09:00:00' + n * interval '1 minute'
from generate_series(1, 11) as n;
