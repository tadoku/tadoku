-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'scoring_rule_sets_tenant_id'.
-- Drop an invalid index with drop index concurrently scoring_rule_sets_tenant_id;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently scoring_rule_sets_tenant_id
  on scoring_rule_sets (tenant, id);
