begin;

set local lock_timeout = '5s';

alter table users drop column tenant;
alter table user_roles drop column tenant;
alter table scoring_rules drop column tenant;
alter table scoring_rule_sets drop column tenant;
alter table profiles drop column tenant;
alter table posts_content drop column tenant;
alter table posts drop column tenant;
alter table platform_scoring_config drop column tenant;
alter table pages_content drop column tenant;
alter table pages drop column tenant;
alter table moderation_audit_log drop column tenant;
alter table logs drop column tenant;
alter table log_tags drop column tenant;
alter table jobs drop column tenant;
alter table contests drop column tenant;
alter table contest_registrations drop column tenant;
alter table contest_logs drop column tenant;
alter table announcements drop column tenant;
alter table account_deletion_requests drop column tenant;

drop table tenant_overrides;
drop table tenants;

commit;
