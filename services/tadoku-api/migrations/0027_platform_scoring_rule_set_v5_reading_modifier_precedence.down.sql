begin;

update platform_scoring_config
set active_rule_set_id = (
  select id
  from scoring_rule_sets
  where scope = 'platform'
    and version = 4
);

delete from scoring_rule_sets
where scope = 'platform'
  and version = 5;

commit;
