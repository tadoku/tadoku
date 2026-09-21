-- name: CreateAudit :exec
insert into moderation_audit_log (
  user_id,
  action,
  metadata,
  description,
  created_at
) values (
  sqlc.arg(actor_id)::uuid,
  sqlc.arg(action),
  sqlc.arg(metadata)::jsonb,
  sqlc.arg(description)::text,
  sqlc.arg(recorded_at)::timestamp
);
