begin;

set local lock_timeout = '5s';

drop index contest_registrations_user_id_contest_id;
drop index contest_logs_contest_id;
drop index scoring_rule_sets_platform_version;
drop index scoring_rule_sets_contest_version;
drop index scoring_rules_rule_set_priority;
drop index pages_slug;
drop index posts_slug;
alter table account_deletion_requests drop constraint account_deletion_requests_identity_id_key;

alter table user_roles drop constraint user_roles_user_id_fkey;
alter table user_roles drop constraint user_roles_pkey,
  add constraint user_roles_pkey primary key using index user_roles_tenant_user_id;
alter table users drop constraint users_pkey,
  add constraint users_pkey primary key using index users_tenant_id;
alter table user_roles add constraint user_roles_user_fkey
  foreign key (tenant, user_id) references users (tenant, id) not valid;
alter table user_roles validate constraint user_roles_user_fkey;
alter table profiles drop constraint profiles_pkey,
  add constraint profiles_pkey primary key using index profiles_tenant_user_id;
alter table log_tags drop constraint log_tags_pkey1,
  add constraint log_tags_pkey1 primary key using index log_tags_tenant_log_id_tag;
alter table platform_scoring_config drop constraint platform_scoring_config_pkey,
  add constraint platform_scoring_config_pkey primary key using index platform_scoring_config_tenant_singleton;

commit;
