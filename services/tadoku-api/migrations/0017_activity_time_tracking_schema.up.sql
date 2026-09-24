begin;

alter table logs add column duration_seconds integer;
alter table logs add column computed_score real;

alter table logs alter column unit_id drop not null;
alter table logs alter column amount drop not null;
alter table logs alter column modifier drop not null;
alter table logs alter column score drop not null;

alter table logs
  add constraint logs_duration_seconds_positive
  check (duration_seconds is null or duration_seconds > 0);

alter table logs
  add constraint logs_has_tracking_data
  check (
    duration_seconds is not null
    or (amount is not null and unit_id is not null and modifier is not null)
  );

alter table contest_logs add column duration_seconds integer;
alter table contest_logs add column computed_score real;

alter table contest_logs alter column amount drop not null;
alter table contest_logs alter column modifier drop not null;

alter table contest_logs
  add constraint contest_logs_duration_seconds_positive
  check (duration_seconds is null or duration_seconds > 0);

alter table contest_logs
  add constraint contest_logs_has_tracking_data
  check (
    duration_seconds is not null
    or (amount is not null and modifier is not null)
  );

commit;
