begin;

alter table logs
  add column score_rule_set_id uuid references scoring_rule_sets(id),
  add column score_rule_ids uuid[],
  add column score_rates real[],
  add column score_source text;

alter table logs
  add constraint logs_score_source_valid
  check (
    score_source is null
    or score_source in ('amount', 'duration_minutes')
  ),
  add constraint logs_score_provenance_valid
  check (
    (
      score_rule_set_id is null
      and score_rule_ids is null
      and score_rates is null
    )
    or (
      score_rule_set_id is not null
      and score_rule_ids is not null
      and score_rates is not null
      and cardinality(score_rule_ids) > 0
      and cardinality(score_rule_ids) = cardinality(score_rates)
      and score_source is not null
    )
  );

alter table contest_logs
  add column score_rule_set_id uuid references scoring_rule_sets(id),
  add column score_rule_ids uuid[],
  add column score_rates real[],
  add column score_source text;

alter table contest_logs
  add constraint contest_logs_score_source_valid
  check (
    score_source is null
    or score_source in ('amount', 'duration_minutes')
  ),
  add constraint contest_logs_score_provenance_valid
  check (
    (
      score_rule_set_id is null
      and score_rule_ids is null
      and score_rates is null
    )
    or (
      score_rule_set_id is not null
      and score_rule_ids is not null
      and score_rates is not null
      and cardinality(score_rule_ids) > 0
      and cardinality(score_rule_ids) = cardinality(score_rates)
      and score_source is not null
    )
  );

commit;
