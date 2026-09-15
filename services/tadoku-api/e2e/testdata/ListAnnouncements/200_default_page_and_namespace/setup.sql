insert into announcements (id, namespace, title, content, starts_at, ends_at, created_at, updated_at)
select
  ('10000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  'main',
  'Announcement ' || lpad(n::text, 2, '0'),
  '<p>Announcement ' || lpad(n::text, 2, '0') || '</p>',
  timestamp '2026-09-12 10:00:00',
  timestamp '2026-09-12 14:00:00',
  timestamp '2026-09-12 11:00:00' + n * interval '1 minute',
  timestamp '2026-09-12 11:00:00' + n * interval '1 minute'
from generate_series(1, 11) as n;

insert into announcements (id, namespace, title, content, starts_at, ends_at, created_at, updated_at, deleted_at)
values
  ('10000000-0000-4000-8000-000000000012', 'other', 'other namespace', '<p>other</p>',
   '2026-09-12 10:00:00', '2026-09-12 14:00:00', '2026-09-12 12:00:00', '2026-09-12 12:00:00', null),
  ('10000000-0000-4000-8000-000000000013', 'main', 'deleted', '<p>deleted</p>',
   '2026-09-12 10:00:00', '2026-09-12 14:00:00', '2026-09-12 12:01:00', '2026-09-12 12:01:00', '2026-09-12 12:02:00');
