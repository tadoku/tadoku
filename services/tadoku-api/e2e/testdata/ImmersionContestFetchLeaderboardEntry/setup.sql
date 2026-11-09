insert into users (id, display_name, created_at, updated_at, deleted_at)
values ('11111111-1111-4111-8111-111111111111', 'Cedar', '2026-01-01', '2026-01-01', null);

insert into users (id, display_name, created_at, updated_at, deleted_at)
select
  ('80000000-0000-4000-8000-' || lpad(i::text, 12, '0'))::uuid,
  'Reader ' || i,
  '2026-01-01',
  '2026-01-01',
  null
from generate_series(1, 62) as i;

insert into contests (
  id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
  registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at
)
values (
  'f4444444-4444-4444-8444-444444444444', '11111111-1111-4111-8111-111111111111', 'Cedar', false, '2026-01-01', '2026-12-31',
  '2026-01-31', 'Sixty readers', '{jpn}', '{1}', true, '2026-01-01', '2026-01-01'
);

-- Readers 1 to 60 score 61 - i, except reader 59, who ties reader 58. Reader 61 has no logs and reader 62 withdrew.
insert into contest_registrations (id, contest_id, user_id, language_codes, created_at, updated_at, deleted_at)
select
  ('e0000000-0000-4000-8000-' || lpad(i::text, 12, '0'))::uuid,
  'f4444444-4444-4444-8444-444444444444',
  ('80000000-0000-4000-8000-' || lpad(i::text, 12, '0'))::uuid,
  '{jpn}',
  '2026-01-01',
  '2026-01-01',
  case when i = 62 then timestamp '2026-02-01' end
from generate_series(1, 62) as i;

insert into logs (
  id, user_id, language_code, log_activity_id, amount, modifier, computed_score, duration_seconds,
  eligible_official_leaderboard, created_at, updated_at, deleted_at
)
select
  ('a0000000-0000-4000-8000-' || lpad(i::text, 12, '0'))::uuid,
  ('80000000-0000-4000-8000-' || lpad(i::text, 12, '0'))::uuid,
  'jpn',
  1,
  scores.value,
  1,
  scores.value,
  60,
  true,
  '2026-02-01',
  '2026-02-01',
  null
from generate_series(1, 60) as i
cross join lateral (select case when i = 59 then 3 else 61 - i end as value) as scores;

insert into contest_logs (contest_id, log_id, amount, modifier, duration_seconds, computed_score)
select
  'f4444444-4444-4444-8444-444444444444',
  ('a0000000-0000-4000-8000-' || lpad(i::text, 12, '0'))::uuid,
  scores.value,
  1,
  60,
  scores.value
from generate_series(1, 60) as i
cross join lateral (select case when i = 59 then 3 else 61 - i end as value) as scores;
