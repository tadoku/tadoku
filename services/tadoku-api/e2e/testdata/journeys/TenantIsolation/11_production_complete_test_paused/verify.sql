select tenant, task_type, state
from jobs
order by tenant, task_type, state;
