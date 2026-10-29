with target as (
  select tenant, id, display_name, deletion_locked_at, deleted_at
  from users
  where tenant = 'tadoku/prod'
    and id = '11111111-1111-4111-8111-111111111111'
), finished_contests as (
  select contests.id
  from contests
  join target on target.tenant = contests.tenant
  where contests.contest_end < '2026-09-12'
)
select
  target.display_name,
  target.deletion_locked_at,
  target.deleted_at,
  (select count(*) from contest_registrations
    where tenant = target.tenant and user_id = target.id) as kept_registrations,
  (select count(*) from logs
    where tenant = target.tenant and user_id = target.id) as kept_logs,
  (select count(*) from contest_registrations
    where tenant = target.tenant
      and user_id = target.id
      and contest_id not in (select id from finished_contests)) as other_registrations,
  (select count(*) from contest_logs
    join logs on logs.tenant = contest_logs.tenant and logs.id = contest_logs.log_id
    where logs.tenant = target.tenant
      and logs.user_id = target.id
      and contest_logs.contest_id not in (select id from finished_contests)) as other_contest_logs,
  (select count(*) from logs
    where logs.tenant = target.tenant
      and logs.user_id = target.id
      and (
        logs.deleted_at is not null
        or logs.frozen_at is null
        or logs.description is not null
        or not exists (
          select 1 from contest_logs
          where contest_logs.tenant = logs.tenant and contest_logs.log_id = logs.id
        )
      )) as unscrubbed_logs,
  (select count(*) from log_tags
    where tenant = target.tenant and user_id = target.id) as log_tags,
  (select count(*) from permission_relationships
    where tenant = target.tenant
      and subject_type = 'user'
      and subject_id = target.id::text) as permission_relationships,
  (select count(*) from accounts
    where tenant = target.tenant and user_id = target.id) as accounts,
  (select count(*) from sessions
    where tenant = target.tenant and user_id = target.id) as sessions,
  (select count(*) from recovery_tokens
    where tenant = target.tenant and user_id = target.id) as recovery_tokens,
  (select count(*) from contests
    where tenant = target.tenant
      and owner_user_id = target.id
      and owner_user_display_name <> 'Deleted organizer') as named_contests
from target
order by target.id
