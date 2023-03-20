begin;

create table users (
  id uuid primary key not null,
  display_name varchar(255) not null,

  created_at timestamp not null default now(),
  updated_at timestamp not null default now()
);

commit;