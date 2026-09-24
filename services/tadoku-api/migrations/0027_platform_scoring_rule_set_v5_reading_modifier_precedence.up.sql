begin;

insert into scoring_rule_sets (
  scope,
  version,
  status,
  published_at
) values (
  'platform',
  5,
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
      and version = 5
  ),
  case
    when activity_id = 1
      and unit_key = 'reading_page'
      and tag = 'comic'
      then 70
    when activity_id = 1
      and unit_key = 'reading_page'
      and tag = 'manga'
      then 75
    else priority
  end,
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
    and version = 4
);

update platform_scoring_config
set active_rule_set_id = (
  select id
  from scoring_rule_sets
  where scope = 'platform'
    and version = 5
);

commit;
