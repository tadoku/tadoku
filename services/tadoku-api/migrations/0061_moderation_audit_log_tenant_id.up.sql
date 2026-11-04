-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'moderation_audit_log_tenant_id'.
-- Drop an invalid index with drop index concurrently moderation_audit_log_tenant_id;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently moderation_audit_log_tenant_id
  on moderation_audit_log (tenant, id);
