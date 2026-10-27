select
  users.display_name,
  users.deletion_locked_at,
  users.deleted_at,
  (select count(*) from accounts
    where accounts.tenant = users.tenant and accounts.user_id = users.id) as accounts,
  (select count(*) from permission_relationships
    where permission_relationships.tenant = users.tenant
      and permission_relationships.subject_type = 'user'
      and permission_relationships.subject_id = users.id::text) as permission_relationships,
  (select json_agg(json_build_object('task_type', jobs.task_type, 'state', jobs.state) order by jobs.id)
    from jobs
    where jobs.tenant = users.tenant) as jobs,
  (select json_agg(audit.action order by audit.action)
    from moderation_audit_log as audit
    where audit.tenant = users.tenant) as audit_actions,
  (select json_agg(emails.template order by emails.created_at, emails.template)
    from emails
    where emails.tenant = users.tenant
      and emails.recipient = 'reader@example.test') as emails
from users
where users.tenant = 'tadoku/prod'
  and users.id = '11111111-1111-4111-8111-111111111111'
order by users.id
