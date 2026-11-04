-- Requires logs.unit_id on amount logs again. Fails while any amount log has
-- no unit_id; restore those ids before running it.
begin;

set local lock_timeout = '5s';

alter table logs drop constraint logs_has_tracking_data;
alter table logs
  add constraint logs_has_tracking_data
  check (
    duration_seconds is not null
    or (
      amount is not null
      and unit_id is not null
      and unit_key is not null
      and modifier is not null
    )
  )
  not valid;
alter table logs validate constraint logs_has_tracking_data;

commit;
