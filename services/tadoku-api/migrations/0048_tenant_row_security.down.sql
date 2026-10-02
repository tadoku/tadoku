begin;

set local lock_timeout = '5s';

drop policy tenant_isolation on account_deletion_requests;

alter table account_deletion_requests disable row level security;

alter table account_deletion_requests alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on announcements;

alter table announcements disable row level security;

alter table announcements alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on contest_logs;

alter table contest_logs disable row level security;

alter table contest_logs alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on contest_registrations;

alter table contest_registrations disable row level security;

alter table contest_registrations alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on contests;

alter table contests disable row level security;

alter table contests alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on log_tags;

alter table log_tags disable row level security;

alter table log_tags alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on logs;

alter table logs disable row level security;

alter table logs alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on moderation_audit_log;

alter table moderation_audit_log disable row level security;

alter table moderation_audit_log alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on pages;

alter table pages disable row level security;

alter table pages alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on pages_content;

alter table pages_content disable row level security;

alter table pages_content alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on platform_scoring_config;

alter table platform_scoring_config disable row level security;

alter table platform_scoring_config alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on posts;

alter table posts disable row level security;

alter table posts alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on posts_content;

alter table posts_content disable row level security;

alter table posts_content alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on profiles;

alter table profiles disable row level security;

alter table profiles alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on scoring_rule_sets;

alter table scoring_rule_sets disable row level security;

alter table scoring_rule_sets alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on scoring_rules;

alter table scoring_rules disable row level security;

alter table scoring_rules alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on user_roles;

alter table user_roles disable row level security;

alter table user_roles alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on users;

alter table users disable row level security;

alter table users alter column tenant set default 'tadoku/prod';

drop policy tenant_isolation on jobs;

alter table jobs disable row level security;

alter table jobs alter column tenant set default 'tadoku/prod';

drop policy shared_read on languages;

drop policy production_insert on languages;

drop policy production_update on languages;

drop policy production_delete on languages;

alter table languages disable row level security;

drop policy shared_read on log_units;

drop policy production_insert on log_units;

drop policy production_update on log_units;

drop policy production_delete on log_units;

alter table log_units disable row level security;

drop policy read_all on tenants;

drop policy test_insert on tenants;

drop policy test_update on tenants;

drop policy test_delete on tenants;

alter table tenants disable row level security;

drop policy read_all on tenant_overrides;

drop policy test_write on tenant_overrides;

alter table tenant_overrides disable row level security;

commit;
