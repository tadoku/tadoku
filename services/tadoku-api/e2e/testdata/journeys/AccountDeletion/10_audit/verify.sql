select user_id as actor_id, action, metadata, description, created_at as recorded_at
from moderation_audit_log
order by action, created_at
