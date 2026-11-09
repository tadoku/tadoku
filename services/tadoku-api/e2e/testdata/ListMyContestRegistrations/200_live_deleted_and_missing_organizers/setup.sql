insert into users (id, display_name, created_at, updated_at, deleted_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-01-01', null),
  ('77777777-7777-4777-8777-777777777777', 'Missing Owner', '2026-01-01', '2026-01-01', null),
  ('88888888-8888-4888-8888-888888888888', 'Tombstone Name', '2026-01-01', '2026-01-01', '2026-06-01'),
  ('99999999-9999-4999-8999-999999999999', 'Raw Owner Name', '2026-01-01', '2026-01-01', null);
insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at, deleted_at) values
  ('f0000000-0000-4000-8000-000000000001', '99999999-9999-4999-8999-999999999999', 'stale', false, '2026-03-01', '2026-03-31', '2026-03-31', 'Live organizer', null, '{1}', false, '2026-01-01', '2026-01-01', null),
  ('f0000000-0000-4000-8000-000000000002', '88888888-8888-4888-8888-888888888888', 'stale', false, '2026-02-01', '2026-02-28', '2026-02-28', 'Deleted organizer', null, '{1}', false, '2026-01-01', '2026-01-01', null),
  ('f0000000-0000-4000-8000-000000000003', '77777777-7777-4777-8777-777777777777', 'stale', false, '2026-01-01', '2026-01-31', '2026-01-31', 'Missing organizer', null, '{1}', false, '2026-01-01', '2026-01-01', null);
insert into contest_registrations (id, contest_id, user_id, language_codes) values
  ('eeeeeeee-eeee-4eee-8eee-000000000001', 'f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', '{eng}'),
  ('eeeeeeee-eeee-4eee-8eee-000000000002', 'f0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', '{eng}'),
  ('eeeeeeee-eeee-4eee-8eee-000000000003', 'f0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', '{eng}');

delete from users where id = '77777777-7777-4777-8777-777777777777';
