package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/tadoku/tadoku/services/authz-api/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateModerationAuditLog(ctx context.Context, req *domain.ModerationAuditLogCreateRequest) error {
	if req == nil {
		return fmt.Errorf("audit req is required")
	}

	metadata, err := json.Marshal(req.Metadata)
	if err != nil {
		return fmt.Errorf("could not marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		insert into moderation_audit_log (user_id, action, metadata, description)
		values ($1, $2, $3, $4)
	`, req.ModeratorUserID, req.Action, metadata, req.Description)
	if err != nil {
		return fmt.Errorf("could not insert audit log: %w", err)
	}

	return nil
}

var _ domain.ModerationAuditRepository = (*Repository)(nil)
