begin;

insert into users (id, display_name, created_at, updated_at)
values
  (:'admin_user_id'::uuid, 'Dev Admin', now(), now()),
  (:'reader_user_id'::uuid, 'Dev Reader', now(), now())
on conflict (id) do update
set
  display_name = excluded.display_name,
  updated_at = now();

insert into user_roles (user_id, role, updated_at)
values
  (:'admin_user_id'::uuid, 'admin', now())
on conflict (user_id) do update
set
  role = excluded.role,
  updated_at = now();

insert into contests (
  id,
  owner_user_id,
  owner_user_display_name,
  "private",
  contest_start,
  contest_end,
  registration_end,
  title,
  "description",
  language_code_allow_list,
  activity_type_id_allow_list,
  official,
  created_at,
  updated_at
)
values
  (
    '00000000-0000-4000-8000-000000000101',
    :'admin_user_id'::uuid,
    'Dev Admin',
    false,
    date_trunc('year', current_date)::date,
    (date_trunc('year', current_date) + interval '1 year - 1 day')::date,
    current_date + interval '30 days',
    'Dev Tadoku Round',
    'A seeded contest for local and shared development.',
    array['jpn', 'spa', 'deu']::varchar(10)[],
    array[1, 2, 3, 4, 5]::integer[],
    true,
    now(),
    now()
  ),
  (
    '00000000-0000-4000-8000-000000000102',
    :'reader_user_id'::uuid,
    'Dev Reader',
    true,
    current_date - interval '7 days',
    current_date + interval '21 days',
    current_date + interval '7 days',
    'Private Reading Sprint',
    'A private seeded contest for owner/admin flows.',
    array['jpn']::varchar(10)[],
    array[1, 2]::integer[],
    false,
    now(),
    now()
  )
on conflict (id) do update
set
  owner_user_id = excluded.owner_user_id,
  owner_user_display_name = excluded.owner_user_display_name,
  "private" = excluded."private",
  contest_start = excluded.contest_start,
  contest_end = excluded.contest_end,
  registration_end = excluded.registration_end,
  title = excluded.title,
  "description" = excluded."description",
  language_code_allow_list = excluded.language_code_allow_list,
  activity_type_id_allow_list = excluded.activity_type_id_allow_list,
  official = excluded.official,
  updated_at = now(),
  deleted_at = null;

insert into contest_registrations (
  id,
  contest_id,
  user_id,
  language_codes,
  created_at,
  updated_at
)
values
  (
    '00000000-0000-4000-8000-000000000201',
    '00000000-0000-4000-8000-000000000101',
    :'admin_user_id'::uuid,
    array['jpn', 'spa']::varchar(10)[],
    now(),
    now()
  ),
  (
    '00000000-0000-4000-8000-000000000202',
    '00000000-0000-4000-8000-000000000101',
    :'reader_user_id'::uuid,
    array['jpn']::varchar(10)[],
    now(),
    now()
  ),
  (
    '00000000-0000-4000-8000-000000000203',
    '00000000-0000-4000-8000-000000000102',
    :'reader_user_id'::uuid,
    array['jpn']::varchar(10)[],
    now(),
    now()
  )
on conflict (user_id, contest_id) do update
set
  language_codes = excluded.language_codes,
  updated_at = now(),
  deleted_at = null;

insert into logs (
  id,
  user_id,
  language_code,
  log_activity_id,
  unit_id,
  "description",
  amount,
  modifier,
  computed_score,
  eligible_official_leaderboard,
  duration_seconds,
  created_at,
  updated_at
)
values
  (
    '00000000-0000-4000-8000-000000000301',
    :'admin_user_id'::uuid,
    'jpn',
    1,
    (select id from log_units where log_activity_id = 1 and name = 'Page' and language_code is null limit 1),
    'Seeded reading log',
    42,
    1,
    42,
    true,
    null,
    now() - interval '2 days',
    now()
  ),
  (
    '00000000-0000-4000-8000-000000000302',
    :'reader_user_id'::uuid,
    'jpn',
    2,
    null,
    'Seeded listening log',
    null,
    null,
    30,
    true,
    3600,
    now() - interval '1 day',
    now()
  ),
  (
    '00000000-0000-4000-8000-000000000303',
    :'reader_user_id'::uuid,
    'spa',
    1,
    (select id from log_units where log_activity_id = 1 and name = 'Page' and language_code is null limit 1),
    'Seeded Spanish reading',
    18,
    1,
    18,
    true,
    null,
    now(),
    now()
  )
on conflict (id) do update
set
  user_id = excluded.user_id,
  language_code = excluded.language_code,
  log_activity_id = excluded.log_activity_id,
  unit_id = excluded.unit_id,
  "description" = excluded."description",
  amount = excluded.amount,
  modifier = excluded.modifier,
  computed_score = excluded.computed_score,
  eligible_official_leaderboard = excluded.eligible_official_leaderboard,
  duration_seconds = excluded.duration_seconds,
  updated_at = now(),
  deleted_at = null;

insert into contest_logs (
  contest_id,
  log_id,
  amount,
  modifier,
  duration_seconds,
  computed_score
)
values
  ('00000000-0000-4000-8000-000000000101', '00000000-0000-4000-8000-000000000301', 42, 1, null, 42),
  ('00000000-0000-4000-8000-000000000101', '00000000-0000-4000-8000-000000000302', null, null, 3600, 30),
  ('00000000-0000-4000-8000-000000000101', '00000000-0000-4000-8000-000000000303', 18, 1, null, 18)
on conflict (contest_id, log_id) do update
set
  amount = excluded.amount,
  modifier = excluded.modifier,
  duration_seconds = excluded.duration_seconds,
  computed_score = excluded.computed_score;

insert into log_tags (log_id, user_id, tag, created_at)
values
  ('00000000-0000-4000-8000-000000000301', :'admin_user_id'::uuid, 'book', now()),
  ('00000000-0000-4000-8000-000000000302', :'reader_user_id'::uuid, 'podcast', now()),
  ('00000000-0000-4000-8000-000000000303', :'reader_user_id'::uuid, 'fiction', now())
on conflict (log_id, tag) do nothing;

commit;
