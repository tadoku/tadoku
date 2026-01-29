begin;

-- create new log_tags table
create table log_tags (
  log_id uuid not null references logs(id) on delete cascade,
  user_id uuid not null,
  tag varchar(50) not null,
  primary key (log_id, tag)
);
create index idx_log_tags_user_tag on log_tags (user_id, tag);
create index idx_log_tags_user_prefix on log_tags (user_id, tag text_pattern_ops);

-- migrate existing tags
insert into log_tags (log_id, user_id, tag)
select l.id, l.user_id, lower(trim(unnest(l.tags)))
from logs l
where l.deleted_at is null and array_length(l.tags, 1) > 0
on conflict do nothing;

commit;
