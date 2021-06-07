create sequence user_seq;

create table users (
  id bigint check (id > 0) not null default nextval ('user_seq'),
  email varchar(255) not null unique,
  display_name varchar(255) not null,
  password bytea not null,
  role smallint not null default 0,
  preferences jsonb not null default '{}'::jsonb,
  primary key (id)
);

alter sequence user_seq restart with 1;
