create sequence contest_seq;
create sequence contest_log_seq;
create sequence medium_seq;

create table contests (
  id bigint check (id > 0) not null default nextval ('contest_seq'),
  start date not null,
  "end" date not null,
  open boolean not null default false,
  primary key (id)
);

create table mediums (
  id smallint check (id > 0) not null default nextval ('medium_seq'),
  description text not null,
  points float(2) not null,
  primary key (id)
);

create table languages (
  iso_code varchar(3) not null,
  name varchar(50) not null,
  primary key (iso_code)
);

create table contest_logs (
  id bigint check (id > 0) not null default nextval ('contest_log_seq'),
  contest_id bigint not null,
  user_id bigint not null,
  language_code varchar(3) not null,
  medium_id smallint not null,
  amount float(3) not null,
  created_at timestamp not null,
  updated_at timestamp not null,
  deleted_at timestamp default null,
  primary key (id)
);

create index content_logs_contest_id on contest_logs(contest_id);
create index content_logs_user_id on contest_logs(user_id);
create index content_logs_language_code on contest_logs(language_code);
create index content_logs_medium_id on contest_logs(medium_id);

alter sequence contest_seq restart with 1;
alter sequence medium_seq restart with 1;
alter sequence contest_log_seq restart with 1;
