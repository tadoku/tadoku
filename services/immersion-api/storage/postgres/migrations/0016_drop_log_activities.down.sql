begin;

create table log_activities (
  id integer primary key not null,
  name varchar(100) not null,
  "default" boolean not null default false
);

insert into log_activities
  (id, name, "default")
values
  (1, 'Reading', true),
  (2, 'Listening', true),
  (3, 'Writing', false),
  (4, 'Speaking', false),
  (5, 'Study', false);

commit;
