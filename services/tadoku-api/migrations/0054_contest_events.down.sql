begin;

set local lock_timeout = '5s';

drop table contest_events;

commit;
