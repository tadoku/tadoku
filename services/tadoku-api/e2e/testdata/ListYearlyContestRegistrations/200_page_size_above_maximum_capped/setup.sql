insert into users (id, display_name, created_at, updated_at, deleted_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-01-01', null),
  ('99999999-9999-4999-8999-999999999999', 'Raw Owner Name', '2026-01-01', '2026-01-01', null);

insert into contests (
  id,
  owner_user_id,
  owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  activity_type_id_allow_list,
  official,
  created_at,
  updated_at
)
select
  ('f0000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  '99999999-9999-4999-8999-999999999999',
  'stale',
  false,
  '2026-01-01',
  '2026-01-31',
  '2026-01-31',
  'Contest ' || n,
  '{1}',
  false,
  '2026-01-01',
  '2026-01-01'
from generate_series(1, 51) as n;

insert into contest_registrations (id, contest_id, user_id, language_codes, created_at, updated_at)
select
  ('eeeeeeee-eeee-4eee-8eee-' || lpad(n::text, 12, '0'))::uuid,
  ('f0000000-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  '11111111-1111-4111-8111-111111111111',
  '{jpn}',
  timestamp '2026-01-01 00:00:00' + n * interval '1 minute',
  timestamp '2026-01-01 00:00:00' + n * interval '1 minute'
from generate_series(1, 51) as n;
