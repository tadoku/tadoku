-- Validates the foreign keys migration 0070 added. Validation scans each
-- referencing table under a lock that permits reads and writes. It fails if a
-- row references an id that exists only in another tenant. Repair such rows
-- in a reviewed change before retrying.
begin;

set local lock_timeout = '5s';

alter table log_tags validate constraint log_tags_log_fkey;
alter table scoring_rule_sets validate constraint scoring_rule_sets_contest_fkey;
alter table scoring_rule_sets validate constraint scoring_rule_sets_fallback_rule_set_fkey;
alter table scoring_rules validate constraint scoring_rules_rule_set_fkey;
alter table platform_scoring_config validate constraint platform_scoring_config_active_rule_set_fkey;
alter table contests validate constraint contests_scoring_rule_set_fkey;
alter table logs validate constraint logs_score_rule_set_fkey;
alter table contest_logs validate constraint contest_logs_score_rule_set_fkey;

commit;
