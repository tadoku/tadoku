insert into users (id, display_name, created_at, updated_at)
values ('11111111-1111-4111-8111-111111111111', 'Owner One', '2026-01-01', '2026-01-01');

insert into contests (
  id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
  registration_end, title, "description", language_code_allow_list,
  activity_type_id_allow_list, official, created_at, updated_at, deleted_at
)
values
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-01-01', '2026-01-31', '2025-12-15', 'Live official', null, '{jpn}', '{1}', true,
   '2026-09-12 10:00:00', '2026-09-12 10:01:00', null),
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc2', '11111111-1111-4111-8111-111111111111', 'stale', true,
   '2027-01-01', '2027-01-31', '2026-12-15', 'Future deleted official', 'latest', '{eng,jpn}', '{5,2}', true,
   '2026-09-12 11:00:00', '2026-09-12 11:01:00', '2026-09-12 11:02:00'),
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc3', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2028-01-01', '2028-01-31', '2027-12-15', 'Future unofficial', null, null, '{3}', false,
   '2026-09-12 12:00:00', '2026-09-12 12:01:00', null);
