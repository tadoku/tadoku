select 'contest_events' as kind, count(*)::text as value from contest_events
union all
select 'jobs', count(*)::text from jobs order by 1,2;
