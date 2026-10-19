begin;

set local lock_timeout = '5s';

drop table access_tokens;
drop table emails;
drop table recovery_tokens;
drop table sessions;
drop table accounts;

commit;
