begin;

alter table contest_logs
  drop constraint contest_logs_score_provenance_valid,
  drop constraint contest_logs_score_source_valid,
  drop column score_source,
  drop column score_rates,
  drop column score_rule_ids,
  drop column score_rule_set_id;

alter table logs
  drop constraint logs_score_provenance_valid,
  drop constraint logs_score_source_valid,
  drop column score_source,
  drop column score_rates,
  drop column score_rule_ids,
  drop column score_rule_set_id;

commit;
