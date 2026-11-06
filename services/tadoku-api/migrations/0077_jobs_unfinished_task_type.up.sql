-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'jobs_unfinished_task_type'.
-- Drop an invalid index with drop index concurrently jobs_unfinished_task_type;
-- before retrying this migration after the reviewed metadata repair.
create index concurrently jobs_unfinished_task_type
  on jobs (task_type)
  include (tenant, state, next_attempt_at, lease_expires_at)
  where state <> 'completed';
