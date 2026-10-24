begin;

set local lock_timeout = '5s';

create table contest_events (
  tenant text not null default current_setting('tadoku.tenant'),
  id uuid not null,
  contest_id uuid not null,
  log_id uuid not null,
  event text not null,
  occurred_at timestamptz not null default current_timestamp,

  primary key (tenant, id),
  unique (tenant, contest_id, occurred_at, id),
  constraint contest_events_event_valid check (
    event in ('log.submitted', 'log.updated', 'log.withdrawn')
  )
);

alter table contest_events
  add constraint contest_events_tenant_fkey foreign key (tenant)
    references tenants (key) on delete cascade not valid,
  add constraint contest_events_contest_fkey foreign key (tenant, contest_id)
    references contests (tenant, id) on delete cascade not valid;

alter table contest_events validate constraint contest_events_tenant_fkey;
alter table contest_events validate constraint contest_events_contest_fkey;

alter table contest_events enable row level security;

create policy tenant_isolation on contest_events
  using (tenant = current_setting('tadoku.tenant'))
  with check (tenant = current_setting('tadoku.tenant'));

commit;
