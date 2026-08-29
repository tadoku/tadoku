begin;

alter table logs drop column frozen_at;
alter table users drop column deleted_at;
alter table users drop column deletion_locked_at;

commit;
