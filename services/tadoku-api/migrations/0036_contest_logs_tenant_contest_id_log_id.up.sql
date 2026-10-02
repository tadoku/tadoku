-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'contest_logs_tenant_contest_id_log_id'.
-- Drop an invalid index with drop index concurrently contest_logs_tenant_contest_id_log_id;
-- before retrying this migration after the reviewed metadata repair.
create unique index concurrently contest_logs_tenant_contest_id_log_id
  on contest_logs (tenant, contest_id, log_id);
