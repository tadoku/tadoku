insert into users (id, display_name, created_at, updated_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-09-13'),
  ('44444444-4444-4444-8444-444444444444', 'Reader Two', '2026-01-01', '2026-01-01'),
  ('99999999-9999-4999-8999-999999999999', 'Owner', '2026-01-01', '2026-01-01');
insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at) values
  ('f1111111-1111-4111-8111-111111111111', '99999999-9999-4999-8999-999999999999', 'stale', false, '2025-12-01', '2026-09-30', '2026-08-31', 'Detach fixture', '{jpn,eng}', '{1}', true, '2026-01-01', '2026-01-01');
insert into contest_registrations (id, contest_id, user_id, language_codes, created_at, updated_at) values
  ('e1111111-1111-4111-8111-111111111111', 'f1111111-1111-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111', '{jpn,eng}', '2026-01-01', '2026-01-01');
insert into logs (id, user_id, language_code, log_activity_id, duration_seconds, computed_score, eligible_official_leaderboard, created_at, updated_at) values
  ('d1111111-1111-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 60, 1, true, '2026-09-01', '2026-09-01'),
  ('d2222222-2222-4222-8222-222222222222', '11111111-1111-4111-8111-111111111111', 'eng', 1, 60, 1, true, '2026-09-01', '2026-09-01'),
  ('d3333333-3333-4333-8333-333333333333', '44444444-4444-4444-8444-444444444444', 'eng', 1, 60, 1, true, '2026-09-01', '2026-09-01');
insert into contest_logs (contest_id, log_id, duration_seconds, computed_score) values
  ('f1111111-1111-4111-8111-111111111111', 'd1111111-1111-4111-8111-111111111111', 60, 1),
  ('f1111111-1111-4111-8111-111111111111', 'd2222222-2222-4222-8222-222222222222', 60, 1),
  ('f1111111-1111-4111-8111-111111111111', 'd3333333-3333-4333-8333-333333333333', 60, 1);
