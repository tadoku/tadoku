select event_type as kind, user_id::text, coalesce(contest_id::text, '') as reference,
  year::text, '' as rule_ids, '' as rates, '' as source, count(*)::text
from leaderboard_outbox
group by event_type, user_id, contest_id, year
union all
select 'log_provenance', user_id::text, coalesce(score_rule_set_id::text, ''),
  extract(year from created_at)::int::text, coalesce(array_to_string(score_rule_ids, ','), ''),
  coalesce(array_to_string(score_rates, ','), ''), coalesce(score_source, ''), ''
from logs where id='52fdfc07-2182-454f-963f-5f0f9a621d72'
union all
select 'contest_provenance', logs.user_id::text, contest_logs.contest_id::text,
  extract(year from logs.created_at)::int::text, coalesce(array_to_string(contest_logs.score_rule_ids, ','), ''),
  coalesce(array_to_string(contest_logs.score_rates, ','), ''), coalesce(contest_logs.score_source, ''), ''
from contest_logs inner join logs on logs.id=contest_logs.log_id
where contest_logs.log_id='52fdfc07-2182-454f-963f-5f0f9a621d72'
order by 1,3;
