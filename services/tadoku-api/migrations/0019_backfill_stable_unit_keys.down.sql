begin;

update contest_logs set unit_key = null;
update logs set unit_key = null;
update log_units set unit_key = null;

commit;
