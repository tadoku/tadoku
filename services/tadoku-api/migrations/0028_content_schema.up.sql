begin;

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

create table posts (
  id uuid primary key default uuid_generate_v4(),
  "namespace" varchar(50) not null,
  slug varchar(200) not null,
  current_content_id uuid not null,
  published_at timestamp,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  deleted_at timestamp
);

create unique index posts_slug on posts("namespace", slug);
create index posts_namespace on posts("namespace");

create table posts_content (
  id uuid primary key default uuid_generate_v4(),
  post_id uuid not null,
  title text not null,
  content text not null,
  created_at timestamp not null default now()
);

create index posts_content_post_id on posts_content(post_id);

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
