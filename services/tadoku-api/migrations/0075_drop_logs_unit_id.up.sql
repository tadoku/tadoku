-- Drops logs.unit_id. Logs identify their unit by unit_key, and the log_units
-- table it referenced was dropped in migration 0074; no deployed code reads or
-- writes the column.
begin;

set local lock_timeout = '5s';

alter table logs drop column unit_id;

commit;
