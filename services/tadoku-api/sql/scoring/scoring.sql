-- name: FindActivePlatformScoringRuleSet :one
select
  scoring_rule_sets.scope,
  scoring_rule_sets.id,
  scoring_rule_sets.contest_id,
  scoring_rule_sets.version,
  scoring_rule_sets.status,
  scoring_rule_sets.mode,
  scoring_rule_sets.fallback_rule_set_id,
  scoring_rule_sets.created_at,
  scoring_rule_sets.published_at
from platform_scoring_config
inner join scoring_rule_sets on scoring_rule_sets.id = platform_scoring_config.active_rule_set_id
where platform_scoring_config.singleton = true;

-- name: FindScoringRuleSetByID :one
select
  scope,
  id,
  contest_id,
  version,
  status,
  mode,
  fallback_rule_set_id,
  created_at,
  published_at
from scoring_rule_sets
where id = sqlc.arg('id');

-- name: FindContestScoringRuleSetID :one
select scoring_rule_set_id
from contests
where id = sqlc.arg('contest_id')
  and deleted_at is null;

-- name: ListScoringRulesForRuleSet :many
select
  priority,
  id,
  rule_set_id,
  stackable,
  activity_id,
  unit_key,
  language_code,
  tag,
  score_source,
  rate
from scoring_rules
where rule_set_id = sqlc.arg('rule_set_id')
order by priority asc;

-- name: ListPlatformScoringRuleSets :many
select
  scope,
  id,
  contest_id,
  version,
  status,
  mode,
  fallback_rule_set_id,
  created_at,
  published_at
from scoring_rule_sets
where scope = 'platform'
order by version desc;

-- name: ListContestScoringRuleSets :many
select
  scope,
  id,
  contest_id,
  version,
  status,
  mode,
  fallback_rule_set_id,
  created_at,
  published_at
from scoring_rule_sets
where scope = 'contest'
  and contest_id = sqlc.arg('contest_id')
order by version desc;

-- name: ListUnitsForScoringByID :many
select id, unit_key, modifier, log_activity_id, language_code
from log_units
where id = sqlc.arg('id');

-- name: ListUnitsForScoringByKey :many
select id, unit_key, modifier, log_activity_id, language_code
from log_units
where unit_key = sqlc.arg('unit_key');

-- name: NextPlatformScoringRuleSetVersion :one
select (coalesce(max(version), 0) + 1)::integer
from scoring_rule_sets
where scope = 'platform';

-- name: NextContestScoringRuleSetVersion :one
select (coalesce(max(version), 0) + 1)::integer
from scoring_rule_sets
where scope = 'contest'
  and contest_id = sqlc.arg('contest_id');

-- name: CreateScoringRuleSet :one
insert into scoring_rule_sets (
  id,
  scope,
  contest_id,
  version,
  status,
  mode,
  fallback_rule_set_id,
  created_at
) values (
  sqlc.arg('id'),
  sqlc.arg('scope'),
  sqlc.arg('contest_id'),
  sqlc.arg('version'),
  sqlc.arg('status'),
  sqlc.arg('mode'),
  sqlc.arg('fallback_rule_set_id'),
  sqlc.arg('created_at')
)
returning
  scope,
  id,
  contest_id,
  version,
  status,
  mode,
  fallback_rule_set_id,
  created_at,
  published_at;

-- name: CreateScoringRule :exec
insert into scoring_rules (
  id,
  rule_set_id,
  priority,
  stackable,
  activity_id,
  unit_key,
  language_code,
  tag,
  score_source,
  rate
) values (
  sqlc.arg('id'),
  sqlc.arg('rule_set_id'),
  sqlc.arg('priority'),
  sqlc.arg('stackable'),
  sqlc.arg('activity_id'),
  sqlc.arg('unit_key'),
  sqlc.arg('language_code'),
  sqlc.arg('tag'),
  sqlc.arg('score_source'),
  sqlc.arg('rate')
);

-- name: PublishScoringRuleSet :one
update scoring_rule_sets
set
  status = 'published',
  published_at = sqlc.arg('published_at')
where id = sqlc.arg('id')
  and status = 'draft'
returning
  scope,
  id,
  contest_id,
  version,
  status,
  mode,
  fallback_rule_set_id,
  created_at,
  published_at;

-- name: ActivatePlatformScoringRuleSet :exec
insert into platform_scoring_config (
  singleton,
  active_rule_set_id
) values (
  true,
  sqlc.arg('rule_set_id')
)
on conflict (tenant, singleton) do update
set active_rule_set_id = excluded.active_rule_set_id;

-- name: ActivateContestScoringRuleSet :execrows
update contests
set
  scoring_rule_set_id = sqlc.arg('rule_set_id'),
  updated_at = sqlc.arg('updated_at')
where id = sqlc.arg('contest_id')
  and deleted_at is null;
