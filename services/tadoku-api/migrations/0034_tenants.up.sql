begin;

set local lock_timeout = '5s';

create table tenants (
  key text primary key,
  kind text not null,
  constraint tenants_key_valid
    check (key ~ '^[a-z0-9][a-z0-9-]{0,55}/[a-z0-9][a-z0-9-]{0,55}$'),
  constraint tenants_kind_valid check (kind in ('production', 'test')),
  constraint tenants_single_production
    check ((kind = 'production') = (key = 'tadoku/prod'))
);

insert into tenants (key, kind) values ('tadoku/prod', 'production');

create table tenant_overrides (
  tenant text not null references tenants (key) on delete cascade,
  component text not null,
  primary key (tenant, component),
  constraint tenant_overrides_not_production check (tenant <> 'tadoku/prod'),
  constraint tenant_overrides_component_valid check (component ~ '^[a-z][a-z0-9-]*$')
);

alter table account_deletion_requests
  add column tenant text not null default 'tadoku/prod',
  add constraint account_deletion_requests_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table announcements
  add column tenant text not null default 'tadoku/prod',
  add constraint announcements_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table contest_logs
  add column tenant text not null default 'tadoku/prod',
  add constraint contest_logs_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table contest_registrations
  add column tenant text not null default 'tadoku/prod',
  add constraint contest_registrations_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table contests
  add column tenant text not null default 'tadoku/prod',
  add constraint contests_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table jobs
  add column tenant text not null default 'tadoku/prod',
  add constraint jobs_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table log_tags
  add column tenant text not null default 'tadoku/prod',
  add constraint log_tags_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table logs
  add column tenant text not null default 'tadoku/prod',
  add constraint logs_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table moderation_audit_log
  add column tenant text not null default 'tadoku/prod',
  add constraint moderation_audit_log_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table pages
  add column tenant text not null default 'tadoku/prod',
  add constraint pages_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table pages_content
  add column tenant text not null default 'tadoku/prod',
  add constraint pages_content_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table platform_scoring_config
  add column tenant text not null default 'tadoku/prod',
  add constraint platform_scoring_config_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table posts
  add column tenant text not null default 'tadoku/prod',
  add constraint posts_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table posts_content
  add column tenant text not null default 'tadoku/prod',
  add constraint posts_content_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table profiles
  add column tenant text not null default 'tadoku/prod',
  add constraint profiles_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table scoring_rule_sets
  add column tenant text not null default 'tadoku/prod',
  add constraint scoring_rule_sets_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table scoring_rules
  add column tenant text not null default 'tadoku/prod',
  add constraint scoring_rules_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table user_roles
  add column tenant text not null default 'tadoku/prod',
  add constraint user_roles_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table users
  add column tenant text not null default 'tadoku/prod',
  add constraint users_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

commit;
