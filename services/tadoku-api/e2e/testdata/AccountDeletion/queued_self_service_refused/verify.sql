select
  users.display_name,
  users.deletion_locked_at,
  users.deleted_at,
  (select count(*) from accounts
    where accounts.tenant = users.tenant and accounts.user_id = users.id) as accounts,
  (select json_agg(json_build_object(
      'task_type', jobs.task_type,
      'state', jobs.state,
      'last_error', jobs.last_error
    ) order by jobs.id)
    from jobs) as jobs,
  (select count(*) from moderation_audit_log) as audit_events
from users
where users.tenant = 'tadoku/prod'
  and users.id = '11111111-1111-4111-8111-111111111111'
order by users.id
