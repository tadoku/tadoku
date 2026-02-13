package domain

import (
	"context"

	"github.com/google/uuid"
)

type ModerationAuditLogCreateRequest struct {
	ModeratorUserID uuid.UUID
	Action          string
	Metadata        map[string]any
	Description     *string
}

type ModerationAuditRepository interface {
	CreateModerationAuditLog(ctx context.Context, req *ModerationAuditLogCreateRequest) error
}
