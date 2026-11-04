-- Restores the activity-prefixed unit keys and, in each tenant whose active
-- platform rule set is the one the up migration created, the previous
-- published version. That set is deleted when no score refers to it; otherwise
-- it stays as history with its renamed keys, which scoring_rules keeps
-- accepting. Comic, two-column and minute conversions cannot be split back:
-- those logs keep their tags and durations and keep scoring the same.
begin;

set local lock_timeout = '5s';

update platform_scoring_config
set active_rule_set_id = previous.id
from scoring_rule_sets as active, scoring_rule_sets as previous
where active.tenant = platform_scoring_config.tenant
  and active.id = platform_scoring_config.active_rule_set_id
  and active.id = md5('0073_reduce_log_units:' || active.tenant)::uuid
  and previous.tenant = active.tenant
  and previous.scope = 'platform'
  and previous.version = (
    select max(earlier.version)
    from scoring_rule_sets as earlier
    where earlier.tenant = active.tenant
      and earlier.scope = 'platform'
      and earlier.status = 'published'
      and earlier.version < active.version
  );

delete from scoring_rule_sets
where scoring_rule_sets.id = md5('0073_reduce_log_units:' || scoring_rule_sets.tenant)::uuid
  and not exists (
    select
    from platform_scoring_config
    where platform_scoring_config.tenant = scoring_rule_sets.tenant
      and platform_scoring_config.active_rule_set_id = scoring_rule_sets.id
  )
  and not exists (
    select
    from scoring_rule_sets as contest_sets
    where contest_sets.tenant = scoring_rule_sets.tenant
      and contest_sets.fallback_rule_set_id = scoring_rule_sets.id
  )
  and not exists (
    select
    from logs
    where logs.tenant = scoring_rule_sets.tenant
      and logs.score_rule_set_id = scoring_rule_sets.id
  )
  and not exists (
    select
    from contest_logs
    where contest_logs.tenant = scoring_rule_sets.tenant
      and contest_logs.score_rule_set_id = scoring_rule_sets.id
  );

update logs
set unit_key = case logs.log_activity_id
  when 1 then 'reading_'
  else 'writing_'
end || case logs.unit_key
  when 'pages' then 'page'
  when 'sentences' then 'sentence'
  else 'character'
end
where logs.unit_key in ('pages', 'sentences', 'characters');

update contest_logs
set unit_key = logs.unit_key
from logs
where logs.tenant = contest_logs.tenant
  and logs.id = contest_logs.log_id
  and contest_logs.unit_key in ('pages', 'sentences', 'characters');

update log_units
set unit_key = case log_units.log_activity_id
  when 1 then 'reading_'
  else 'writing_'
end || case log_units.unit_key
  when 'pages' then 'page'
  when 'sentences' then 'sentence'
  else 'character'
end
where log_units.unit_key in ('pages', 'sentences', 'characters');

update scoring_rules
set unit_key = case scoring_rules.activity_id
  when 1 then 'reading_'
  else 'writing_'
end || case scoring_rules.unit_key
  when 'pages' then 'page'
  when 'sentences' then 'sentence'
  else 'character'
end
where scoring_rules.unit_key in ('pages', 'sentences', 'characters')
  and scoring_rules.rule_set_id <> md5('0073_reduce_log_units:' || scoring_rules.tenant)::uuid;

commit;
