begin;

alter table contest_logs drop constraint contest_logs_has_tracking_data;
alter table contest_logs
  add constraint contest_logs_has_tracking_data
  check (
    duration_seconds is not null
    or (amount is not null and modifier is not null)
  );

alter table logs drop constraint logs_has_tracking_data;
alter table logs
  add constraint logs_has_tracking_data
  check (
    duration_seconds is not null
    or (amount is not null and unit_id is not null and modifier is not null)
  );

drop index log_units_unit_key;
alter table log_units alter column unit_key drop not null;

commit;
