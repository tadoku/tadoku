select contest_id, log_id, event, occurred_at
from contest_events
where tenant = 'tadoku/prod'
order by contest_id, log_id, event
