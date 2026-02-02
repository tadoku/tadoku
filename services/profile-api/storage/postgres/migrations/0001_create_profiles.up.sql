begin;

create extension if not exists "uuid-ossp";

create table profiles (
  id uuid primary key default uuid_generate_v4(),
  user_id varchar(255) not null unique,
  display_name varchar(200) not null,
  bio text,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  deleted_at timestamp
);

create index profiles_user_id on profiles(user_id);

commit;
