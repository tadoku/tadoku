insert into users (id, display_name, created_at, updated_at, deleted_at) values
  ('11111111-1111-4111-8111-111111111111', 'Stored Reader', '2026-01-01', '2026-01-01', null),
  ('22222222-2222-4222-8222-222222222222', 'Admin One', '2026-01-01', '2026-01-01', null),
  ('99999999-9999-4999-8999-999999999999', 'Deleted Owner Raw', '2026-01-01', '2026-01-01', '2026-09-01');
insert into contests (id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title, language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at, deleted_at) values
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '22222222-2222-4222-8222-222222222222', 'stale', false, '2026-09-01', '2026-09-30', '2026-08-31', 'Live contest', null, '{1}', false, '2026-01-01', '2026-01-01', null),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '99999999-9999-4999-8999-999999999999', 'stale', true, '2026-09-01', '2026-09-30', '2026-08-31', 'Deleted contest', null, '{2}', false, '2026-01-01', '2026-01-01', '2026-09-02');
insert into contest_registrations (id, contest_id, user_id, language_codes, deleted_at) values
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', '{jpn,eng}', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '11111111-1111-4111-8111-111111111111', '{eng}', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb9', 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa9', '11111111-1111-4111-8111-111111111111', '{eng}', null);
