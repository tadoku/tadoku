-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'contest_registrations_tenant_id'.
-- Drop an invalid index with drop index concurrently contest_registrations_tenant_id;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently contest_registrations_tenant_id
  on contest_registrations (tenant, id);
