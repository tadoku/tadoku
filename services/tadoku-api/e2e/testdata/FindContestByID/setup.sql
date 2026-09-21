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
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '11111111-1111-4111-8111-111111111111', 'stale', true,
   '2026-02-01', '2026-02-28', '2026-02-15', 'Private contest', 'visible detail', '{jpn,eng}', '{2,1}', false,
   '2026-09-12 10:00:00', '2026-09-12 10:01:00', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2', '22222222-2222-4222-8222-222222222222', 'stale', false,
   '2026-03-01', '2026-03-31', '2026-02-15', 'Deleted contest', null, null, '{5}', true,
   '2026-09-12 11:00:00', '2026-09-12 11:01:00', '2026-09-12 11:02:00'),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-04-01', '2026-04-30', '2026-03-15', 'Invalid activity contest', null, '{}', '{999}', false,
   '2026-09-12 12:00:00', '2026-09-12 12:01:00', null);
