-- Global keys cannot be restored after test tenants hold duplicate values.
-- This transaction fails atomically on duplicates; it never removes test data.
begin;

set local lock_timeout = '5s';

create unique index contest_registrations_user_id_contest_id
  on contest_registrations(user_id, contest_id);
create unique index contest_logs_contest_id on contest_logs(contest_id, log_id);
create unique index scoring_rule_sets_platform_version
  on scoring_rule_sets(version) where scope = 'platform';
create unique index scoring_rule_sets_contest_version
  on scoring_rule_sets(contest_id, version) where scope = 'contest';
create unique index scoring_rules_rule_set_priority on scoring_rules(rule_set_id, priority);
create unique index pages_slug on pages("namespace", slug);
create unique index posts_slug on posts("namespace", slug);
alter table account_deletion_requests
  add constraint account_deletion_requests_identity_id_key unique (identity_id);

create unique index user_roles_tenant_user_id on user_roles(tenant, user_id);
create unique index users_tenant_id on users(tenant, id);
create unique index profiles_tenant_user_id on profiles(tenant, user_id);
create unique index log_tags_tenant_log_id_tag on log_tags(tenant, log_id, tag);
create unique index platform_scoring_config_tenant_singleton
  on platform_scoring_config(tenant, singleton);

alter table user_roles drop constraint user_roles_user_fkey;
alter table user_roles drop constraint user_roles_pkey,
  add constraint user_roles_pkey primary key (user_id);
alter table users drop constraint users_pkey,
  add constraint users_pkey primary key (id);
alter table user_roles add constraint user_roles_user_id_fkey
  foreign key (user_id) references users(id) not valid;
alter table user_roles validate constraint user_roles_user_id_fkey;
alter table profiles drop constraint profiles_pkey,
  add constraint profiles_pkey primary key (user_id);
alter table log_tags drop constraint log_tags_pkey1,
  add constraint log_tags_pkey1 primary key (log_id, tag);
alter table platform_scoring_config drop constraint platform_scoring_config_pkey,
  add constraint platform_scoring_config_pkey primary key (singleton);

commit;
