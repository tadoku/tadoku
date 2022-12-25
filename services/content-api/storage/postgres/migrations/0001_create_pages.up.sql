begin;

create extension if not exists "uuid-ossp";

create table pages (
  id uuid primary key default uuid_generate_v4(),
  "namespace" varchar(50) not null,
  slug varchar(200) not null,
  current_content_id uuid not null,
  published_at timestamp,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  deleted_at timestamp
);

create unique index pages_slug on pages("namespace", slug);
create index pages_namespace on pages("namespace");

create table pages_content (
  id uuid primary key default uuid_generate_v4(),
  page_id uuid not null,
  title text not null,
  html text not null,
  created_at timestamp not null default now()
);

create index pages_content_page_id on pages_content(page_id);

commit;