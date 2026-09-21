select 'contest_log' as kind, logs.user_id::text as user_id, contest_logs.log_id::text as reference
from contest_logs
inner join logs on logs.id = contest_logs.log_id
where contest_logs.contest_id = 'f1111111-1111-4111-8111-111111111111'
union all
select 'outbox' as kind, user_id::text as user_id,
  event_type || ':' || coalesce(contest_id::text, year::text) as reference
from leaderboard_outbox
order by kind, user_id, reference;
