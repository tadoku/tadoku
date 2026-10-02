-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'user_roles_tenant_user_id'.
-- Drop an invalid index with drop index concurrently user_roles_tenant_user_id;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently user_roles_tenant_user_id
  on user_roles (tenant, user_id);
