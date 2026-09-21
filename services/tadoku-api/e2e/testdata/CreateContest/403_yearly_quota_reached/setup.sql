insert into contests (
  id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
  registration_end, title, activity_type_id_allow_list, official, created_at, updated_at
)
select
  ('aaaaaaaa-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  '11111111-1111-4111-8111-111111111111', 'Reader One', n % 2 = 0,
  '2025-01-01', '2025-01-31', '2025-01-01', 'Quota contest ' || n, '{}', n % 2 = 1,
  '2026-01-01'::timestamp + n * interval '1 day', '2026-01-01'
from generate_series(1, 12) as n;
