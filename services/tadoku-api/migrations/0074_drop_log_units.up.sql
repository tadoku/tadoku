-- Drops the log_units table. Unit names, rates and language variants live in
-- the application's unit catalog, and logs identify their unit by unit_key. The
-- row-security policies on log_units are dropped with it. logs.unit_id stays as
-- an unused nullable column until a reviewed change may drop it. Rules
-- stored with an activity-prefixed page, sentence or character key since
-- migration 0073 get the canonical key.
begin;

set local lock_timeout = '5s';

update scoring_rules
set unit_key = case unit_key
  when 'reading_page' then 'pages'
  when 'writing_page' then 'pages'
  when 'reading_sentence' then 'sentences'
  when 'writing_sentence' then 'sentences'
  else 'characters'
end
where unit_key in (
  'reading_page',
  'writing_page',
  'reading_sentence',
  'writing_sentence',
  'reading_character',
  'writing_character'
);

drop table log_units;

commit;
