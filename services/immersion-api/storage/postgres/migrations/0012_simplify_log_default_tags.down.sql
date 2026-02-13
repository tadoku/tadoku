-- Cannot restore activity associations after they've been dropped
alter table log_default_tags add column log_activity_id smallint not null default 1;
create index log_tags_log_activity_id on log_default_tags (log_activity_id);
