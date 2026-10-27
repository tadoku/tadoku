select
  users.display_name,
  users.deletion_locked_at,
  users.deleted_at,
  (select count(*) from accounts
    where accounts.tenant = users.tenant and accounts.user_id = users.id) as accounts,
  (select count(*) from sessions
    where sessions.tenant = users.tenant and sessions.user_id = users.id) as sessions,
  (select count(*) from recovery_tokens
    where recovery_tokens.tenant = users.tenant and recovery_tokens.user_id = users.id) as recovery_tokens,
  (select count(*) from permission_relationships
    where permission_relationships.tenant = users.tenant
      and permission_relationships.subject_type = 'user'
      and permission_relationships.subject_id = users.id::text) as permission_relationships,
  (select count(*) from log_tags
    where log_tags.tenant = users.tenant and log_tags.user_id = users.id) as log_tags,
  (select json_agg(json_build_object(
      'id', logs.id,
      'description', logs.description,
      'frozen_at', logs.frozen_at,
      'deleted_at', logs.deleted_at,
      'contests', (select json_agg(contest_logs.contest_id order by contest_logs.contest_id)
        from contest_logs
        where contest_logs.tenant = logs.tenant and contest_logs.log_id = logs.id)
    ) order by logs.id)
    from logs
    where logs.tenant = users.tenant and logs.user_id = users.id) as logs,
  (select json_agg(registrations.contest_id order by registrations.contest_id)
    from contest_registrations as registrations
    where registrations.tenant = users.tenant and registrations.user_id = users.id) as registrations,
  (select json_agg(json_build_object(
      'id', contests.id,
      'organizer', contests.owner_user_display_name
    ) order by contests.id)
    from contests
    where contests.tenant = users.tenant and contests.owner_user_id = users.id) as owned_contests
from users
where users.tenant = 'tadoku/prod'
  and users.id = '33333333-3333-4333-8333-333333333333'
order by users.id
