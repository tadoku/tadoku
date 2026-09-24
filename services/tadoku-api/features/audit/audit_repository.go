package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	queries "github.com/tadoku/tadoku/services/tadoku-api/generated/sqlc/audit"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/postgres"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, event Event) error {
	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}

	executor, err := postgres.Executor(ctx, r.db)
	if err != nil {
		return err
	}
	err = queries.New(executor).CreateAudit(ctx, queries.CreateAuditParams{
		ActorID:     postgres.UUID(event.ActorID),
		Action:      event.Action,
		Metadata:    metadata,
		Description: event.Description,
		RecordedAt:  postgres.Timestamp(event.recordedAt),
	})
	if err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}

	return nil
}
