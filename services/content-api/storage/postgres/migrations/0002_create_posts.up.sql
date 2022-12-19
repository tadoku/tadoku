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

create table posts_content (
  id uuid primary key default uuid_generate_v4(),
  post_id uuid not null,
  title text not null,
  content text not null,
  created_at timestamp not null default now()
);

create index posts_content_post_id on posts_content(post_id);
