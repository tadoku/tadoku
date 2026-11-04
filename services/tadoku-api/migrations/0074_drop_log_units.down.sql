-- Recreates log_units with the rows production held when it was dropped and its
-- row-security policies, and points amount logs' unit_id at the row for their
-- unit, activity and language, or the unit's generic row.
begin;

set local lock_timeout = '5s';

create table log_units (
  id uuid primary key default uuid_generate_v4(),
  log_activity_id smallint not null,
  name varchar(50) not null,
  modifier real not null,
  language_code varchar(10),
  unit_key text not null
);

create index log_units_log_activity_id on log_units(log_activity_id);
create index log_units_unit_key on log_units(unit_key);

insert into log_units (id, log_activity_id, name, modifier, language_code, unit_key)
values
  ('ce71aa77-ad84-4055-ad4f-f56433a6fa99', 1, 'Character', 0.000833333, null, 'characters'),
  ('5348a122-1701-43a6-a03e-d5b974727661', 1, 'Character', 0.0025, 'cmn', 'characters'),
  ('a49da82f-4c96-4e9f-aff7-be6968fb96f8', 1, 'Character', 0.0025, 'jpn', 'characters'),
  ('5d437a28-1e7c-4af3-a12d-986b94f158e1', 1, 'Character', 0.0025, 'kor', 'characters'),
  ('c5fa1ba6-a913-4596-ad79-3c6364b50b04', 1, 'Character', 0.0025, 'wuu', 'characters'),
  ('a5151543-dc6b-478c-814c-6f4d7f34a071', 1, 'Character', 0.0025, 'yue', 'characters'),
  ('22cb14f3-2642-475f-a620-fba3602fa8f5', 1, 'Character', 0.0025, 'zho', 'characters'),
  ('17295a2f-fd84-4217-845b-96d6e12039d9', 1, 'Page', 1, null, 'pages'),
  ('4c1925e2-abce-418b-88c7-d2f9f82a571a', 1, 'Comic page', 0.2, null, 'reading_comic_page'),
  ('b6a0aafc-ade6-43c2-a8b2-794608f0a12d', 1, '2 Column page', 1.6, 'jpn', 'reading_two_column_page'),
  ('db2f3370-63d6-49d8-9b2c-acc7e1a9ee76', 1, 'Sentence', 0.05, null, 'sentences'),
  ('1a86ac9e-ddf0-4795-8b32-00d01db54d09', 2, 'Dense minute', 0.6, null, 'listening_dense_minutes'),
  ('946bc5ee-5f1f-42a0-a96d-296913ef3d6f', 2, 'Minute', 0.4, null, 'listening_minute'),
  ('f6ca1ee5-3d87-4d50-be61-8ac9e9e9eba1', 3, 'Character', 0.00833333, null, 'characters'),
  ('1c1d3e55-a117-4422-99e8-c296baeb02cc', 3, 'Character', 0.025, 'cmn', 'characters'),
  ('3aaf23f7-f845-45ba-a936-63c4c1150695', 3, 'Character', 0.025, 'jpn', 'characters'),
  ('643acca5-2141-40a1-a740-fb8a2f61ed55', 3, 'Character', 0.025, 'kor', 'characters'),
  ('e482d586-a6ed-4f3d-9cb6-a06142d42dc0', 3, 'Character', 0.025, 'wuu', 'characters'),
  ('422d7715-cfbe-439a-b5e9-7d368b4dd69b', 3, 'Character', 0.025, 'yue', 'characters'),
  ('97b0ec5f-a431-485b-9edc-70954220e21b', 3, 'Character', 0.025, 'zho', 'characters'),
  ('395908fd-3407-4642-a544-b761a57b98b1', 3, 'Page', 10, null, 'pages'),
  ('872fab8b-6f89-4ae6-93ef-b3d394a8f2bb', 3, 'Sentence', 0.5, null, 'sentences'),
  ('53c032e5-36bc-44ed-8f97-226d602572ee', 4, 'Dense minute', 0.7, null, 'speaking_dense_minutes'),
  ('681fa865-eba8-48c1-b48c-8e99ffedb625', 4, 'Minute', 0.5, null, 'speaking_minute'),
  ('a981cfca-0cb7-4c3f-bcad-fed272859beb', 5, 'Minute', 0.5, null, 'study_minute');

alter table log_units enable row level security;

create policy shared_read on log_units for select using (true);

create policy production_insert on log_units for insert with check (current_setting('tadoku.tenant', true) = 'tadoku/prod');

create policy production_update on log_units for update using (current_setting('tadoku.tenant', true) = 'tadoku/prod');

create policy production_delete on log_units for delete using (current_setting('tadoku.tenant', true) = 'tadoku/prod');

update logs
set unit_id = (
  select log_units.id
  from log_units
  where log_units.unit_key = logs.unit_key
    and log_units.log_activity_id = logs.log_activity_id
    and (log_units.language_code = logs.language_code or log_units.language_code is null)
  order by log_units.language_code nulls last
  limit 1
)
where logs.unit_key is not null;

commit;
