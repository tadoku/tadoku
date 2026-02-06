begin;

create table announcements (
  id uuid primary key default uuid_generate_v4(),
  "namespace" varchar(50) not null,
  title text not null,
  content text not null,
  style varchar(20) not null default 'info',
  href text,
  starts_at timestamp not null,
  ends_at timestamp not null,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  deleted_at timestamp
);

create index announcements_namespace on announcements("namespace");
create index announcements_active on announcements("namespace", starts_at, ends_at) where deleted_at is null;

commit;
