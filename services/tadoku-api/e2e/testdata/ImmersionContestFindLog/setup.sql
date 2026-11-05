insert into users (id, display_name, created_at, updated_at) values
  ('11111111-1111-4111-8111-111111111111', 'Reader One', '2026-01-01', '2026-01-01');

insert into contests (
  id, owner_user_id, owner_user_display_name, private, contest_start, contest_end,
  registration_end, title, language_code_allow_list, activity_type_id_allow_list, official,
  created_at, updated_at
) values
  ('f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Contest X', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('f0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', 'Reader One', true,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Contest Y', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('f0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Empty contest', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01'),
  ('f0000000-0000-4000-8000-000000000004', '11111111-1111-4111-8111-111111111111', 'Reader One', false,
   '2026-09-01', '2026-09-30', '2026-09-20', 'Deleted contest', '{eng,jpn}', '{1}', false,
   '2026-01-01', '2026-01-01');

update contests set deleted_at = '2026-09-10' where id = 'f0000000-0000-4000-8000-000000000004';

insert into logs (
  id, user_id, language_code, log_activity_id, unit_key, amount,
  modifier, computed_score, eligible_official_leaderboard, description, created_at, updated_at
) values
  ('d0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   'pages', 10, 1, 20, false,
   'Contest preview', '2026-09-12', '2026-09-12'),
  ('d0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   'pages', 3, 1, 6, false,
   'Unattached personal log', '2026-09-12', '2026-09-12'),
  ('d0000000-0000-4000-8000-000000000004', '11111111-1111-4111-8111-111111111111', 'eng', 1,
   'pages', 4, 1, 8, false,
   'Deleted log', '2026-09-12', '2026-09-12');

update logs set deleted_at = '2026-09-12' where id = 'd0000000-0000-4000-8000-000000000004';

insert into contest_logs (contest_id, log_id, unit_key, amount, modifier, computed_score) values
  ('f0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000001', 'pages', 10, 4, 40),
  ('f0000000-0000-4000-8000-000000000002', 'd0000000-0000-4000-8000-000000000001', 'pages', 10, 6, 60),
  ('f0000000-0000-4000-8000-000000000001', 'd0000000-0000-4000-8000-000000000004', 'pages', 4, 4, 16);

insert into log_tags (log_id, user_id, tag) values
  ('d0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', 'book');

insert into contest_registrations (id, contest_id, user_id, language_codes) values
  ('e0000000-0000-4000-8000-000000000001', 'f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', '{eng,jpn}'),
  ('e0000000-0000-4000-8000-000000000002', 'f0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', '{eng,jpn}');
