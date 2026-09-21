package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/features/logs"
)

func (a *Application) YearlyActivity(ctx context.Context, userID uuid.UUID, year int) (*logs.YearlyActivity, error) {
	return a.logs.YearlyActivity(ctx, userID, year)
}

func (a *Application) YearlyScores(ctx context.Context, userID uuid.UUID, year int) (*logs.YearlyScores, error) {
	return a.logs.YearlyScores(ctx, userID, year)
}

func (a *Application) YearlyActivitySplit(ctx context.Context, userID uuid.UUID, year int) ([]logs.ActivitySplitScore, error) {
	return a.logs.YearlyActivitySplit(ctx, userID, year)
}
