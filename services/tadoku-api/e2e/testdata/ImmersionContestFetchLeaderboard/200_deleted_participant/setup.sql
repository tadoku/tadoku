insert into users (id, display_name, created_at, updated_at, deleted_at) values
  ('11111111-1111-4111-8111-111111111111', 'Cedar', '2026-01-01', '2026-01-01', null),
  ('22222222-2222-4222-8222-222222222222', 'Baker', '2026-01-01', '2026-01-01', null),
  ('77777777-7777-4777-8777-777777777777', 'Deleted participant', '2026-01-01', '2026-08-01', '2026-08-01'),
  ('88888888-8888-4888-8888-888888888888', 'Deleted participant', '2026-01-01', '2026-08-01', '2026-08-01');

insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at) values
  ('f3333333-3333-4333-8333-333333333333', '11111111-1111-4111-8111-111111111111', 'Cedar', false, '2026-01-01', '2026-03-31', '2026-01-31', 'Finished leaderboard fixture', '{jpn,eng}', '{1,2}', false, '2026-01-01', '2026-01-01');

insert into contest_registrations (id, contest_id, user_id, language_codes, created_at, updated_at, deleted_at) values
  ('e1111111-1111-4111-8111-111111111111', 'f3333333-3333-4333-8333-333333333333', '11111111-1111-4111-8111-111111111111', '{jpn}', '2026-01-01', '2026-01-01', null),
  ('e2222222-2222-4222-8222-222222222222', 'f3333333-3333-4333-8333-333333333333', '22222222-2222-4222-8222-222222222222', '{eng}', '2026-01-02', '2026-01-02', null),
  ('e7777777-7777-4777-8777-777777777777', 'f3333333-3333-4333-8333-333333333333', '77777777-7777-4777-8777-777777777777', '{jpn}', '2026-01-03', '2026-01-03', null),
  ('e8888888-8888-4888-8888-888888888888', 'f3333333-3333-4333-8333-333333333333', '88888888-8888-4888-8888-888888888888', '{jpn}', '2026-01-04', '2026-01-04', null);

insert into logs (id, user_id, language_code, log_activity_id, amount, modifier, computed_score, duration_seconds, eligible_official_leaderboard, created_at, updated_at, deleted_at) values
  ('a1111111-1111-4111-8111-000000000001', '11111111-1111-4111-8111-111111111111', 'jpn', 1, 10, 1, null, 60, false, '2026-02-01', '2026-02-01', null),
  ('a2222222-2222-4222-8222-000000000001', '22222222-2222-4222-8222-222222222222', 'eng', 2, 25, 1, null, 60, false, '2026-02-02', '2026-02-02', null),
  ('a7777777-7777-4777-8777-000000000001', '77777777-7777-4777-8777-777777777777', 'jpn', 1, 40, 1, null, 60, false, '2026-02-03', '2026-02-03', null);

insert into contest_logs (contest_id, log_id, amount, modifier, duration_seconds, computed_score) values
  ('f3333333-3333-4333-8333-333333333333', 'a1111111-1111-4111-8111-000000000001', 10, 1, 60, null),
  ('f3333333-3333-4333-8333-333333333333', 'a2222222-2222-4222-8222-000000000001', 25, 1, 60, null),
  ('f3333333-3333-4333-8333-333333333333', 'a7777777-7777-4777-8777-000000000001', 40, 1, 60, null);
