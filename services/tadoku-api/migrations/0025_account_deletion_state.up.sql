begin;

alter table users add column deletion_locked_at timestamp;
alter table users add column deleted_at timestamp;
alter table logs add column frozen_at timestamp;

commit;
