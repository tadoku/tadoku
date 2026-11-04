-- Global keys cannot be restored after test tenants hold duplicate ids.
-- This transaction fails atomically on duplicates; it never removes test data.
begin;

set local lock_timeout = '5s';

alter table log_tags drop constraint log_tags_log_fkey;
alter table scoring_rule_sets drop constraint scoring_rule_sets_contest_fkey;
alter table scoring_rule_sets drop constraint scoring_rule_sets_fallback_rule_set_fkey;
alter table scoring_rules drop constraint scoring_rules_rule_set_fkey;
alter table platform_scoring_config drop constraint platform_scoring_config_active_rule_set_fkey;
alter table contests drop constraint contests_scoring_rule_set_fkey;
alter table logs drop constraint logs_score_rule_set_fkey;
alter table contest_logs drop constraint contest_logs_score_rule_set_fkey;
alter table contest_events drop constraint contest_events_contest_fkey;

create unique index contests_tenant_id_idx on contests (tenant, id);
create unique index contest_registrations_tenant_id on contest_registrations (tenant, id);
create unique index logs_tenant_id on logs (tenant, id);
create unique index moderation_audit_log_tenant_id on moderation_audit_log (tenant, id);
create unique index scoring_rule_sets_tenant_id on scoring_rule_sets (tenant, id);
create unique index scoring_rules_tenant_id on scoring_rules (tenant, id);
create unique index pages_tenant_id on pages (tenant, id);
create unique index pages_content_tenant_id on pages_content (tenant, id);
create unique index posts_tenant_id on posts (tenant, id);
create unique index posts_content_tenant_id on posts_content (tenant, id);
create unique index announcements_tenant_id on announcements (tenant, id);

alter table contests drop constraint contests_pkey,
  add constraint contests_pkey primary key (id);
alter table contest_registrations drop constraint contest_registrations_pkey,
  add constraint contest_registrations_pkey primary key (id);
alter table logs drop constraint logs_pkey,
  add constraint logs_pkey primary key (id);
alter table moderation_audit_log drop constraint moderation_audit_log_pkey,
  add constraint moderation_audit_log_pkey primary key (id);
alter table scoring_rule_sets drop constraint scoring_rule_sets_pkey,
  add constraint scoring_rule_sets_pkey primary key (id);
alter table scoring_rules drop constraint scoring_rules_pkey,
  add constraint scoring_rules_pkey primary key (id);
alter table pages drop constraint pages_pkey,
  add constraint pages_pkey primary key (id);
alter table pages_content drop constraint pages_content_pkey,
  add constraint pages_content_pkey primary key (id);
alter table posts drop constraint posts_pkey,
  add constraint posts_pkey primary key (id);
alter table posts_content drop constraint posts_content_pkey,
  add constraint posts_content_pkey primary key (id);
alter table announcements drop constraint announcements_pkey,
  add constraint announcements_pkey primary key (id);

alter table log_tags add constraint log_tags_log_id_fkey
  foreign key (log_id) references logs (id) on delete cascade not valid;
alter table log_tags validate constraint log_tags_log_id_fkey;
alter table scoring_rule_sets add constraint scoring_rule_sets_contest_id_fkey
  foreign key (contest_id) references contests (id) not valid;
alter table scoring_rule_sets validate constraint scoring_rule_sets_contest_id_fkey;
alter table scoring_rule_sets add constraint scoring_rule_sets_fallback_rule_set_id_fkey
  foreign key (fallback_rule_set_id) references scoring_rule_sets (id) not valid;
alter table scoring_rule_sets validate constraint scoring_rule_sets_fallback_rule_set_id_fkey;
alter table scoring_rules add constraint scoring_rules_rule_set_id_fkey
  foreign key (rule_set_id) references scoring_rule_sets (id) on delete cascade not valid;
alter table scoring_rules validate constraint scoring_rules_rule_set_id_fkey;
alter table platform_scoring_config add constraint platform_scoring_config_active_rule_set_id_fkey
  foreign key (active_rule_set_id) references scoring_rule_sets (id) not valid;
alter table platform_scoring_config validate constraint platform_scoring_config_active_rule_set_id_fkey;
alter table contests add constraint contests_scoring_rule_set_id_fkey
  foreign key (scoring_rule_set_id) references scoring_rule_sets (id) not valid;
alter table contests validate constraint contests_scoring_rule_set_id_fkey;
alter table logs add constraint logs_score_rule_set_id_fkey
  foreign key (score_rule_set_id) references scoring_rule_sets (id) not valid;
alter table logs validate constraint logs_score_rule_set_id_fkey;
alter table contest_logs add constraint contest_logs_score_rule_set_id_fkey
  foreign key (score_rule_set_id) references scoring_rule_sets (id) not valid;
alter table contest_logs validate constraint contest_logs_score_rule_set_id_fkey;
alter table contest_events add constraint contest_events_contest_fkey
  foreign key (tenant, contest_id) references contests (tenant, id) on delete cascade not valid;
alter table contest_events validate constraint contest_events_contest_fkey;

commit;
