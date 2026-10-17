select user_id::text as moderator_user_id,
       action,
       metadata ->> 'target_user_id' as target_user_id,
       metadata ->> 'new_role' as new_role,
       description,
       to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
from moderation_audit_log
order by created_at, action;
