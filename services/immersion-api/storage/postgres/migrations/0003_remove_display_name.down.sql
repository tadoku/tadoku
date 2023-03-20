begin;

alter table contest_registrations add column user_display_name varchar(255) not null;

commit;