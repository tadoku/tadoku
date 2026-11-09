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
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb1', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-09-11', '2026-09-30', '2026-09-20', 'Started yesterday', null, null, '{1}', true,
   '2026-08-01 12:00:00', '2026-08-01 12:00:00', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb2', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-09-12', '2026-09-30', '2026-09-20', 'Starts today', null, null, '{1}', true,
   '2026-08-02 12:00:00', '2026-08-02 12:00:00', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb3', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-09-13', '2026-09-30', '2026-09-20', 'Starts tomorrow', null, null, '{1}', true,
   '2026-08-03 12:00:00', '2026-08-03 12:00:00', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb4', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-11-01', '2026-11-30', '2026-11-15', 'Starts last, created first', null, null, '{1}', true,
   '2026-08-04 12:00:00', '2026-08-04 12:00:00', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb5', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-10-01', '2026-10-31', '2026-10-15', 'Starts next, created last', null, null, '{1}', true,
   '2026-08-05 12:00:00', '2026-08-05 12:00:00', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb6', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-09-20', '2026-09-30', '2026-09-25', 'Upcoming unofficial', null, null, '{1}', false,
   '2026-08-06 12:00:00', '2026-08-06 12:00:00', null),
  ('bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbb7', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-09-21', '2026-09-30', '2026-09-25', 'Upcoming deleted', null, null, '{1}', true,
   '2026-08-07 12:00:00', '2026-08-07 12:00:00', '2026-08-08 12:00:00');
