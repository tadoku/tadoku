select user_id::text as moderator_user_id,
       action,
       metadata ->> 'target_user_id' as target_user_id,
       metadata ->> 'flag_key' as flag_key,
       metadata ->> 'environment' as environment,
       metadata -> 'changed' as changed,
       metadata ->> 'resulting_revision' as resulting_revision,
       to_char(created_at, 'YYYY-MM-DD"T"HH24:MI:SS"Z"') as created_at
from moderation_audit_log
where action in ('feature_access_grant', 'feature_access_revoke')
order by created_at, action, (metadata -> 'changed')::text desc;
