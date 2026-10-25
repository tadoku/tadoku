-- Run outside a transaction so writes can continue while the index builds.
-- If creation fails, inspect pg_index.indisvalid for 'permission_relationships_subject'.
-- Drop an invalid index with drop index concurrently permission_relationships_subject;
-- before retrying this migration after the reviewed metadata repair.
create index concurrently permission_relationships_subject
  on permission_relationships (tenant, subject_type, subject_id, subject_relation, object_type, relation);
