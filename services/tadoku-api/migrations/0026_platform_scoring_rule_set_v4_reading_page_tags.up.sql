begin;

insert into scoring_rule_sets (
  scope,
  version,
  status,
  published_at
) values (
  'platform',
  4,
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
      and version = 4
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
    and version = 3
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
  rule_set.id,
  rule.priority,
  rule.stackable,
  rule.activity_id,
  rule.unit_key,
  rule.language_code,
  rule.tag,
  rule.score_source,
  rule.rate
from scoring_rule_sets as rule_set
cross join (
  values
    (80, false, 1, 'reading_page', 'jpn', 'two_column', 'amount', 1.6),
    (90, false, 1, 'reading_page', null, 'comic', 'amount', 0.2),
    (95, false, 1, 'reading_page', null, 'manga', 'amount', 0.2),
    (420, true, 4, null, null, 'dense', 'duration_minutes', 1.4)
) as rule (
  priority,
  stackable,
  activity_id,
  unit_key,
  language_code,
  tag,
  score_source,
  rate
)
where rule_set.scope = 'platform'
  and rule_set.version = 4;

update platform_scoring_config
set active_rule_set_id = (
  select id
  from scoring_rule_sets
  where scope = 'platform'
    and version = 4
);

commit;
