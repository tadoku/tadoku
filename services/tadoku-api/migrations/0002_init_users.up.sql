begin;

create table users (
  id uuid primary key not null,
  display_name varchar(255) not null,

  created_at timestamp not null default now(),
  updated_at timestamp not null default now()
);

-- backfill data based on existing registrations
insert into users (id, display_name)
(
  select
    user_id,
    (select user_display_name from contest_registrations where user_id = r.user_id order by created_at desc limit 1)
  from contest_registrations as r
  group by user_id
);

commit;