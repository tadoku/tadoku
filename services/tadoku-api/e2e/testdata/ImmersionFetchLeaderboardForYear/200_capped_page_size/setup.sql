insert into users (id, display_name, created_at, updated_at)
select
  ('80000000-0000-4000-8000-' || lpad(series::text, 12, '0'))::uuid,
  'Cap ' || lpad(series::text, 3, '0'),
  '2026-01-01',
  '2026-01-01'
from generate_series(1, 101) as series;
