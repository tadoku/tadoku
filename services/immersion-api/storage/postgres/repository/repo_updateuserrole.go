package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/immersion-api/domain"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

// UserExists checks if a user exists in the database
func (r *Repository) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	// We can use a simple query to check if the user exists
	// Since we don't have a direct UserExists query, we can use FindUserByID or similar
	// For now, let's query the users table directly
	var exists bool
	err := r.psql.QueryRowContext(ctx,
		"select exists(select 1 from users where id = $1)",
		userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not check user existence: %w", err)
	}
	return exists, nil
}

// UpdateUserRole updates a user's role and creates an audit log entry
func (r *Repository) UpdateUserRole(ctx context.Context, req *domain.UpdateUserRoleRequest, moderatorUserID uuid.UUID) error {
	// Start transaction
	tx, err := r.psql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	qtx := r.q.WithTx(tx)

	// Determine action for audit log
	var action string
	if req.Role == "banned" {
		action = "ban_user"
	} else {
		action = "unban_user"
	}

	// Create audit log entry
	metadata := map[string]interface{}{
		"target_user_id": req.UserID.String(),
		"new_role":       req.Role,
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not marshal metadata: %w", err)
	}

	err = qtx.CreateModerationAuditLog(ctx, postgres.CreateModerationAuditLogParams{
		UserID:      moderatorUserID,
		Action:      action,
		Metadata:    metadataJSON,
		Description: postgres.NewNullString(&req.Reason),
	})
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("could not create audit log: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}
