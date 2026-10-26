begin;

set local lock_timeout = '5s';

drop table account_deletion_requests;
drop table profiles;
drop table user_roles;

commit;
