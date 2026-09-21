package authz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/authz"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type AuthzRepository struct {
	db *pgxpool.Pool
}

func NewAuthzRepository(db *pgxpool.Pool) *AuthzRepository {
	return &AuthzRepository{db: db}
}

func (r *AuthzRepository) CreateModerationAudit(ctx context.Context, audit ModerationAudit) error {
	metadata, err := json.Marshal(map[string]string{
		"target_user_id": audit.TargetUserID.String(),
		"new_role":       string(audit.NewRole),
	})
	if err != nil {
		return fmt.Errorf("marshal moderation audit metadata: %w", err)
	}

	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	err = queries.New(executor).CreateModerationAudit(ctx, queries.CreateModerationAuditParams{
		ModeratorUserID: pgtype.UUID{Bytes: audit.ModeratorUserID, Valid: true},
		Action:          string(audit.Action),
		Metadata:        metadata,
		Description:     audit.Description,
		CreatedAt:       postgres.Timestamp(audit.CreatedAt),
	})
	if err != nil {
		return fmt.Errorf("insert moderation audit: %w", err)
	}

	return nil
}
