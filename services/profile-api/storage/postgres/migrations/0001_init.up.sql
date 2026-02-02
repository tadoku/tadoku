begin;

create extension if not exists "uuid-ossp";

create table profiles (
  user_id uuid primary key not null,

  created_at timestamp not null default now(),
  updated_at timestamp not null default now()
);

commit;
