select 'log' as kind, count(*)::text as count from logs where id='52fdfc07-2182-454f-963f-5f0f9a621d72' and deleted_at is not null
union all
select event_type, count(*)::text from leaderboard_outbox where user_id='11111111-1111-4111-8111-111111111111' group by event_type order by 1;
