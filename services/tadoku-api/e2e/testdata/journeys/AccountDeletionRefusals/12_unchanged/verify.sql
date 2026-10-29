select
  users.display_name,
  users.deletion_locked_at,
  (select count(*) from sessions
    where sessions.tenant = users.tenant and sessions.user_id = users.id) as sessions,
  (select count(*) from jobs where jobs.tenant = users.tenant) as jobs,
  (select count(*) from moderation_audit_log as audit where audit.tenant = users.tenant) as audit_events
from users
where users.tenant = 'tadoku/prod'
  and users.id in (
    '11111111-1111-4111-8111-111111111111',
    '22222222-2222-4222-8222-222222222222',
    '33333333-3333-4333-8333-333333333333'
  )
order by users.id
