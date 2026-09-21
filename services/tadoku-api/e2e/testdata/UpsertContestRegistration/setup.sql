insert into users (id, display_name, created_at, updated_at, deletion_locked_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-01-01', null),
  ('22222222-2222-4222-8222-222222222222', 'Admin One', '2026-01-01', '2026-01-01', '2026-09-01'),
  ('99999999-9999-4999-8999-999999999999', 'Owner', '2026-01-01', '2026-01-01', null);
insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at, deleted_at) values
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '99999999-9999-4999-8999-999999999999', 'stale', true, '2020-01-01', '2020-01-31', '2019-12-31', 'Ended private unrestricted', null, '{99}', false, '2020-01-01', '2020-01-01', null),
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc2', '99999999-9999-4999-8999-999999999999', 'stale', false, '2026-09-01', '2026-09-30', '2026-08-31', 'Existing registration', '{jpn,eng}', '{1}', true, '2026-01-01', '2026-01-01', null),
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc3', '99999999-9999-4999-8999-999999999999', 'stale', false, '2026-09-01', '2026-09-30', '2026-08-31', 'Restricted', '{jpn}', '{1}', false, '2026-01-01', '2026-01-01', null),
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc4', '99999999-9999-4999-8999-999999999999', 'stale', false, '2026-09-01', '2026-09-30', '2026-08-31', 'Deleted', null, '{1}', false, '2026-01-01', '2026-01-01', '2026-09-01'),
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc5', '99999999-9999-4999-8999-999999999999', 'stale', false, '2026-09-01', '2026-09-30', '2026-08-31', 'Soft deleted registration', null, '{1}', false, '2026-01-01', '2026-01-01', null);
insert into contest_registrations (id, contest_id, user_id, language_codes, deleted_at) values
  ('dddddddd-dddd-4ddd-8ddd-ddddddddddd2', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc2', '11111111-1111-4111-8111-111111111111', '{jpn,eng}', null),
  ('dddddddd-dddd-4ddd-8ddd-ddddddddddd5', 'cccccccc-cccc-4ccc-8ccc-ccccccccccc5', '11111111-1111-4111-8111-111111111111', '{eng}', '2026-09-01');
