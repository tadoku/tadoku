begin;

set local lock_timeout = '5s';

create table accounts (
  tenant text not null default current_setting('tadoku.tenant'),
  user_id uuid not null,
  email text not null,
  email_normalized text not null,
  password_hash text not null,
  state text not null default 'active',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),

  primary key (tenant, user_id),
  unique (tenant, email_normalized),
  constraint accounts_state_valid check (state in ('active', 'inactive'))
);

create table sessions (
  tenant text not null default current_setting('tadoku.tenant'),
  token_hash text not null,
  user_id uuid not null,
  authenticated_at timestamptz not null,
  expires_at timestamptz not null,
  created_at timestamptz not null default now(),

  primary key (tenant, token_hash)
);

create table recovery_tokens (
  tenant text not null default current_setting('tadoku.tenant'),
  token_hash text not null,
  user_id uuid not null,
  expires_at timestamptz not null,
  used_at timestamptz,

  primary key (tenant, token_hash)
);

create table emails (
  tenant text not null default current_setting('tadoku.tenant'),
  id uuid not null default uuid_generate_v4(),
  template text not null,
  recipient text not null,
  subject text not null,
  body text not null,
  params jsonb not null,
  status text not null default 'pending',
  sent_at timestamptz,
  provider_message_id text,
  error text,
  created_at timestamptz not null default now(),

  primary key (tenant, id),
  constraint emails_status_valid check (status in ('pending', 'sent', 'held', 'failed'))
);

create table access_tokens (
  tenant text not null default current_setting('tadoku.tenant'),
  id uuid not null default uuid_generate_v4(),
  name text not null,
  token_hash text not null,
  scopes text[] not null,
  created_at timestamptz not null default now(),
  expires_at timestamptz not null,
  last_used_at timestamptz,
  revoked_at timestamptz,

  primary key (tenant, id),
  unique (tenant, token_hash)
);

alter table accounts
  add constraint accounts_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid,
  add constraint accounts_user_fkey foreign key (tenant, user_id)
    references users (tenant, id) on delete cascade not valid;

alter table sessions
  add constraint sessions_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid,
  add constraint sessions_user_fkey foreign key (tenant, user_id)
    references users (tenant, id) on delete cascade not valid;

alter table recovery_tokens
  add constraint recovery_tokens_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid,
  add constraint recovery_tokens_user_fkey foreign key (tenant, user_id)
    references users (tenant, id) on delete cascade not valid;

alter table emails
  add constraint emails_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table access_tokens
  add constraint access_tokens_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table accounts validate constraint accounts_tenant_fkey;
alter table accounts validate constraint accounts_user_fkey;
alter table sessions validate constraint sessions_tenant_fkey;
alter table sessions validate constraint sessions_user_fkey;
alter table recovery_tokens validate constraint recovery_tokens_tenant_fkey;
alter table recovery_tokens validate constraint recovery_tokens_user_fkey;
alter table emails validate constraint emails_tenant_fkey;
alter table access_tokens validate constraint access_tokens_tenant_fkey;

alter table accounts enable row level security;
alter table sessions enable row level security;
alter table recovery_tokens enable row level security;
alter table emails enable row level security;
alter table access_tokens enable row level security;

create policy tenant_isolation on accounts
  using (tenant = current_setting('tadoku.tenant'))
  with check (tenant = current_setting('tadoku.tenant'));

create policy tenant_isolation on sessions
  using (tenant = current_setting('tadoku.tenant'))
  with check (tenant = current_setting('tadoku.tenant'));

create policy tenant_isolation on recovery_tokens
  using (tenant = current_setting('tadoku.tenant'))
  with check (tenant = current_setting('tadoku.tenant'));

create policy tenant_isolation on emails
  using (tenant = current_setting('tadoku.tenant'))
  with check (tenant = current_setting('tadoku.tenant'));

create policy tenant_isolation on access_tokens
  using (tenant = current_setting('tadoku.tenant'))
  with check (tenant = current_setting('tadoku.tenant'));

commit;
