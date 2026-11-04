-- Makes (tenant, id) the primary key of every tenant-owned table keyed by id
-- alone, using the indexes built concurrently by migrations 0053 and 0059 to
-- 0068, and makes every foreign key to those tables include tenant. The
-- foreign keys are added not valid; the next migration validates them
-- without holding this transaction's locks.
begin;

set local lock_timeout = '5s';

alter table log_tags drop constraint log_tags_log_id_fkey;
alter table scoring_rule_sets drop constraint scoring_rule_sets_contest_id_fkey;
alter table scoring_rule_sets drop constraint scoring_rule_sets_fallback_rule_set_id_fkey;
alter table scoring_rules drop constraint scoring_rules_rule_set_id_fkey;
alter table platform_scoring_config drop constraint platform_scoring_config_active_rule_set_id_fkey;
alter table contests drop constraint contests_scoring_rule_set_id_fkey;
alter table logs drop constraint logs_score_rule_set_id_fkey;
alter table contest_logs drop constraint contest_logs_score_rule_set_id_fkey;

alter table contests drop constraint contests_pkey,
  add constraint contests_pkey primary key using index contests_tenant_id_idx;
alter table contest_registrations drop constraint contest_registrations_pkey,
  add constraint contest_registrations_pkey primary key using index contest_registrations_tenant_id;
alter table logs drop constraint logs_pkey,
  add constraint logs_pkey primary key using index logs_tenant_id;
alter table moderation_audit_log drop constraint moderation_audit_log_pkey,
  add constraint moderation_audit_log_pkey primary key using index moderation_audit_log_tenant_id;
alter table scoring_rule_sets drop constraint scoring_rule_sets_pkey,
  add constraint scoring_rule_sets_pkey primary key using index scoring_rule_sets_tenant_id;
alter table scoring_rules drop constraint scoring_rules_pkey,
  add constraint scoring_rules_pkey primary key using index scoring_rules_tenant_id;
alter table pages drop constraint pages_pkey,
  add constraint pages_pkey primary key using index pages_tenant_id;
alter table pages_content drop constraint pages_content_pkey,
  add constraint pages_content_pkey primary key using index pages_content_tenant_id;
alter table posts drop constraint posts_pkey,
  add constraint posts_pkey primary key using index posts_tenant_id;
alter table posts_content drop constraint posts_content_pkey,
  add constraint posts_content_pkey primary key using index posts_content_tenant_id;
alter table announcements drop constraint announcements_pkey,
  add constraint announcements_pkey primary key using index announcements_tenant_id;

alter table log_tags add constraint log_tags_log_fkey
  foreign key (tenant, log_id) references logs (tenant, id) on delete cascade not valid;
alter table scoring_rule_sets add constraint scoring_rule_sets_contest_fkey
  foreign key (tenant, contest_id) references contests (tenant, id) not valid;
alter table scoring_rule_sets add constraint scoring_rule_sets_fallback_rule_set_fkey
  foreign key (tenant, fallback_rule_set_id) references scoring_rule_sets (tenant, id) not valid;
alter table scoring_rules add constraint scoring_rules_rule_set_fkey
  foreign key (tenant, rule_set_id) references scoring_rule_sets (tenant, id) on delete cascade not valid;
alter table platform_scoring_config add constraint platform_scoring_config_active_rule_set_fkey
  foreign key (tenant, active_rule_set_id) references scoring_rule_sets (tenant, id) not valid;
alter table contests add constraint contests_scoring_rule_set_fkey
  foreign key (tenant, scoring_rule_set_id) references scoring_rule_sets (tenant, id) not valid;
alter table logs add constraint logs_score_rule_set_fkey
  foreign key (tenant, score_rule_set_id) references scoring_rule_sets (tenant, id) not valid;
alter table contest_logs add constraint contest_logs_score_rule_set_fkey
  foreign key (tenant, score_rule_set_id) references scoring_rule_sets (tenant, id) not valid;

commit;
