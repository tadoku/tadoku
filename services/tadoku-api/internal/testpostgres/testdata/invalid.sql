-- Fail after cleanup and an earlier seed file to exercise setup rollback.
insert into missing_test_table values (1);
