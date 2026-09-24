begin;

create index announcements_list on announcements("namespace", created_at desc, id desc) where deleted_at is null;

commit;
