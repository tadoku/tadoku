-- Add scoring columns to contest_logs (nullable first for backfill)
alter table contest_logs add column amount real;
alter table contest_logs add column modifier real;

-- Backfill from parent logs
update contest_logs
set
  amount = logs.amount,
  modifier = logs.modifier
from logs
where contest_logs.log_id = logs.id;

-- Make columns NOT NULL after backfill
alter table contest_logs alter column amount set not null;
alter table contest_logs alter column modifier set not null;

-- Add generated score column
alter table contest_logs add column score real generated always as (amount * modifier) stored;
