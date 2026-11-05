insert into users (id, display_name, created_at, updated_at) values
  ('11111111-1111-4111-8111-111111111111', 'Branch Reader', '2026-01-01', '2026-01-01');

insert into contests (
  id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
  registration_end, title, language_code_allow_list, activity_type_id_allow_list, official,
  created_at, updated_at
) values
  ('f1000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'Branch Reader', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Branch contest', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01');

insert into logs (
  id, user_id, language_code, log_activity_id, unit_key, amount,
  modifier, computed_score, eligible_official_leaderboard, description, created_at, updated_at
) values
  ('d1000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   'pages', 7, 1, 14, false,
   'Branch-only preview', '2026-09-12', '2026-09-12');

insert into contest_logs (contest_id, log_id, unit_key, amount, modifier, computed_score) values
  ('f1000000-0000-4000-8000-000000000001', 'd1000000-0000-4000-8000-000000000001', 'pages', 7, 8, 56);

insert into log_tags (log_id, user_id, tag) values
  ('d1000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'branch');
