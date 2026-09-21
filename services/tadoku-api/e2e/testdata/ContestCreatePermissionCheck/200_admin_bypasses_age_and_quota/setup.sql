insert into contests (
  id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
  registration_end, title, activity_type_id_allow_list, official, created_at, updated_at
)
select
  ('bbbbbbbb-0000-4000-8000-' || lpad(n::text, 12, '0'))::uuid,
  '22222222-2222-4222-8222-222222222222', 'Admin One', false,
  '2026-01-01', '2026-01-31', '2026-01-01', 'Admin quota contest ' || n, '{}', false,
  '2026-01-01'::timestamp + n * interval '1 day', '2026-01-01'
from generate_series(1, 12) as n;
