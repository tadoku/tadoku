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
  (select json_agg(jobs.state order by jobs.id)
    from jobs
    where jobs.tenant = users.tenant
      and jobs.task_type = 'account.delete.v1') as deletion_jobs,
  (select json_agg(audit.action order by audit.action)
    from moderation_audit_log as audit
    where audit.tenant = users.tenant) as audit_actions
from users
where users.tenant = 'tadoku/prod'
  and users.id = '33333333-3333-4333-8333-333333333333'
order by users.id
