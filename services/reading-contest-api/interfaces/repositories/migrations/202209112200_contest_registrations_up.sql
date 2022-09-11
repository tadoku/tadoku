create sequence contest_registrations_seq;

create table contest_registrations (
  id bigint check (id > 0) not null default nextval ('contest_registrations_seq'),
  contest_id bigint not null,
  user_id bigint not null,
  user_display_name varchar(255) default '' not null,
  language_codes varchar(3)[] not null,
  created_at timestamp not null,
  updated_at timestamp not null,
  primary key (id)
);

create index contest_registrations_contest_id on contest_registrations(contest_id);
create index contest_registrations_user_id on contest_registrations(user_id);
create unique index contest_registrations_unique_contest_user_language on contest_registrations(contest_id, user_id);

alter sequence contest_registrations_seq restart with 1;
