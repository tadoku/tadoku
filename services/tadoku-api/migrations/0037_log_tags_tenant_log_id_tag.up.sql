-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'log_tags_tenant_log_id_tag'.
-- Drop an invalid index with drop index concurrently log_tags_tenant_log_id_tag;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently log_tags_tenant_log_id_tag
  on log_tags (tenant, log_id, tag);
