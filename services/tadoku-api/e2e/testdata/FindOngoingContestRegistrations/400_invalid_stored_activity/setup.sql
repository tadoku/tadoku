insert into users (id, display_name, created_at, updated_at, deleted_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-01-01', null),
  ('33333333-3333-4333-8333-333333333333', 'Banned One', '2026-01-01', '2026-01-01', null),
  ('99999999-9999-4999-8999-999999999999', 'Raw Owner Name', '2026-01-01', '2026-01-01', null);
insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at, deleted_at) values
  ('f0000000-0000-4000-8000-000000000006', '99999999-9999-4999-8999-999999999999', 'stale', false, '2026-09-01', '2026-09-30', '2026-08-31', 'Invalid activity', null, '{99}', true, '2026-01-01', '2026-01-01', null);
insert into contest_registrations (id, contest_id, user_id, language_codes) values
  ('eeeeeeee-eeee-4eee-8eee-000000000006', 'f0000000-0000-4000-8000-000000000006', '11111111-1111-4111-8111-111111111111', '{eng}');
