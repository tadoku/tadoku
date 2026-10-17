select action,
       metadata ->> 'target_user_id' as target_user_id,
       metadata ->> 'flag_key' as flag_key,
       metadata ->> 'environment' as environment,
       (metadata ->> 'error') <> '' as error_recorded
from moderation_audit_log
order by created_at, action;
