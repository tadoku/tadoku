-- Re-adds logs.unit_id as a nullable column. Its former values referenced the
-- dropped log_units table and are not restored.
begin;

set local lock_timeout = '5s';

alter table logs add column unit_id uuid;

commit;
