begin;

alter table contest_logs drop column unit_key;
alter table logs drop column unit_key;
alter table log_units drop column unit_key;

commit;
