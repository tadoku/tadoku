select
  (select json_agg(json_build_object(
      'actor_id', audit.user_id,
      'action', audit.action,
      'metadata', audit.metadata,
      'description', audit.description,
      'recorded_at', audit.created_at
    ) order by audit.action)
    from moderation_audit_log as audit
    where audit.tenant = tenants.key) as audit,
  (select json_agg(json_build_object(
      'template', emails.template,
      'recipient', emails.recipient,
      'subject', emails.subject,
      'body', emails.body,
      'params', emails.params,
      'status', emails.status,
      'created_at', emails.created_at
    ) order by emails.created_at)
    from emails
    where emails.tenant = tenants.key) as emails,
  (select json_agg(json_build_object(
      'task_type', jobs.task_type,
      'state', jobs.state,
      'payload', jobs.payload - 'email_id',
      'names_email', jobs.payload ? 'email_id'
    ) order by jobs.task_type, jobs.payload::text)
    from jobs
    where jobs.tenant = tenants.key) as jobs
from tenants
where tenants.key = 'tadoku/prod'
order by tenants.key
