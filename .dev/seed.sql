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
