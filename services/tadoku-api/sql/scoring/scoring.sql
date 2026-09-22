-- name: FindActivePlatformScoringRuleSet :one
select scoring_rule_sets.*
from platform_scoring_config
inner join scoring_rule_sets on scoring_rule_sets.id = platform_scoring_config.active_rule_set_id
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

-- name: FindUnitForScoringByID :one
select id, unit_key, modifier
from log_units
where id = sqlc.arg('id')
  and log_activity_id = sqlc.arg('activity_id')
  and (language_code is null or language_code = sqlc.arg('language_code'));

-- name: FindUnitForScoringByKey :one
select id, unit_key, modifier
from log_units
where unit_key = sqlc.arg('unit_key')
  and log_activity_id = sqlc.arg('activity_id')
  and (language_code is null or language_code = sqlc.arg('language_code'))
order by language_code is null asc
limit 1;
