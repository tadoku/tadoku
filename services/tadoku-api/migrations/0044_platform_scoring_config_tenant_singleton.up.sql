-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'platform_scoring_config_tenant_singleton'.
-- Drop an invalid index with drop index concurrently platform_scoring_config_tenant_singleton;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently platform_scoring_config_tenant_singleton
  on platform_scoring_config (tenant, singleton);
