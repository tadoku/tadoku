select action
from moderation_audit_log
where action in ('feature_access_grant', 'feature_access_revoke')
order by action;
