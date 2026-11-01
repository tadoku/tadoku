insert into users (id, display_name, created_at, updated_at, deleted_at)
values ('11111111-1111-4111-8111-111111111111', 'Owner One', '2026-01-01', '2026-01-01', '2026-09-12 09:00:00');

insert into contests (
  id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end,
  registration_end, title, "description", language_code_allow_list,
  activity_type_id_allow_list, official, created_at, updated_at, deleted_at
)
values
  ('cccccccc-cccc-4ccc-8ccc-ccccccccccc1', '11111111-1111-4111-8111-111111111111', 'stale', false,
   '2026-01-01', '2026-01-31', '2025-12-15', 'Live official', null, '{jpn}', '{1}', true,
   '2026-09-12 10:00:00', '2026-09-12 10:01:00', null);
