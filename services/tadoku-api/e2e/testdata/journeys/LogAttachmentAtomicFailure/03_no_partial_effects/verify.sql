select 'attachment' as kind, contest_id::text as value from contest_logs where log_id='d1000000-0000-4000-8000-000000000001'
union all
select 'outbox', count(*)::text from leaderboard_outbox order by 1,2;
