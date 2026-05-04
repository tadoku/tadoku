-- Add duration column to logs
alter table logs
  add column duration_seconds int;

-- Make amount/unit/modifier optional (for time-only logs)
alter table logs
  alter column amount drop not null,
  alter column unit_id drop not null,
  alter column modifier drop not null;

-- Ensure at least one complete tracking method:
-- either duration, or the full amount triple (amount + unit_id + modifier)
alter table logs
  add constraint logs_has_tracking_data
  check (
    duration_seconds is not null
    or (amount is not null and unit_id is not null and modifier is not null)
  );

-- Add time-based modifier and input type to activities
alter table log_activities
  add column time_modifier real not null default 0.3,
  add column input_type varchar(10) not null default 'amount';

-- 'time' = Listening, Speaking, Study (time-only input)
-- 'amount' = Reading, Writing (amount+unit primary, time optional)
update log_activities set input_type = 'time' where id in (2, 4, 5);
update log_activities set input_type = 'amount' where id in (1, 3);

-- Keep existing generated score column untouched (preserves all historical scores).
-- Add a new computed_score column for application-written scores (used for new logs).
alter table logs
  add column computed_score real;

-- contest_logs: same treatment (has amount, modifier, and generated score from migration 0013)
alter table contest_logs
  add column duration_seconds int;

alter table contest_logs
  alter column amount drop not null,
  alter column modifier drop not null;

-- Keep existing generated score (amount * modifier) intact for historical contest entries.
-- Add computed_score for new entries.
alter table contest_logs
  add column computed_score real;
