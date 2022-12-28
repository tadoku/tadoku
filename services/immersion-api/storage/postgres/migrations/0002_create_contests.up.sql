begin;

create table contests (
  id uuid primary key default uuid_generate_v4(),

  -- contest owner
  owner_user_id uuid not null,
  owner_user_display_name varchar(255) not null,
  "private" boolean not null,

  -- contest info
  contest_start date not null,
  contest_end date not null,
  registration_start date not null,
  registration_end date not null,

  "description" varchar(255) not null,
  language_code_allow_list varchar(10)[],
  activity_type_id_allow_list integer[],
  official boolean not null,

  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  deleted_at timestamp default null
);

create table contest_registrations (
  id uuid primary key default uuid_generate_v4(),
  contest_id uuid not null,

  user_id uuid not null,
  user_display_name varchar(255) not null,
  language_codes varchar(10)[] not null,

  created_at timestamp not null default now(),
  updated_at timestamp not null default now(),
  deleted_at timestamp default null
);

commit;