begin;

alter table contests drop column scoring_rule_set_id;

drop table platform_scoring_config;

drop table scoring_rules;
drop table scoring_rule_sets;

commit;
