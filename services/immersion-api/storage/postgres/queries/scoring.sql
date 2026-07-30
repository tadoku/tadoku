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

-- name: ListScoringRulesForRuleSet :many
select *
from scoring_rules
where rule_set_id = sqlc.arg('rule_set_id')
order by priority asc;
