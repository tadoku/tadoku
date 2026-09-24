select 'audit' as kind, user_id::text as actor, action as value, count(*)::text as count from moderation_audit_log where action='detach_log' group by user_id, action
union all
select 'legacy_outbox', '' as actor, '' as value, count(*)::text from leaderboard_outbox
union all
select task_type, '' as actor, coalesce(payload->>'contest_id', payload->>'year', ''), count(*)::text
from async_outbox group by task_type, payload
union all
select 'eligible', user_id::text, eligible_official_leaderboard::text, '1' from logs where id='d1000000-0000-4000-8000-000000000001'
order by 1,3;
