begin;

create table scoring_rule_sets (
  id uuid primary key default uuid_generate_v4(),
  scope text not null,
  contest_id uuid references contests(id),
  version integer not null,
  status text not null,
  mode text,
  fallback_rule_set_id uuid references scoring_rule_sets(id),
  created_at timestamp not null default now(),
  published_at timestamp,

  constraint scoring_rule_sets_scope_valid
    check (scope in ('platform', 'contest')),
  constraint scoring_rule_sets_version_positive
    check (version > 0),
  constraint scoring_rule_sets_status_valid
    check (status in ('draft', 'published')),
  constraint scoring_rule_sets_status_timestamp_valid
    check (
      (status = 'draft' and published_at is null)
      or (status = 'published' and published_at is not null)
    ),
  constraint scoring_rule_sets_ownership_valid
    check (
      (
        scope = 'platform'
        and contest_id is null
        and mode is null
        and fallback_rule_set_id is null
      )
      or (
        scope = 'contest'
        and contest_id is not null
        and (
          (mode = 'replace' and fallback_rule_set_id is null)
          or (mode = 'override' and fallback_rule_set_id is not null)
        )
      )
    )
);

create unique index scoring_rule_sets_platform_version
  on scoring_rule_sets(version)
  where scope = 'platform';

create unique index scoring_rule_sets_contest_version
  on scoring_rule_sets(contest_id, version)
  where scope = 'contest';

create table scoring_rules (
  id uuid primary key default uuid_generate_v4(),
  rule_set_id uuid not null references scoring_rule_sets(id) on delete cascade,
  priority integer not null,
  stackable boolean not null,
  activity_id smallint not null,
  unit_key text,
  language_code varchar(10) references languages(code),
  tag varchar(50),
  score_source text not null,
  rate real not null,

  constraint scoring_rules_priority_non_negative
    check (priority >= 0),
  constraint scoring_rules_activity_valid
    check (activity_id between 1 and 5),
  constraint scoring_rules_unit_key_valid
    check (
      unit_key is null
      or (
        activity_id = 1
        and unit_key in (
          'reading_page',
          'reading_two_column_page',
          'reading_comic_page',
          'reading_sentence',
          'reading_character'
        )
      )
      or (
        activity_id = 2
        and unit_key in ('listening_minute', 'listening_dense_minutes')
      )
      or (
        activity_id = 3
        and unit_key in ('writing_page', 'writing_sentence', 'writing_character')
      )
      or (
        activity_id = 4
        and unit_key in ('speaking_minute', 'speaking_dense_minutes')
      )
      or (activity_id = 5 and unit_key = 'study_minute')
    ),
  constraint scoring_rules_language_code_normalized
    check (
      language_code is null
      or (
        language_code <> ''
        and language_code = lower(btrim(language_code))
      )
    ),
  constraint scoring_rules_tag_normalized
    check (
      tag is null
      or (tag <> '' and tag = lower(btrim(tag)))
    ),
  constraint scoring_rules_score_source_valid
    check (score_source in ('amount', 'duration_minutes')),
  constraint scoring_rules_rate_valid
    check (rate >= 0 and rate < 'Infinity'::real)
);

create unique index scoring_rules_rule_set_priority
  on scoring_rules(rule_set_id, priority);

create table platform_scoring_config (
  singleton boolean primary key default true,
  active_rule_set_id uuid not null references scoring_rule_sets(id),

  constraint platform_scoring_config_singleton
    check (singleton)
);

alter table contests
  add column scoring_rule_set_id uuid references scoring_rule_sets(id);

insert into scoring_rule_sets (
  scope,
  version,
  status,
  published_at
) values (
  'platform',
  1,
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
    (300, false, 3, 'writing_page', null, 'amount', 1),
    (310, false, 3, 'writing_sentence', null, 'amount', 0.05),
    (320, false, 3, 'writing_character', 'jpn', 'amount', 0.0025),
    (321, false, 3, 'writing_character', 'kor', 'amount', 0.0025),
    (322, false, 3, 'writing_character', 'zho', 'amount', 0.0025),
    (323, false, 3, 'writing_character', 'cmn', 'amount', 0.0025),
    (324, false, 3, 'writing_character', 'yue', 'amount', 0.0025),
    (325, false, 3, 'writing_character', 'wuu', 'amount', 0.0025),
    (330, false, 3, 'writing_character', null, 'amount', 0.000833333),
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
  and rule_set.version = 1;

insert into platform_scoring_config (
  singleton,
  active_rule_set_id
) select
  true,
  id
from scoring_rule_sets
where scope = 'platform'
  and version = 1;

commit;
