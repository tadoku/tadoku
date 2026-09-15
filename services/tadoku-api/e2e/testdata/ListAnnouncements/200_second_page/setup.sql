insert into announcements (id, namespace, title, content, starts_at, ends_at, created_at, updated_at)
select
  ('20000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  'main',
  'Announcement ' || n,
  '<p>Announcement ' || n || '</p>',
  timestamp '2026-09-12 10:00:00',
  timestamp '2026-09-12 14:00:00',
  timestamp '2026-09-12 11:00:00' + n * interval '1 minute',
  timestamp '2026-09-12 11:00:00' + n * interval '1 minute'
from generate_series(1, 5) as n;
