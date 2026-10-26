select user_id as actor_id, action, metadata, description, created_at as recorded_at
from moderation_audit_log
where tenant = 'tadoku/prod'
order by action, created_at
