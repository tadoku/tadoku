begin;

insert into scoring_rule_sets (
  scope,
  version,
  status,
  published_at
) values (
  'platform',
  2,
  'published',
  now()
);

insert into scoring_rules (
  rule_set_id,
  priority,
  stackable,
  activity_id,
  unit_key,
  language_code,
  score_source,
  rate
) select
  rule_set.id,
  rule.priority,
  rule.stackable,
  rule.activity_id,
  rule.unit_key,
  rule.language_code,
  rule.score_source,
  rule.rate
from scoring_rule_sets as rule_set
cross join (
  values
    (100, false, 1, 'reading_two_column_page', 'jpn', 'amount', 1.6),
    (110, false, 1, 'reading_page', null, 'amount', 1),
    (120, false, 1, 'reading_comic_page', null, 'amount', 0.2),
    (130, false, 1, 'reading_sentence', null, 'amount', 0.05),
    (140, false, 1, 'reading_character', 'jpn', 'amount', 0.0025),
    (141, false, 1, 'reading_character', 'kor', 'amount', 0.0025),
    (142, false, 1, 'reading_character', 'zho', 'amount', 0.0025),
    (143, false, 1, 'reading_character', 'cmn', 'amount', 0.0025),
    (144, false, 1, 'reading_character', 'yue', 'amount', 0.0025),
    (145, false, 1, 'reading_character', 'wuu', 'amount', 0.0025),
    (150, false, 1, 'reading_character', null, 'amount', 0.000833333),
    (200, false, 2, null, null, 'amount', 0.4),
    (210, true, 2, 'listening_dense_minutes', null, 'amount', 1.5),
    (300, false, 3, 'writing_page', null, 'amount', 10),
    (310, false, 3, 'writing_sentence', null, 'amount', 0.5),
    (320, false, 3, 'writing_character', 'jpn', 'amount', 0.025),
    (321, false, 3, 'writing_character', 'kor', 'amount', 0.025),
    (322, false, 3, 'writing_character', 'zho', 'amount', 0.025),
    (323, false, 3, 'writing_character', 'cmn', 'amount', 0.025),
    (324, false, 3, 'writing_character', 'yue', 'amount', 0.025),
    (325, false, 3, 'writing_character', 'wuu', 'amount', 0.025),
    (330, false, 3, 'writing_character', null, 'amount', 0.00833333),
    (400, false, 4, null, null, 'amount', 0.5),
    (410, true, 4, 'speaking_dense_minutes', null, 'amount', 1.4),
    (500, false, 5, 'study_minute', null, 'amount', 0.5),
    (600, false, 1, null, null, 'duration_minutes', 0.2),
    (610, false, 2, null, null, 'duration_minutes', 0.4),
    (620, false, 3, null, null, 'duration_minutes', 0.2),
    (630, false, 4, null, null, 'duration_minutes', 0.5),
    (640, false, 5, null, null, 'duration_minutes', 0.5)
) as rule (
  priority,
  stackable,
  activity_id,
  unit_key,
  language_code,
  score_source,
  rate
)
where rule_set.scope = 'platform'
  and rule_set.version = 2;

update platform_scoring_config
set active_rule_set_id = (
  select id
  from scoring_rule_sets
  where scope = 'platform'
    and version = 2
);

commit;
