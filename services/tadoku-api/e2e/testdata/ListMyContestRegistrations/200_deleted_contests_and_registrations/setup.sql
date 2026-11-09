insert into users (id, display_name, created_at, updated_at, deleted_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-01-01', null),
  ('44444444-4444-4444-8444-444444444444', 'Other Reader', '2026-01-01', '2026-01-01', null),
  ('99999999-9999-4999-8999-999999999999', 'Raw Owner Name', '2026-01-01', '2026-01-01', null);
insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at, deleted_at) values
  ('f0000000-0000-4000-8000-000000000001', '99999999-9999-4999-8999-999999999999', 'stale', false, '2025-01-01', '2025-01-31', '2025-01-31', 'Deleted public', null, '{1}', false, '2025-01-01', '2025-01-01', '2025-02-01'),
  ('f0000000-0000-4000-8000-000000000002', '99999999-9999-4999-8999-999999999999', 'stale', true, '2025-02-01', '2025-02-28', '2025-02-28', 'Deleted private', null, '{1}', false, '2025-01-01', '2025-01-01', '2025-03-01'),
  ('f0000000-0000-4000-8000-000000000003', '99999999-9999-4999-8999-999999999999', 'stale', false, '2025-03-01', '2025-03-31', '2025-03-31', 'Withdrawn', null, '{1}', false, '2025-01-01', '2025-01-01', null);
insert into contest_registrations (id, contest_id, user_id, language_codes, deleted_at) values
  ('eeeeeeee-eeee-4eee-8eee-000000000001', 'f0000000-0000-4000-8000-000000000001', '11111111-1111-4111-8111-111111111111', '{eng}', null),
  ('eeeeeeee-eeee-4eee-8eee-000000000002', 'f0000000-0000-4000-8000-000000000002', '11111111-1111-4111-8111-111111111111', '{eng}', null),
  ('eeeeeeee-eeee-4eee-8eee-000000000003', 'f0000000-0000-4000-8000-000000000003', '11111111-1111-4111-8111-111111111111', '{eng}', '2025-03-02');
