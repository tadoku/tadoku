\if :{?tenant}
\else
  do $$ begin raise exception 'tenant psql variable is required'; end $$;
\endif

begin;
select set_config('tadoku.tenant', :'tenant', true);

select
  case
    when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000101'::uuid
    else uuid_generate_v5('00000000-0000-4000-8000-000000000101'::uuid, :'tenant')
  end as public_contest_id,
  case
    when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000102'::uuid
    else uuid_generate_v5('00000000-0000-4000-8000-000000000102'::uuid, :'tenant')
  end as private_contest_id,
  case
    when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000201'::uuid
    else uuid_generate_v5('00000000-0000-4000-8000-000000000201'::uuid, :'tenant')
  end as admin_registration_id,
  case
    when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000202'::uuid
    else uuid_generate_v5('00000000-0000-4000-8000-000000000202'::uuid, :'tenant')
  end as reader_registration_id,
  case
    when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000203'::uuid
    else uuid_generate_v5('00000000-0000-4000-8000-000000000203'::uuid, :'tenant')
  end as private_registration_id,
  case
    when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000301'::uuid
    else uuid_generate_v5('00000000-0000-4000-8000-000000000301'::uuid, :'tenant')
  end as reading_log_id,
  case
    when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000302'::uuid
    else uuid_generate_v5('00000000-0000-4000-8000-000000000302'::uuid, :'tenant')
  end as listening_log_id,
  case
    when :'tenant' = 'tadoku/prod' then '00000000-0000-4000-8000-000000000303'::uuid
    else uuid_generate_v5('00000000-0000-4000-8000-000000000303'::uuid, :'tenant')
  end as spanish_log_id
\gset

-- Stable fixtures: reruns insert only missing rows and keep existing ones, including
-- user edits. Rows that reference a fixture user follow that user's current UUID.

insert into user_roles (tenant, user_id, role, updated_at)
values
  (:'tenant', :'admin_user_id'::uuid, 'admin', now())
on conflict (tenant, user_id) do update
set
  role = excluded.role,
  updated_at = now()
where user_roles.role is distinct from excluded.role;

insert into contests (
  tenant,
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
    :'tenant',
    :'public_contest_id'::uuid,
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
    :'tenant',
    :'private_contest_id'::uuid,
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
  updated_at = now()
where contests.owner_user_id is distinct from excluded.owner_user_id;

-- remove stale seed registrations left over from runs with different seed
-- identities, so the fixed ids below never collide on the primary key
delete from contest_registrations
where tenant = :'tenant'
  and (
    (id = :'admin_registration_id'::uuid and user_id <> :'admin_user_id'::uuid)
    or (id = :'reader_registration_id'::uuid and user_id <> :'reader_user_id'::uuid)
    or (id = :'private_registration_id'::uuid and user_id <> :'reader_user_id'::uuid)
  );

insert into contest_registrations (
  tenant,
  id,
  contest_id,
  user_id,
  language_codes,
  created_at,
  updated_at
)
values
  (
    :'tenant',
    :'admin_registration_id'::uuid,
    :'public_contest_id'::uuid,
    :'admin_user_id'::uuid,
    array['jpn', 'spa']::varchar(10)[],
    now(),
    now()
  ),
  (
    :'tenant',
    :'reader_registration_id'::uuid,
    :'public_contest_id'::uuid,
    :'reader_user_id'::uuid,
    array['jpn', 'spa']::varchar(10)[],
    now(),
    now()
  ),
  (
    :'tenant',
    :'private_registration_id'::uuid,
    :'private_contest_id'::uuid,
    :'reader_user_id'::uuid,
    array['jpn']::varchar(10)[],
    now(),
    now()
  )
on conflict (tenant, user_id, contest_id) do nothing;

insert into logs (
  tenant,
  id,
  user_id,
  language_code,
  log_activity_id,
  unit_id,
  unit_key,
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
    :'tenant',
    :'reading_log_id'::uuid,
    :'admin_user_id'::uuid,
    'jpn',
    1,
    (select id from log_units where log_activity_id = 1 and name = 'Page' and language_code is null limit 1),
    (select unit_key from log_units where log_activity_id = 1 and name = 'Page' and language_code is null limit 1),
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
    :'tenant',
    :'listening_log_id'::uuid,
    :'reader_user_id'::uuid,
    'jpn',
    2,
    null,
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
    :'tenant',
    :'spanish_log_id'::uuid,
    :'reader_user_id'::uuid,
    'spa',
    1,
    (select id from log_units where log_activity_id = 1 and name = 'Page' and language_code is null limit 1),
    (select unit_key from log_units where log_activity_id = 1 and name = 'Page' and language_code is null limit 1),
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
  updated_at = now()
where logs.user_id is distinct from excluded.user_id;

insert into contest_logs (
  tenant,
  contest_id,
  log_id,
  unit_key,
  amount,
  modifier,
  duration_seconds,
  computed_score
)
values
  (
    :'tenant',
    :'public_contest_id'::uuid,
    :'reading_log_id'::uuid,
    (select unit_key from logs where tenant = :'tenant' and id = :'reading_log_id'::uuid),
    42,
    1,
    null,
    42
  ),
  (
    :'tenant',
    :'public_contest_id'::uuid,
    :'listening_log_id'::uuid,
    (select unit_key from logs where tenant = :'tenant' and id = :'listening_log_id'::uuid),
    null,
    null,
    3600,
    30
  ),
  (
    :'tenant',
    :'public_contest_id'::uuid,
    :'spanish_log_id'::uuid,
    (select unit_key from logs where tenant = :'tenant' and id = :'spanish_log_id'::uuid),
    18,
    1,
    null,
    18
  )
on conflict (tenant, contest_id, log_id) do nothing;

insert into log_tags (tenant, log_id, user_id, tag, created_at)
values
  (:'tenant', :'reading_log_id'::uuid, :'admin_user_id'::uuid, 'book', now()),
  (:'tenant', :'listening_log_id'::uuid, :'reader_user_id'::uuid, 'podcast', now()),
  (:'tenant', :'spanish_log_id'::uuid, :'reader_user_id'::uuid, 'fiction', now())
on conflict (tenant, log_id, tag) do update
set user_id = excluded.user_id
where log_tags.user_id is distinct from excluded.user_id;

commit;
