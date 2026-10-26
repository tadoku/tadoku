-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'log_tags_log_id'.
-- Drop an invalid index with drop index concurrently log_tags_log_id;
-- before retrying this migration after the reviewed metadata repair.
create index concurrently log_tags_log_id
  on log_tags (log_id);
