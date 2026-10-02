begin;

set local lock_timeout = '5s';

alter table account_deletion_requests alter column tenant set default current_setting('tadoku.tenant');

alter table account_deletion_requests enable row level security;

create policy tenant_isolation on account_deletion_requests using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table announcements alter column tenant set default current_setting('tadoku.tenant');

alter table announcements enable row level security;

create policy tenant_isolation on announcements using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table contest_logs alter column tenant set default current_setting('tadoku.tenant');

alter table contest_logs enable row level security;

create policy tenant_isolation on contest_logs using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table contest_registrations alter column tenant set default current_setting('tadoku.tenant');

alter table contest_registrations enable row level security;

create policy tenant_isolation on contest_registrations using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table contests alter column tenant set default current_setting('tadoku.tenant');

alter table contests enable row level security;

create policy tenant_isolation on contests using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table log_tags alter column tenant set default current_setting('tadoku.tenant');

alter table log_tags enable row level security;

create policy tenant_isolation on log_tags using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table logs alter column tenant set default current_setting('tadoku.tenant');

alter table logs enable row level security;

create policy tenant_isolation on logs using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table moderation_audit_log alter column tenant set default current_setting('tadoku.tenant');

alter table moderation_audit_log enable row level security;

create policy tenant_isolation on moderation_audit_log using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table pages alter column tenant set default current_setting('tadoku.tenant');

alter table pages enable row level security;

create policy tenant_isolation on pages using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table pages_content alter column tenant set default current_setting('tadoku.tenant');

alter table pages_content enable row level security;

create policy tenant_isolation on pages_content using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table platform_scoring_config alter column tenant set default current_setting('tadoku.tenant');

alter table platform_scoring_config enable row level security;

create policy tenant_isolation on platform_scoring_config using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table posts alter column tenant set default current_setting('tadoku.tenant');

alter table posts enable row level security;

create policy tenant_isolation on posts using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table posts_content alter column tenant set default current_setting('tadoku.tenant');

alter table posts_content enable row level security;

create policy tenant_isolation on posts_content using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table profiles alter column tenant set default current_setting('tadoku.tenant');

alter table profiles enable row level security;

create policy tenant_isolation on profiles using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table scoring_rule_sets alter column tenant set default current_setting('tadoku.tenant');

alter table scoring_rule_sets enable row level security;

create policy tenant_isolation on scoring_rule_sets using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table scoring_rules alter column tenant set default current_setting('tadoku.tenant');

alter table scoring_rules enable row level security;

create policy tenant_isolation on scoring_rules using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table user_roles alter column tenant set default current_setting('tadoku.tenant');

alter table user_roles enable row level security;

create policy tenant_isolation on user_roles using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table users alter column tenant set default current_setting('tadoku.tenant');

alter table users enable row level security;

create policy tenant_isolation on users using (tenant = current_setting('tadoku.tenant')) with check (tenant = current_setting('tadoku.tenant'));

alter table jobs alter column tenant set default current_setting('tadoku.tenant');

alter table jobs enable row level security;

create policy tenant_isolation on jobs using (tenant = current_setting('tadoku.tenant', true) or current_setting('tadoku.all_tenants', true) = 'on') with check (tenant = current_setting('tadoku.tenant', true) or current_setting('tadoku.all_tenants', true) = 'on');

alter table languages enable row level security;

create policy shared_read on languages for select using (true);

create policy production_insert on languages for insert with check (current_setting('tadoku.tenant', true) = 'tadoku/prod');

create policy production_update on languages for update using (current_setting('tadoku.tenant', true) = 'tadoku/prod');

create policy production_delete on languages for delete using (current_setting('tadoku.tenant', true) = 'tadoku/prod');

alter table log_units enable row level security;

create policy shared_read on log_units for select using (true);

create policy production_insert on log_units for insert with check (current_setting('tadoku.tenant', true) = 'tadoku/prod');

create policy production_update on log_units for update using (current_setting('tadoku.tenant', true) = 'tadoku/prod');

create policy production_delete on log_units for delete using (current_setting('tadoku.tenant', true) = 'tadoku/prod');

alter table tenants enable row level security;

create policy read_all on tenants for select using (true);

create policy test_insert on tenants for insert with check (kind = 'test');

create policy test_update on tenants for update using (kind = 'test') with check (kind = 'test');

create policy test_delete on tenants for delete using (kind = 'test');

alter table tenant_overrides enable row level security;

create policy read_all on tenant_overrides for select using (true);

create policy test_write on tenant_overrides for all using (tenant in (select key from tenants where kind = 'test')) with check (tenant in (select key from tenants where kind = 'test'));

commit;
