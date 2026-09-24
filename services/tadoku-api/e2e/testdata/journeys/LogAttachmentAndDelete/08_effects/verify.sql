select 'log' as kind, count(*)::text as count from logs where id='52fdfc07-2182-454f-963f-5f0f9a621d72' and deleted_at is not null
union all
select 'legacy_outbox', count(*)::text from leaderboard_outbox
union all
select task_type, count(*)::text from async_outbox group by task_type order by 1;
