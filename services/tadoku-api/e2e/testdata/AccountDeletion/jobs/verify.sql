select
  task_type,
  state,
  payload - 'email_id' as payload,
  payload ? 'email_id' as names_email
from jobs
where tenant = 'tadoku/prod'
order by task_type, payload::text
