begin;

create table log_tags (
  log_id uuid not null references logs(id) on delete cascade,
  user_id uuid not null,
  tag varchar(50) not null,
  created_at timestamp not null default now(),
  primary key (log_id, tag)
);

create index log_tags_tag on log_tags(tag);
create index log_tags_user_id_tag on log_tags(user_id, tag);

commit;
