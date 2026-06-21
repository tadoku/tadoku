begin;

alter table contest_logs drop constraint contest_logs_has_tracking_data;
alter table contest_logs drop constraint contest_logs_duration_seconds_positive;

alter table contest_logs drop column computed_score;
alter table contest_logs drop column duration_seconds;

alter table contest_logs alter column amount set not null;
alter table contest_logs alter column modifier set not null;

alter table logs drop constraint logs_has_tracking_data;
alter table logs drop constraint logs_duration_seconds_positive;

alter table logs drop column computed_score;
alter table logs drop column duration_seconds;

alter table logs alter column unit_id set not null;
alter table logs alter column amount set not null;
alter table logs alter column modifier set not null;
alter table logs alter column score set not null;

commit;
