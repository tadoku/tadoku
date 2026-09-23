begin;

insert into scoring_rule_sets (
  scope,
  version,
  status,
  published_at
) values (
  'platform',
  3,
  'published',
  now()
);

insert into scoring_rules (
  rule_set_id,
  priority,
  stackable,
  activity_id,
  unit_key,
  language_code,
  tag,
  score_source,
  rate
) select
  (
    select id
    from scoring_rule_sets
    where scope = 'platform'
      and version = 3
  ),
  priority,
  stackable,
  activity_id,
  unit_key,
  language_code,
  tag,
  score_source,
  rate
from scoring_rules
where rule_set_id = (
  select id
  from scoring_rule_sets
  where scope = 'platform'
    and version = 2
);

insert into scoring_rules (
  rule_set_id,
  priority,
  stackable,
  activity_id,
  tag,
  score_source,
  rate
) select
  id,
  220,
  true,
  2,
  'dense',
  'duration_minutes',
  1.5
from scoring_rule_sets
where scope = 'platform'
  and version = 3;

update platform_scoring_config
set active_rule_set_id = (
  select id
  from scoring_rule_sets
  where scope = 'platform'
    and version = 3
);

commit;
