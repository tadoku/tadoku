begin;

update log_units
set unit_key = case
  when log_activity_id = 1 and name = 'Page' then 'reading_page'
  when log_activity_id = 1 and name = '2 Column page' then 'reading_two_column_page'
  when log_activity_id = 1 and name = 'Comic page' then 'reading_comic_page'
  when log_activity_id = 1 and name = 'Sentence' then 'reading_sentence'
  when log_activity_id = 1 and name = 'Character' then 'reading_character'
  when log_activity_id = 2 and name = 'Minute' then 'listening_minute'
  when log_activity_id = 2 and name = 'Minute (high density)' then 'listening_dense_minutes'
  when log_activity_id = 3 and name = 'Page' then 'writing_page'
  when log_activity_id = 3 and name = 'Sentence' then 'writing_sentence'
  when log_activity_id = 3 and name = 'Character' then 'writing_character'
  when log_activity_id = 4 and name = 'Minute' then 'speaking_minute'
  when log_activity_id = 4 and name = 'Minute (high density)' then 'speaking_dense_minutes'
  when log_activity_id = 5 and name = 'Minute' then 'study_minute'
end
where unit_key is null;

do $$
begin
  if exists (select 1 from log_units where unit_key is null) then
    raise exception 'cannot assign a stable key to every log unit';
  end if;
end $$;

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
  if exists (select 1 from logs where amount is not null and unit_key is null) then
    raise exception 'cannot assign a stable key to every amount-tracked log';
  end if;

  if exists (select 1 from contest_logs where amount is not null and unit_key is null) then
    raise exception 'cannot assign a stable key to every amount-tracked contest log';
  end if;
end $$;

commit;
