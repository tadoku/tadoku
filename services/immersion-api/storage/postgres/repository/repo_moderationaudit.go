package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) CreateModerationAuditLog(ctx context.Context, req *domain.ModerationAuditLogCreateRequest) error {
	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("could not marshal metadata: %w", err)
	}

	err = r.q.CreateModerationAuditLog(ctx, postgres.CreateModerationAuditLogParams{
		UserID:      req.ModeratorUserID,
		Action:      req.Action,
		Metadata:    metadataJSON,
		Description: postgres.NewNullString(req.Description),
	})
	if err != nil {
		return fmt.Errorf("could not create moderation audit log: %w", err)
	}

	return nil
}
