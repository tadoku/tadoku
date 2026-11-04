-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'pages_tenant_id'.
-- Drop an invalid index with drop index concurrently pages_tenant_id;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently pages_tenant_id
  on pages (tenant, id);
