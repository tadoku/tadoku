-- name: FindActivePlatformScoringRuleSet :one
select scoring_rule_sets.*
from platform_scoring_config
inner join scoring_rule_sets
  on scoring_rule_sets.id = platform_scoring_config.active_rule_set_id
where platform_scoring_config.singleton = true;

-- name: FindScoringRuleSetByID :one
select *
from scoring_rule_sets
where id = sqlc.arg('id');

-- name: FindContestScoringRuleSetID :one
select scoring_rule_set_id
from contests
where id = sqlc.arg('contest_id')
  and deleted_at is null;

-- name: ListScoringRulesForRuleSet :many
select *
from scoring_rules
where rule_set_id = sqlc.arg('rule_set_id')
order by priority asc;

-- name: ListPlatformScoringRuleSets :many
select *
from scoring_rule_sets
where scope = 'platform'
order by version desc;

-- name: ListContestScoringRuleSets :many
select *
from scoring_rule_sets
where scope = 'contest'
  and contest_id = sqlc.arg('contest_id')
order by version desc;

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
  fallback_rule_set_id
) values (
  sqlc.arg('id'),
  sqlc.arg('scope'),
  sqlc.arg('contest_id'),
  sqlc.arg('version'),
  'draft',
  sqlc.arg('mode'),
  sqlc.arg('fallback_rule_set_id')
)
returning *;

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
returning *;

-- name: ActivatePlatformScoringRuleSet :exec
insert into platform_scoring_config (
  singleton,
  active_rule_set_id
) values (
  true,
  sqlc.arg('rule_set_id')
)
on conflict (singleton) do update
set active_rule_set_id = excluded.active_rule_set_id;

-- name: ActivateContestScoringRuleSet :exec
update contests
set
  scoring_rule_set_id = sqlc.arg('rule_set_id'),
  updated_at = sqlc.arg('updated_at')
where id = sqlc.arg('contest_id')
  and deleted_at is null;
