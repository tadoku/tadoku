select
  users.display_name,
  users.deletion_locked_at,
  users.deleted_at,
  (select json_agg(json_build_object(
      'task_type', jobs.task_type,
      'state', jobs.state,
      'payload', jobs.payload
    ) order by jobs.id)
    from jobs
    where jobs.tenant = users.tenant) as jobs,
  (select json_agg(json_build_object(
      'actor_id', audit.user_id,
      'action', audit.action,
      'metadata', audit.metadata,
      'description', audit.description,
      'recorded_at', audit.created_at
    ) order by audit.action)
    from moderation_audit_log as audit
    where audit.tenant = users.tenant) as audit
from users
where users.tenant = 'tadoku/prod'
  and users.id = '33333333-3333-4333-8333-333333333333'
order by users.id
