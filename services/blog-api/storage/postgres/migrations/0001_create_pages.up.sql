create table pages (
  id uuid primary key default uuid_generate_v4(),
  slug varchar(200) not null,
  title text not null,
  html text not null,
  published_at timestamp,
  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  deleted_at timestamp
);

create unique index pages_slug on pages(slug);