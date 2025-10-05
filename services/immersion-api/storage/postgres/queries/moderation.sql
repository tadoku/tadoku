-- name: CreateModerationAuditLog :exec
insert into moderation_audit_log (
  user_id,
  action,
  metadata,
  description
) values (
  sqlc.arg('user_id'),
  sqlc.arg('action'),
  sqlc.arg('metadata'),
  sqlc.arg('description')
);