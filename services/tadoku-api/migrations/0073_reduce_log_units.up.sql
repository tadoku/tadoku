-- Reduces log units to pages, sentences and characters for reading and writing.
-- Comic and two-column pages become pages with a comic or two_column tag.
-- Minute amounts on listening, speaking and study become durations, with a
-- dense tag for dense minutes. Stored scores do not change: computed_score keeps
-- the value of the generated score column before an amount is cleared, and
-- score provenance is never rewritten. Rules in every rule set are renamed in
-- place, so contests that fall back to an older platform set keep matching.
-- Every tenant whose active platform rule set names a retired key or scores an
-- amount on a timed activity gets a new published version without those rules,
-- with the id md5('0073_reduce_log_units:' || tenant). Older rule sets keep
-- their retired-unit rules, so scoring_rules still accepts the retired keys.
-- The comic, two-column and minute log_units rows keep their retired keys until
-- the table is dropped; the application stops offering them.
begin;

set local lock_timeout = '5s';

create temporary table retired_unit_keys (
  unit_key text primary key,
  tag text
) on commit drop;

insert into retired_unit_keys (unit_key, tag)
values
  ('reading_comic_page', 'comic'),
  ('reading_two_column_page', 'two_column'),
  ('listening_minute', null),
  ('listening_dense_minutes', 'dense'),
  ('speaking_minute', null),
  ('speaking_dense_minutes', 'dense'),
  ('study_minute', null);

create temporary table renamed_unit_keys (
  old_key text primary key,
  new_key text not null
) on commit drop;

insert into renamed_unit_keys (old_key, new_key)
values
  ('reading_page', 'pages'),
  ('writing_page', 'pages'),
  ('reading_comic_page', 'pages'),
  ('reading_two_column_page', 'pages'),
  ('reading_sentence', 'sentences'),
  ('writing_sentence', 'sentences'),
  ('reading_character', 'characters'),
  ('writing_character', 'characters');

alter table scoring_rules drop constraint scoring_rules_unit_key_valid;
alter table scoring_rules
  add constraint scoring_rules_unit_key_valid
  check (
    unit_key is null
    or (
      activity_id in (1, 3)
      and unit_key in ('pages', 'sentences', 'characters')
    )
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
    or (
      activity_id = 5
      and unit_key = 'study_minute'
    )
  )
  not valid;
alter table scoring_rules validate constraint scoring_rules_unit_key_valid;

insert into log_tags (tenant, log_id, user_id, tag)
select
  logs.tenant,
  logs.id,
  logs.user_id,
  retired_unit_keys.tag
from logs
inner join retired_unit_keys on retired_unit_keys.unit_key = logs.unit_key
where retired_unit_keys.tag is not null
on conflict do nothing;

update contest_logs
set
  computed_score = coalesce(contest_logs.computed_score, contest_logs.score),
  duration_seconds = coalesce(contest_logs.duration_seconds, round(contest_logs.amount * 60)::integer),
  amount = null,
  modifier = null,
  unit_key = null
from retired_unit_keys
where retired_unit_keys.unit_key = contest_logs.unit_key
  and retired_unit_keys.unit_key not in ('reading_comic_page', 'reading_two_column_page');

update logs
set
  computed_score = coalesce(logs.computed_score, logs.score),
  duration_seconds = coalesce(logs.duration_seconds, round(logs.amount * 60)::integer),
  amount = null,
  modifier = null,
  unit_key = null,
  unit_id = null
from retired_unit_keys
where retired_unit_keys.unit_key = logs.unit_key
  and retired_unit_keys.unit_key not in ('reading_comic_page', 'reading_two_column_page');

update logs
set unit_id = (
  select log_units.id
  from log_units
  where log_units.unit_key = 'reading_page'
    and log_units.log_activity_id = 1
    and log_units.language_code is null
)
where logs.unit_key in ('reading_comic_page', 'reading_two_column_page');

update logs
set unit_key = renamed_unit_keys.new_key
from renamed_unit_keys
where renamed_unit_keys.old_key = logs.unit_key;

update contest_logs
set unit_key = renamed_unit_keys.new_key
from renamed_unit_keys
where renamed_unit_keys.old_key = contest_logs.unit_key;

update log_units
set unit_key = renamed_unit_keys.new_key
from renamed_unit_keys
where renamed_unit_keys.old_key = log_units.unit_key
  and log_units.unit_key not in ('reading_comic_page', 'reading_two_column_page');

update scoring_rules
set unit_key = renamed_unit_keys.new_key
from renamed_unit_keys
where renamed_unit_keys.old_key = scoring_rules.unit_key
  and renamed_unit_keys.old_key not in ('reading_comic_page', 'reading_two_column_page');

create temporary table replaced_platform_rule_sets (
  tenant text primary key,
  old_rule_set_id uuid not null,
  new_rule_set_id uuid not null,
  new_version integer not null
) on commit drop;

insert into replaced_platform_rule_sets (tenant, old_rule_set_id, new_rule_set_id, new_version)
select
  platform_scoring_config.tenant,
  platform_scoring_config.active_rule_set_id,
  md5('0073_reduce_log_units:' || platform_scoring_config.tenant)::uuid,
  (
    select max(versions.version) + 1
    from scoring_rule_sets as versions
    where versions.tenant = platform_scoring_config.tenant
      and versions.scope = 'platform'
  )
from platform_scoring_config
where exists (
  select
  from scoring_rules
  where scoring_rules.tenant = platform_scoring_config.tenant
    and scoring_rules.rule_set_id = platform_scoring_config.active_rule_set_id
    and (
      scoring_rules.unit_key in (select unit_key from retired_unit_keys)
      or (scoring_rules.activity_id in (2, 4, 5) and scoring_rules.score_source = 'amount')
    )
);

insert into scoring_rule_sets (tenant, id, scope, version, status, created_at, published_at)
select
  tenant,
  new_rule_set_id,
  'platform',
  new_version,
  'published',
  now(),
  now()
from replaced_platform_rule_sets
on conflict do nothing;

insert into scoring_rules (
  tenant,
  rule_set_id,
  priority,
  stackable,
  activity_id,
  unit_key,
  language_code,
  tag,
  score_source,
  rate
)
select
  scoring_rules.tenant,
  replaced_platform_rule_sets.new_rule_set_id,
  scoring_rules.priority,
  scoring_rules.stackable,
  scoring_rules.activity_id,
  scoring_rules.unit_key,
  scoring_rules.language_code,
  scoring_rules.tag,
  scoring_rules.score_source,
  scoring_rules.rate
from scoring_rules
inner join replaced_platform_rule_sets
  on replaced_platform_rule_sets.tenant = scoring_rules.tenant
  and replaced_platform_rule_sets.old_rule_set_id = scoring_rules.rule_set_id
where (scoring_rules.unit_key is null or scoring_rules.unit_key not in (select unit_key from retired_unit_keys))
  and not (scoring_rules.activity_id in (2, 4, 5) and scoring_rules.score_source = 'amount')
on conflict do nothing;

update platform_scoring_config
set active_rule_set_id = replaced_platform_rule_sets.new_rule_set_id
from replaced_platform_rule_sets
where replaced_platform_rule_sets.tenant = platform_scoring_config.tenant;

do $$
begin
  if exists (
    select
    from logs
    where unit_key is not null
      and unit_key not in ('pages', 'sentences', 'characters')
  ) then
    raise exception 'a log still names a retired or renamed unit key';
  end if;

  if exists (
    select
    from contest_logs
    where unit_key is not null
      and unit_key not in ('pages', 'sentences', 'characters')
  ) then
    raise exception 'a contest log still names a retired or renamed unit key';
  end if;

  if exists (
    select
    from logs
    where amount is not null
      and log_activity_id in (2, 4, 5)
  ) then
    raise exception 'a timed activity log still has an amount';
  end if;

  if exists (
    select
    from contest_logs
    inner join logs on logs.tenant = contest_logs.tenant and logs.id = contest_logs.log_id
    where contest_logs.amount is not null
      and logs.log_activity_id in (2, 4, 5)
  ) then
    raise exception 'a timed activity contest log still has an amount';
  end if;

  if exists (
    select
    from platform_scoring_config
    inner join scoring_rules
      on scoring_rules.tenant = platform_scoring_config.tenant
      and scoring_rules.rule_set_id = platform_scoring_config.active_rule_set_id
    where (
        scoring_rules.unit_key is not null
        and scoring_rules.unit_key not in ('pages', 'sentences', 'characters')
      )
      or (scoring_rules.activity_id in (2, 4, 5) and scoring_rules.score_source = 'amount')
  ) then
    raise exception 'an active platform rule set still names a retired unit or scores a timed amount';
  end if;
end $$;

commit;
