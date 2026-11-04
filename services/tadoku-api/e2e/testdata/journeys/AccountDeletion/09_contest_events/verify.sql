select contest_id, log_id, event, occurred_at
from contest_events
order by contest_id, log_id, event
