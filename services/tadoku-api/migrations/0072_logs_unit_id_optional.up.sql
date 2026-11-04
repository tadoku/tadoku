-- Stops requiring logs.unit_id on amount logs. unit_key identifies the unit,
-- and the application stops writing unit_id before the column is dropped.
-- Every existing row satisfies the relaxed check, so validation cannot fail.
begin;

set local lock_timeout = '5s';

alter table logs drop constraint logs_has_tracking_data;
alter table logs
  add constraint logs_has_tracking_data
  check (
    duration_seconds is not null
    or (
      amount is not null
      and unit_key is not null
      and modifier is not null
    )
  )
  not valid;
alter table logs validate constraint logs_has_tracking_data;

commit;
