select user_id::text as moderator_user_id,
       action,
       metadata ->> 'target_user_id' as target_user_id,
       metadata ->> 'new_role' as new_role,
       description,
       (metadata ->> 'error') <> '' as error_recorded
from moderation_audit_log
order by created_at, action;
