\if :{?tenant}
\else
  do $$ begin raise exception 'tenant psql variable is required'; end $$;
\endif

begin;
select set_config('tadoku.tenant', :'tenant', true);
select current_user = 'tadoku'
       and :'tenant' ~ '^tadoku/[a-z0-9][a-z0-9-]*-[0-9a-f]{8}$'
       and exists(select from tenants where key = :'tenant' and kind = 'test') as allowed
\gset
\if :allowed
\else
  do $$ begin raise exception 'refusing seed outside a registered branch tenant with the runtime role'; end $$;
\endif
commit;

\ir /seed/immersion.sql
\ir /seed/content.sql

-- Sixty participants scoring far above anyone who logs a little in the seeded round, so a browser test's learner ranks
-- below the first standings page. Accounts cannot do this: registration allows ten per tenant and client IP.
begin;
select set_config('tadoku.tenant', :'tenant', true);

with field as (
  select
    n,
    uuid_generate_v5('00000000-0000-4000-8000-000000000401'::uuid, :'tenant' || '/user/' || n) as user_id,
    uuid_generate_v5('00000000-0000-4000-8000-000000000402'::uuid, :'tenant' || '/registration/' || n) as registration_id,
    uuid_generate_v5('00000000-0000-4000-8000-000000000403'::uuid, :'tenant' || '/log/' || n) as log_id,
    uuid_generate_v5('00000000-0000-4000-8000-000000000101'::uuid, :'tenant') as contest_id
  from generate_series(1, 60) as n
), field_users as (
  insert into users (tenant, id, display_name, created_at, updated_at)
  select :'tenant', user_id, 'Field Reader ' || n, now(), now()
  from field
  on conflict (tenant, id) do nothing
), field_accounts as (
  -- A public profile needs an account. The unusable password hash means nobody can sign in to these.
  insert into accounts (tenant, user_id, email, email_normalized, password_hash, created_at, updated_at)
  select :'tenant', user_id, 'field-reader-' || n || '@example.invalid', 'field-reader-' || n || '@example.invalid', '!', now(), now()
  from field
  on conflict (tenant, user_id) do nothing
), field_registrations as (
  insert into contest_registrations (tenant, id, contest_id, user_id, language_codes, created_at, updated_at)
  select :'tenant', registration_id, contest_id, user_id, array['jpn']::varchar(10)[], now(), now()
  from field
  on conflict (tenant, user_id, contest_id) do nothing
), field_logs as (
  insert into logs (
    tenant, id, user_id, language_code, log_activity_id, unit_key, "description", amount, modifier, computed_score,
    eligible_official_leaderboard, duration_seconds, created_at, updated_at
  )
  select :'tenant', log_id, user_id, 'jpn', 1, 'pages', 'Seeded field reading', 1000 + n, 1, 1000 + n, true, null, now(), now()
  from field
  on conflict (tenant, id) do nothing
)
insert into contest_logs (tenant, contest_id, log_id, unit_key, amount, modifier, duration_seconds, computed_score)
select :'tenant', contest_id, log_id, 'pages', 1000 + n, 1, null, 1000 + n
from field
on conflict (tenant, contest_id, log_id) do nothing;

commit;

-- A finished round holding a Dev Reader log, and a started round scored by its own override rule set. The API attaches
-- logs only to started, unfinished contests and refuses scoring changes from a contest's first day, so no run can
-- create either one.
begin;
select set_config('tadoku.tenant', :'tenant', true);

select
  uuid_generate_v5('00000000-0000-4000-8000-000000000104'::uuid, :'tenant') as finished_contest_id,
  uuid_generate_v5('00000000-0000-4000-8000-000000000204'::uuid, :'tenant') as finished_registration_id,
  uuid_generate_v5('00000000-0000-4000-8000-000000000304'::uuid, :'tenant') as finished_log_id,
  uuid_generate_v5('00000000-0000-4000-8000-000000000105'::uuid, :'tenant') as scoring_contest_id,
  uuid_generate_v5('00000000-0000-4000-8000-000000000501'::uuid, :'tenant') as scoring_rule_set_id
\gset

insert into contests (
  tenant, id, owner_user_id, owner_user_display_name, "private", contest_start, contest_end, registration_end, title,
  "description", language_code_allow_list, activity_type_id_allow_list, official, created_at, updated_at
)
values
  (
    :'tenant', :'finished_contest_id'::uuid, :'admin_user_id'::uuid, 'Dev Admin', false,
    current_date - 14, current_date - 7, current_date - 7, 'Dev Finished Round',
    'A seeded community contest that has ended.', array['jpn']::varchar(10)[], array[1, 2]::integer[], false, now(), now()
  ),
  (
    :'tenant', :'scoring_contest_id'::uuid, :'admin_user_id'::uuid, 'Dev Admin', false,
    current_date, current_date + 30, current_date + 30, 'Dev Scoring Round',
    'A seeded community contest with its own scoring rules.', array['jpn']::varchar(10)[], array[1, 2]::integer[], false,
    now(), now()
  )
on conflict (tenant, id) do nothing;

insert into contest_registrations (tenant, id, contest_id, user_id, language_codes, created_at, updated_at)
values (
  :'tenant', :'finished_registration_id'::uuid, :'finished_contest_id'::uuid, :'reader_user_id'::uuid,
  array['jpn']::varchar(10)[], now(), now()
)
on conflict (tenant, user_id, contest_id) do nothing;

insert into logs (
  tenant, id, user_id, language_code, log_activity_id, unit_key, "description", amount, modifier, computed_score,
  eligible_official_leaderboard, duration_seconds, created_at, updated_at
)
values (
  :'tenant', :'finished_log_id'::uuid, :'reader_user_id'::uuid, 'jpn', 1, 'pages', 'Finished round reading', 10, 1, 10,
  false, null, current_date - 10 + time '12:00', current_date - 10 + time '12:00'
)
on conflict (tenant, id) do nothing;

insert into contest_logs (tenant, contest_id, log_id, unit_key, amount, modifier, duration_seconds, computed_score)
values (:'tenant', :'finished_contest_id'::uuid, :'finished_log_id'::uuid, 'pages', 10, 1, null, 10)
on conflict (tenant, contest_id, log_id) do nothing;

-- Reading scores three points a page, doubled when tagged webv3-bonus; every other input falls back to the platform
-- version that was active when the tenant was seeded.
insert into scoring_rule_sets (
  tenant, id, scope, contest_id, version, status, mode, fallback_rule_set_id, created_at, published_at
)
select
  :'tenant', :'scoring_rule_set_id'::uuid, 'contest', :'scoring_contest_id'::uuid, 1, 'published', 'override',
  active_rule_set_id, now(), now()
from platform_scoring_config
on conflict (tenant, id) do nothing;

insert into scoring_rules (tenant, rule_set_id, priority, stackable, activity_id, tag, score_source, rate)
values
  (:'tenant', :'scoring_rule_set_id'::uuid, 0, false, 1, null, 'amount', 3),
  (:'tenant', :'scoring_rule_set_id'::uuid, 10, true, 1, 'webv3-bonus', 'amount', 2)
on conflict (tenant, rule_set_id, priority) do nothing;

update contests
set scoring_rule_set_id = :'scoring_rule_set_id'::uuid
where id = :'scoring_contest_id'::uuid and scoring_rule_set_id is null;

commit;
