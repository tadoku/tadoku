-- name: CreateModerationAudit :exec
insert into moderation_audit_log (
  user_id,
  action,
  metadata,
  description,
  created_at
) values (
  sqlc.arg(moderator_user_id)::uuid,
  sqlc.arg(action),
  sqlc.arg(metadata)::jsonb,
  sqlc.arg(description)::text,
  sqlc.arg(created_at)::timestamp
);
