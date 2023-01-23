package repository

import (
	"context"
	"fmt"

	"github.com/tadoku/tadoku/services/immersion-api/domain/command"
	"github.com/tadoku/tadoku/services/immersion-api/storage/postgres"
)

func (r *Repository) DeleteLog(ctx context.Context, req *command.DeleteLogRequest) error {
	isValid, err := r.q.CheckIfLogCanBeDeleted(ctx, postgres.CheckIfLogCanBeDeletedParams{
		Now:   req.Now,
		LogID: req.LogID,
	})
	if err != nil {
		return fmt.Errorf("could not check if log can be deleted: %w", err)
	}

	if !isValid {
		return command.ErrForbidden
	}

	if err := r.q.DeleteLog(ctx, req.LogID); err != nil {
		return fmt.Errorf("could not delete log: %w", err)
	}

	return nil
}
