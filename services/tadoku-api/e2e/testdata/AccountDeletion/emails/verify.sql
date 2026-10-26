select template, recipient, subject, body, params, status, created_at
from emails
where tenant = 'tadoku/prod'
order by created_at
