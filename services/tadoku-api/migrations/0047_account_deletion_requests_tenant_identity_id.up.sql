-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'account_deletion_requests_tenant_identity_id'.
-- Drop an invalid index with drop index concurrently account_deletion_requests_tenant_identity_id;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently account_deletion_requests_tenant_identity_id
  on account_deletion_requests (tenant, identity_id);
