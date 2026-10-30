-- At the fixture instant, 2026-09-12 12:00 UTC, the first contest ends that day and is unfinished; the second
-- ended the day before and is finished. The reader has a registration and one log in each.
insert into users (tenant, id, display_name, created_at, updated_at) values
  ('tadoku/prod', '99999999-9999-4999-8999-999999999999', 'Other Owner', '2026-01-01', '2026-01-01');

insert into contests (
  tenant, id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
  registration_end, title, language_code_allow_list, activity_type_id_allow_list, official,
  created_at, updated_at
) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000c1', '99999999-9999-4999-8999-999999999999', 'Other Owner', false,
   '2026-09-01', '2026-09-12', '2026-09-12', 'Ends today', '{eng}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000c2', '99999999-9999-4999-8999-999999999999', 'Other Owner', false,
   '2026-09-01', '2026-09-11', '2026-09-11', 'Ended yesterday', '{eng}', '{1}', false,
   '2026-01-01', '2026-01-01');

insert into contest_registrations (tenant, id, contest_id, user_id, language_codes) values
  ('tadoku/prod', 'e0000000-0000-4000-8000-0000000000d1', 'f0000000-0000-4000-8000-0000000000c1',
   '11111111-1111-4111-8111-111111111111', '{eng}'),
  ('tadoku/prod', 'e0000000-0000-4000-8000-0000000000d2', 'f0000000-0000-4000-8000-0000000000c2',
   '11111111-1111-4111-8111-111111111111', '{eng}');

insert into logs (
  tenant, id, user_id, language_code, log_activity_id, duration_seconds, computed_score,
  eligible_official_leaderboard, description, created_at, updated_at
) values
  ('tadoku/prod', 'd0000000-0000-4000-8000-0000000000e1', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   300, 5, false, 'Ends today', '2026-09-05', '2026-09-05'),
  ('tadoku/prod', 'd0000000-0000-4000-8000-0000000000e2', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   600, 10, false, 'Ended yesterday', '2026-09-05', '2026-09-05');

insert into contest_logs (tenant, contest_id, log_id, duration_seconds, computed_score) values
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000c1', 'd0000000-0000-4000-8000-0000000000e1', 300, 5),
  ('tadoku/prod', 'f0000000-0000-4000-8000-0000000000c2', 'd0000000-0000-4000-8000-0000000000e2', 600, 10);
