insert into users (id, display_name, created_at, updated_at, deleted_at) values
  ('11111111-1111-4111-8111-111111111111', 'Cedar', '2026-01-01', '2026-01-01', null);

insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at) values
  ('f1111111-1111-4111-8111-111111111111', '11111111-1111-4111-8111-111111111111', 'Cedar', false, '2026-01-01', '2026-12-31', '2026-01-31', 'Leaderboard fixture', '{jpn,eng}', '{1,2}', true, '2026-01-01', '2026-01-01');
