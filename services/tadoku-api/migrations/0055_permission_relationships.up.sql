begin;

set local lock_timeout = '5s';

create table permission_relationships (
  tenant text not null default current_setting('tadoku.tenant'),
  object_type text not null,
  object_id text not null,
  relation text not null,
  subject_type text not null,
  subject_id text not null,
  subject_relation text not null default '',
  created_at timestamptz not null default now(),

  primary key (tenant, object_type, object_id, relation, subject_type, subject_id, subject_relation),
  constraint permission_relationships_user_subject check (
    subject_type <> 'user'
    or (
      subject_relation = ''
      and subject_id ~ '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
    )
  )
);

alter table permission_relationships
  add constraint permission_relationships_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid;

alter table permission_relationships validate constraint permission_relationships_tenant_fkey;

alter table permission_relationships enable row level security;

create policy tenant_isolation on permission_relationships
  using (tenant = current_setting('tadoku.tenant'))
  with check (tenant = current_setting('tadoku.tenant'));

commit;
