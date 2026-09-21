insert into users (id, display_name, created_at, updated_at, deleted_at)
values
  ('11111111-1111-4111-8111-111111111111', 'Owner One', '2026-01-01', '2026-01-01', null),
  ('22222222-2222-4222-8222-222222222222', 'Owner Two', '2026-01-01', '2026-01-01', '2026-09-01');

insert into contests (
  id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
  registration_end, title, "description", language_code_allow_list,
  activity_type_id_allow_list, official, created_at, updated_at, deleted_at
)
values
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa1', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-01-01', '2026-01-31', '2026-01-15', 'Public official', null, null, '{1}', true,
   '2026-09-12 11:00:00', '2026-09-12 11:01:00', null),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa2', '22222222-2222-4222-8222-222222222222', 'stale', true,
   '2026-02-01', '2026-02-28', '2026-02-15', 'Private official', 'private', '{jpn,eng}', '{2,1}', true,
   '2026-09-12 10:00:00', '2026-09-12 10:01:00', null),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa3', '22222222-2222-4222-8222-222222222222', 'stale', false,
   '2025-01-01', '2025-01-31', '2024-12-15', 'Deleted official', null, '{}', '{5}', true,
   '2026-09-12 09:00:00', '2026-09-12 09:01:00', '2026-09-12 09:02:00'),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa4', '11111111-1111-4111-8111-111111111111', 'stale', true,
   '2026-03-01', '2026-03-31', '2026-02-15', 'Private unofficial', null, '{eng}', '{3}', false,
   '2026-09-12 12:00:00', '2026-09-12 12:01:00', null),
  ('aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaa5', '22222222-2222-4222-8222-222222222222', 'stale', false,
   '2026-04-01', '2026-04-30', '2026-03-15', 'Public unofficial', null, '{jpn}', '{4}', false,
   '2026-09-12 08:00:00', '2026-09-12 08:01:00', null);
