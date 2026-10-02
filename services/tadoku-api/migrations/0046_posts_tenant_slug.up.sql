-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'posts_tenant_slug'.
-- Drop an invalid index with drop index concurrently posts_tenant_slug;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently posts_tenant_slug
  on posts (tenant, "namespace", slug);
