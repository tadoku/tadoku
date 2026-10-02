-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'scoring_rules_tenant_rule_set_priority'.
-- Drop an invalid index with drop index concurrently scoring_rules_tenant_rule_set_priority;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently scoring_rules_tenant_rule_set_priority
  on scoring_rules (tenant, rule_set_id, priority);
