begin;

alter table log_units add column unit_key text;
alter table logs add column unit_key text;
alter table contest_logs add column unit_key text;

commit;
