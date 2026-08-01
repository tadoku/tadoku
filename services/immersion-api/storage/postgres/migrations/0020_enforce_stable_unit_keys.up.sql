begin;

update logs
set unit_key = log_units.unit_key
from log_units
where logs.unit_key is null
  and logs.unit_id = log_units.id;

update contest_logs
set unit_key = logs.unit_key
from logs
where contest_logs.unit_key is null
  and contest_logs.log_id = logs.id;

do $$
begin
  if exists (select 1 from log_units where unit_key is null) then
    raise exception 'cannot require stable unit keys while log units are unmapped';
  end if;

  if exists (select 1 from logs where amount is not null and unit_key is null) then
    raise exception 'cannot require stable unit keys while amount-tracked logs are unmapped';
  end if;

  if exists (select 1 from contest_logs where amount is not null and unit_key is null) then
    raise exception 'cannot require stable unit keys while amount-tracked contest logs are unmapped';
  end if;
end $$;

alter table log_units alter column unit_key set not null;
create index log_units_unit_key on log_units(unit_key);

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
  );

alter table contest_logs drop constraint contest_logs_has_tracking_data;
alter table contest_logs
  add constraint contest_logs_has_tracking_data
  check (
    duration_seconds is not null
    or (
      amount is not null
      and unit_key is not null
      and modifier is not null
    )
  );

commit;
