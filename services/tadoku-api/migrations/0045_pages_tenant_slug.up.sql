-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'pages_tenant_slug'.
-- Drop an invalid index with drop index concurrently pages_tenant_slug;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently pages_tenant_slug
  on pages (tenant, "namespace", slug);
