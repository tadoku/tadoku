-- Remove computed_score from contest_logs
alter table contest_logs
  drop column if exists computed_score;

-- Remove duration_seconds from contest_logs
alter table contest_logs
  alter column amount set not null,
  alter column modifier set not null;

alter table contest_logs
  drop column if exists duration_seconds;

-- Remove computed_score from logs
alter table logs
  drop column if exists computed_score;

-- Remove input_type and time_modifier from activities
alter table log_activities
  drop column if exists input_type,
  drop column if exists time_modifier;

-- Remove check constraint
alter table logs
  drop constraint if exists logs_has_tracking_data;

-- Restore NOT NULL on amount/unit/modifier
alter table logs
  alter column amount set not null,
  alter column unit_id set not null,
  alter column modifier set not null;

-- Remove duration_seconds from logs
alter table logs
  drop column if exists duration_seconds;
