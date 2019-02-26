create sequence ranking_seq;

create table rankings (
  id bigint check (id > 0) not null default nextval ('ranking_seq'),
  contest_id bigint not null,
  user_id bigint not null,
  language_code varchar(3),
  amount float(3) not null,
  created_at timestamp not null,
  updated_at timestamp not null,
  primary key (id)
);

create index rankings_contest_id on rankings(contest_id);
create index rankings_user_id on rankings(user_id);
create index rankings_language_code on rankings(language_code);

alter sequence ranking_seq restart with 1;
